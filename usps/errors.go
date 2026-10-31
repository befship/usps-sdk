package usps

import (
	"fmt"
	"strings"
)

// APIError is a normalized USPS / transport error.
type APIError struct {
	Op         string
	StatusCode int
	Message    string
	Body       string
	Err        error
}

func (e *APIError) Error() string {
	var b strings.Builder
	b.WriteString("usps")
	if e.Op != "" {
		b.WriteString(": ")
		b.WriteString(e.Op)
	}
	if e.StatusCode != 0 {
		fmt.Fprintf(&b, " HTTP %d", e.StatusCode)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	if e.Err != nil {
		fmt.Fprintf(&b, ": %v", e.Err)
	}
	if e.Body != "" && e.Message == "" {
		fmt.Fprintf(&b, ": %s", truncate(e.Body, 512))
	}
	return b.String()
}

func (e *APIError) Unwrap() error { return e.Err }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
