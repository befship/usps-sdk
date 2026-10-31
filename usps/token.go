package usps

import (
	"sync"
	"time"
)

type tokenCache struct {
	mu   sync.Mutex
	tok  map[string]cachedToken
	skew time.Duration
}

type cachedToken struct {
	AccessToken string
	Expiry      time.Time
}

func newTokenCache() *tokenCache {
	return &tokenCache{tok: map[string]cachedToken{}, skew: 30 * time.Second}
}

func (c *tokenCache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.tok[key]
	if !ok || t.AccessToken == "" {
		return "", false
	}
	if time.Until(t.Expiry) <= c.skew {
		return "", false
	}
	return t.AccessToken, true
}

func (c *tokenCache) set(key, accessToken string, expiresInSec int) {
	if expiresInSec <= 0 {
		expiresInSec = 3600
	}
	c.mu.Lock()
	c.tok[key] = cachedToken{
		AccessToken: accessToken,
		Expiry:      time.Now().Add(time.Duration(expiresInSec) * time.Second),
	}
	c.mu.Unlock()
}
