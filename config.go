package keel

import (
	"net/http"
	"time"
)

// ClientConfig configures the Keel client.
type ClientConfig struct {
	// BaseURL is the Keel API base URL (e.g. "https://api.keelapi.com").
	BaseURL string

	// APIKey is the Bearer token for authentication.
	APIKey string

	// Timeout for HTTP requests. Defaults to 30s.
	Timeout time.Duration

	// RequestFreshness enables X-Keel-Timestamp and X-Keel-Nonce headers.
	RequestFreshness bool

	// RetryConfig configures retry behavior. Nil uses defaults.
	RetryConfig *RetryConfig

	// HTTPClient overrides the default HTTP client. Nil uses a default client.
	HTTPClient *http.Client
}

// RequestOptions provides per-request overrides.
type RequestOptions struct {
	Headers map[string]string
}
