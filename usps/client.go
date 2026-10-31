package usps

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/befship/usps-sdk/internal/generated/addresses"
	"github.com/befship/usps-sdk/internal/generated/labels"
	"github.com/befship/usps-sdk/internal/generated/prices"
	"github.com/befship/usps-sdk/internal/generated/tracking"
)

const (
	// EnvTEM is the USPS test environment (apis-tem.usps.com).
	EnvTEM = "tem"
	// EnvSandbox is an alias for TEM (Adapter / credentials may use "sandbox").
	EnvSandbox = "sandbox"
	// EnvProduction is production (apis.usps.com).
	EnvProduction = "production"
)

// Config holds USPS credentials and runtime options.
type Config struct {
	ClientID      string
	ClientSecret  string
	CRID          string
	MID           string
	ManifestMID   string
	AccountNumber string
	AccountType   string
	Environment   string // tem (default) | production | sandbox→tem
	BaseURL       string // optional override

	HTTPClient *http.Client
	Timeout    time.Duration
	MaxRetries int
	Logger     Logger
}

// Logger is a minimal structured logger hook.
type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

// Creds are USPS API credentials (used by adapters).
type Creds struct {
	ClientID      string
	ClientSecret  string
	CRID          string
	MID           string
	ManifestMID   string
	AccountNumber string
	AccountType   string
	Environment   string
	BaseURL       string
}

// Client talks to USPS APIs (OAuth + payment auth + typed API accessors).
type Client struct {
	cfg      Config
	HTTP     *http.Client
	TokCache *tokenCache
	PayCache *tokenCache
}

// New builds a Client.
func New(cfg Config) *Client {
	cfg.Environment = NormalizeEnv(cfg.Environment)
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	} else if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 2
	}
	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 60 * time.Second
		}
		hc = &http.Client{Timeout: timeout}
	}
	c := &Client{
		cfg:      cfg,
		HTTP:     hc,
		TokCache: newTokenCache(),
		PayCache: newTokenCache(),
	}
	c.ensureRetry()
	return c
}

// Environment returns the configured environment name.
func (c *Client) Environment() string { return c.cfg.Environment }

// HTTPClient returns the underlying HTTP client.
func (c *Client) HTTPClient() *http.Client { return c.HTTP }

// NormalizeEnv lowercases and maps sandbox→tem; empty defaults to tem.
func NormalizeEnv(env string) string {
	env = strings.ToLower(strings.TrimSpace(env))
	switch env {
	case "", EnvSandbox, "test", "cie":
		return EnvTEM
	case "prod":
		return EnvProduction
	default:
		return env
	}
}

// BaseURL returns the USPS API host for env (tem|production) or override.
func BaseURL(env string, override ...string) string {
	if len(override) > 0 && strings.TrimSpace(override[0]) != "" {
		return strings.TrimRight(strings.TrimSpace(override[0]), "/")
	}
	switch NormalizeEnv(env) {
	case EnvProduction:
		return "https://apis.usps.com"
	default:
		return "https://apis-tem.usps.com"
	}
}

// BaseURLFromCreds resolves the API base from a Creds bundle.
func BaseURLFromCreds(cr Creds) string {
	return BaseURL(cr.Environment, cr.BaseURL)
}

func (c *Client) resolveCreds(cr Creds) Creds {
	if cr.ClientID == "" {
		cr.ClientID = c.cfg.ClientID
	}
	if cr.ClientSecret == "" {
		cr.ClientSecret = c.cfg.ClientSecret
	}
	if cr.CRID == "" {
		cr.CRID = c.cfg.CRID
	}
	if cr.MID == "" {
		cr.MID = c.cfg.MID
	}
	if cr.ManifestMID == "" {
		cr.ManifestMID = c.cfg.ManifestMID
		if cr.ManifestMID == "" {
			cr.ManifestMID = cr.MID
		}
	}
	if cr.AccountNumber == "" {
		cr.AccountNumber = c.cfg.AccountNumber
	}
	if cr.AccountType == "" {
		cr.AccountType = c.cfg.AccountType
		if cr.AccountType == "" {
			cr.AccountType = "EPS"
		}
	}
	if cr.Environment == "" {
		cr.Environment = c.cfg.Environment
	}
	if cr.BaseURL == "" {
		cr.BaseURL = c.cfg.BaseURL
	}
	cr.Environment = NormalizeEnv(cr.Environment)
	return cr
}

// AccessToken fetches (or returns cached) OAuth client-credentials token.
func (c *Client) AccessToken(ctx context.Context, cr Creds) (string, error) {
	cr = c.resolveCreds(cr)
	key := cr.Environment + ":" + cr.ClientID
	if tok, ok := c.TokCache.get(key); ok {
		return tok, nil
	}
	payload := map[string]any{
		"client_id":     cr.ClientID,
		"client_secret": cr.ClientSecret,
		"grant_type":    "client_credentials",
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   any    `json:"expires_in"`
	}
	if err := httpJSON(ctx, c.HTTP, http.MethodPost, BaseURLFromCreds(cr)+"/oauth2/v3/token", map[string]string{
		"Content-Type": "application/json",
	}, payload, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("usps oauth: empty access_token")
	}
	sec := int(asFloat(out.ExpiresIn))
	if sec <= 0 {
		sec = 28800
	}
	c.TokCache.set(key, out.AccessToken, sec)
	return out.AccessToken, nil
}

