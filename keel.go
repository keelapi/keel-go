// Package keel provides the official Go SDK for the Keel AI governance and execution control plane.
//
// Keel sits between applications and AI providers, enforcing permits, firewalls,
// routing, accounting, and audit on every AI request.
package keel

import (
	"net/http"
	"os"
	"time"
)

// Client is the top-level Keel API client.
type Client struct {
	config    ClientConfig
	transport *httpTransport

	// Permits manages permit lifecycle — create, dry-run, list, attest, evidence, lineage.
	Permits *PermitsClient

	// Executions manages governed AI executions — sync and streaming.
	Executions *ExecutionsClient

	// Execute provides the combined permit+execute shorthand.
	Execute *ExecuteClient

	// Proxy provides pass-through proxy to AI providers.
	Proxy *ProxyClient

	// Jobs manages batch jobs.
	Jobs *JobsClient

	// ApiKeys manages API key lifecycle.
	ApiKeys *ApiKeysClient

	// Requests provides request timeline inspection.
	Requests *RequestsClient

	// Workflows manages caller-declared workflow intent.
	Workflows *WorkflowsClient
}

// NewClient creates a new Keel client with the given configuration.
// ConfigFromEnv returns a ClientConfig populated from environment variables
// KEEL_BASE_URL and KEEL_API_KEY. Fields already set in the caller's config
// take precedence.
func ConfigFromEnv() ClientConfig {
	return ClientConfig{
		BaseURL: os.Getenv("KEEL_BASE_URL"),
		APIKey:  os.Getenv("KEEL_API_KEY"),
	}
}

// NewClient creates a new Keel client with the given configuration.
// If BaseURL or APIKey are empty, the corresponding KEEL_BASE_URL and
// KEEL_API_KEY environment variables are used as fallbacks.
func NewClient(config ClientConfig) *Client {
	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("KEEL_BASE_URL")
	}
	if config.APIKey == "" {
		config.APIKey = os.Getenv("KEEL_API_KEY")
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	rc := config.RetryConfig
	if rc == nil {
		rc = defaultRetryConfig()
	}

	hc := config.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: config.Timeout}
	}

	t := &httpTransport{
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		freshness:  config.RequestFreshness,
		httpClient: hc,
		retryCfg:   rc,
	}

	c := &Client{
		config:    config,
		transport: t,
	}
	c.Permits = &PermitsClient{t: t}
	c.Executions = &ExecutionsClient{t: t}
	c.Execute = &ExecuteClient{t: t}
	c.Proxy = &ProxyClient{t: t}
	c.Jobs = &JobsClient{t: t}
	c.ApiKeys = &ApiKeysClient{t: t}
	c.Requests = &RequestsClient{t: t}
	c.Workflows = &WorkflowsClient{t: t}

	return c
}
