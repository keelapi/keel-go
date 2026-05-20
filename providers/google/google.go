// Package google provides a drop-in Google Gemini-compatible client that routes through Keel.
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	keel "github.com/keelapi/keel-go"
)

// Config configures the Google-compatible Keel client.
type Config struct {
	APIKey        string
	KeelBaseURL   string
	KeelAPIKey    string
	KeelProjectID string
	KeelSubject   *keel.PermitSubject
	Timeout       time.Duration
	MaxRetries    int
}

func (c *Config) resolve() {
	if c.KeelBaseURL == "" {
		c.KeelBaseURL = os.Getenv("KEEL_BASE_URL")
	}
	if c.KeelAPIKey == "" {
		c.KeelAPIKey = os.Getenv("KEEL_API_KEY")
	}
	if c.KeelProjectID == "" {
		c.KeelProjectID = os.Getenv("KEEL_PROJECT_ID")
	}
}

// Client is a Google Gemini-compatible client that routes through Keel.
type Client struct {
	cfg  Config
	keel *keel.Client
}

// NewClient creates a new Google-compatible Keel client.
func NewClient(cfg Config) *Client {
	cfg.resolve()

	rc := &keel.RetryConfig{
		MaxRetries:        cfg.MaxRetries,
		InitialDelay:      500 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableStatusCodes: map[int]bool{
			408: true, 429: true, 500: true, 502: true, 503: true, 504: true,
		},
	}
	if cfg.MaxRetries == 0 {
		rc.MaxRetries = 3
	}

	k := keel.NewClient(keel.ClientConfig{
		BaseURL:     cfg.KeelBaseURL,
		APIKey:      cfg.KeelAPIKey,
		Timeout:     cfg.Timeout,
		RetryConfig: rc,
	})

	return &Client{cfg: cfg, keel: k}
}

// GenerativeModel returns a handle to a specific model.
func (c *Client) GenerativeModel(model string) *GenerativeModel {
	return &GenerativeModel{model: model, cfg: c.cfg, keel: c.keel}
}

// GenerativeModel represents a specific generative model.
type GenerativeModel struct {
	model string
	cfg   Config
	keel  *keel.Client
}

// Part represents a content part.
type Part struct {
	Text string `json:"text,omitempty"`
}

// Content represents a content block.
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

// GenerateContentRequest is the request for content generation.
type GenerateContentRequest struct {
	Contents           []Content      `json:"contents"`
	Extra              map[string]any `json:"extra,omitempty"`
	KeelParentPermitID *string        `json:"-"`
	KeelSessionID      *string        `json:"-"`
}

// Candidate represents a response candidate.
type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// UsageMetadata represents token usage.
type UsageMetadata struct {
	PromptTokenCount     int `json:"prompt_token_count"`
	CandidatesTokenCount int `json:"candidates_token_count"`
	TotalTokenCount      int `json:"total_token_count"`
}

// GenerateContentResponse is the response from content generation.
type GenerateContentResponse struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usage_metadata"`
}

// GenerateContentChunk represents a streamed chunk.
type GenerateContentChunk struct {
	Candidates    []Candidate    `json:"candidates"`
	UsageMetadata *UsageMetadata `json:"usage_metadata,omitempty"`
}

func defaultSubject() keel.PermitSubject {
	return keel.PermitSubject{Type: "service", ID: "default"}
}

// GenerateContent generates content through Keel governance.
func (m *GenerativeModel) GenerateContent(ctx context.Context, req GenerateContentRequest) (*GenerateContentResponse, error) {
	subject := defaultSubject()
	if m.cfg.KeelSubject != nil {
		subject = *m.cfg.KeelSubject
	}

	payload := map[string]any{
		"model":    m.model,
		"contents": req.Contents,
		"stream":   false,
	}
	for key, value := range req.Extra {
		payload[key] = value
	}

	idempotencyKey := fmt.Sprintf("google-%s-%d", m.model, time.Now().UnixNano())
	managedPayload := keel.BuildManagedProxyPayload(payload, m.cfg.KeelProjectID, &subject, req.KeelParentPermitID, req.KeelSessionID)
	raw, err := m.keel.Proxy.GoogleWithHeaders(ctx, managedPayload, keel.ManagedProxyHeaders(idempotencyKey))
	if err != nil {
		return nil, fmt.Errorf("google: proxy request failed: %w", err)
	}

	data, _ := json.Marshal(raw)
	var resp GenerateContentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("google: decode response: %w", err)
	}

	return &resp, nil
}

// GenerateContentStream generates content with streaming through Keel governance.
func (m *GenerativeModel) GenerateContentStream(ctx context.Context, req GenerateContentRequest) (<-chan GenerateContentChunk, <-chan error) {
	chunks := make(chan GenerateContentChunk)
	errc := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errc)

		subject := defaultSubject()
		if m.cfg.KeelSubject != nil {
			subject = *m.cfg.KeelSubject
		}

		payload := map[string]any{
			"model":    m.model,
			"contents": req.Contents,
			"stream":   true,
		}
		for key, value := range req.Extra {
			payload[key] = value
		}

		idempotencyKey := fmt.Sprintf("google-stream-%s-%d", m.model, time.Now().UnixNano())
		managedPayload := keel.BuildManagedProxyPayload(payload, m.cfg.KeelProjectID, &subject, req.KeelParentPermitID, req.KeelSessionID)
		events, sseErrc := m.keel.Proxy.GoogleStreamWithHeaders(ctx, managedPayload, keel.ManagedProxyHeaders(idempotencyKey))
		for sse := range events {
			var chunk GenerateContentChunk
			if json.Unmarshal(sse.Data, &chunk) == nil {
				chunks <- chunk
			}
		}
		if err := <-sseErrc; err != nil {
			errc <- err
		}
	}()

	return chunks, errc
}