// PaymentToken fetches (or returns cached) payment authorization token.
func (c *Client) PaymentToken(ctx context.Context, cr Creds) (string, error) {
	cr = c.resolveCreds(cr)
	if cr.CRID == "" || cr.MID == "" || cr.AccountNumber == "" {
		return "", fmt.Errorf("usps labels require crid, mid, account_number")
	}
	key := cr.Environment + ":" + cr.AccountNumber
	if tok, ok := c.PayCache.get(key); ok {
		return tok, nil
	}
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"roles": []map[string]any{{
			"roleName":      "PAYER",
			"CRID":          cr.CRID,
			"MID":           cr.MID,
			"manifestMID":   cr.ManifestMID,
			"accountType":   cr.AccountType,
			"accountNumber": cr.AccountNumber,
		}},
	}
	var out struct {
		PaymentAuthorizationToken string `json:"paymentAuthorizationToken"`
		ExpiresIn                 any    `json:"expiresIn"`
	}
	if err := httpJSON(ctx, c.HTTP, http.MethodPost, BaseURLFromCreds(cr)+"/payments/v3/payment-authorization", map[string]string{
		"Authorization": "Bearer " + tok,
		"Content-Type":  "application/json",
	}, payload, &out); err != nil {
		return "", err
	}
	if out.PaymentAuthorizationToken == "" {
		return "", fmt.Errorf("usps payments: empty paymentAuthorizationToken")
	}
	sec := int(asFloat(out.ExpiresIn))
	if sec <= 0 {
		sec = 3600
	}
	c.PayCache.set(key, out.PaymentAuthorizationToken, sec)
	return out.PaymentAuthorizationToken, nil
}

// AuthHeaders returns Bearer + Content-Type for ordinary calls.
func (c *Client) AuthHeaders(ctx context.Context, cr Creds) (map[string]string, error) {
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"Authorization": "Bearer " + tok,
		"Content-Type":  "application/json",
	}, nil
}

// PaymentHeaders returns Bearer + payment authorization headers for label ops.
func (c *Client) PaymentHeaders(ctx context.Context, cr Creds) (map[string]string, error) {
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	pay, err := c.PaymentToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"Authorization":                 "Bearer " + tok,
		"X-Payment-Authorization-Token": pay,
		"Content-Type":                  "application/json",
	}, nil
}

func bearerEditor(tok string) func(context.Context, *http.Request) error {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+tok)
		if req.Header.Get("Accept") == "" {
			req.Header.Set("Accept", "application/json")
		}
		return nil
	}
}

// apiBase returns a path-qualified API base (e.g. .../labels/v3).
// Creds.BaseURL, when set, overrides the entire base URL (must include API path).
func (c *Client) apiBase(cr Creds, pkg string) string {
	if strings.TrimSpace(cr.BaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(cr.BaseURL), "/")
	}
	return ServerURL(cr.Environment, pkg)
}

// Addresses returns a typed Addresses client with Bearer auth.
func (c *Client) Addresses(ctx context.Context, cr Creds) (*AddressesClient, error) {
	cr = c.resolveCreds(cr)
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	return addresses.NewClientWithResponses(
		c.apiBase(cr, "addresses"),
		addresses.WithHTTPClient(c.HTTP),
		addresses.WithRequestEditorFn(bearerEditor(tok)),
	)
}

// Prices returns a typed Prices client with Bearer auth.
func (c *Client) Prices(ctx context.Context, cr Creds) (*PricesClient, error) {
	cr = c.resolveCreds(cr)
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	return prices.NewClientWithResponses(
		c.apiBase(cr, "prices"),
		prices.WithHTTPClient(c.HTTP),
		prices.WithRequestEditorFn(bearerEditor(tok)),
	)
}

// Tracking returns a typed Tracking client with Bearer auth.
func (c *Client) Tracking(ctx context.Context, cr Creds) (*TrackingClient, error) {
	cr = c.resolveCreds(cr)
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	return tracking.NewClientWithResponses(
		c.apiBase(cr, "tracking"),
		tracking.WithHTTPClient(c.HTTP),
		tracking.WithRequestEditorFn(bearerEditor(tok)),
	)
}

// Labels returns a typed Labels client with Bearer auth.
// Callers must still pass CreateLabelParams / GetLabelParams / CancelLabelParams
// with X-Payment-Authorization-Token (see PaymentToken).
func (c *Client) Labels(ctx context.Context, cr Creds) (*LabelsClient, error) {
	cr = c.resolveCreds(cr)
	tok, err := c.AccessToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	return labels.NewClientWithResponses(
		c.apiBase(cr, "labels"),
		labels.WithHTTPClient(c.HTTP),
		labels.WithRequestEditorFn(bearerEditor(tok)),
	)
}

// LabelsWithPayment returns Labels client plus payment-auth params for label ops.
func (c *Client) LabelsWithPayment(ctx context.Context, cr Creds) (*LabelsClient, PaymentAuth, error) {
	cr = c.resolveCreds(cr)
	api, err := c.Labels(ctx, cr)
	if err != nil {
		return nil, "", err
	}
	pay, err := c.PaymentToken(ctx, cr)
	if err != nil {
		return nil, "", err
	}
	return api, PaymentAuth(pay), nil
}
