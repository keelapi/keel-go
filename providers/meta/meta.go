// Package meta provides a drop-in Meta Llama-compatible client that routes through Keel.
// It uses the OpenAI-compatible interface but routes to /v1/proxy/meta.
package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	keel "github.com/keelapi/keel-go"
)

// Config configures the Meta-compatible Keel client.
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

// Client is a Meta-compatible client that routes through Keel.
type Client struct {
	Chat ChatNamespace
	cfg  Config
	keel *keel.Client
}

// ChatNamespace groups chat-related resources.
type ChatNamespace struct {
	Completions *CompletionsResource
}

// CompletionsResource provides chat completion methods.
type CompletionsResource struct {
	cfg  Config
	keel *keel.Client
}

// ChatCompletionMessage represents a message in a chat completion request.
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionParams are the parameters for a chat completion request.
type ChatCompletionParams struct {
	Model              string                  `json:"model"`
	Messages           []ChatCompletionMessage `json:"messages"`
	MaxTokens          *int                    `json:"max_tokens,omitempty"`
	Temperature        *float64                `json:"temperature,omitempty"`
	TopP               *float64                `json:"top_p,omitempty"`
	Stream             bool                    `json:"stream,omitempty"`
	Extra              map[string]any          `json:"extra,omitempty"`
	KeelParentPermitID *string                 `json:"-"`
	KeelSessionID      *string                 `json:"-"`
}

// ChatCompletionChoice represents a choice in a chat completion response.
type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

// ChatCompletionUsage represents token usage.
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse is the response from a chat completion request.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   ChatCompletionUsage    `json:"usage"`
}

// ChatCompletionChunk represents a streamed chunk.
type ChatCompletionChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                   `json:"index"`
		Delta        ChatCompletionMessage `json:"delta"`
		FinishReason *string               `json:"finish_reason"`
	} `json:"choices"`
}

// NewClient creates a new Meta-compatible Keel client.
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

	comp := &CompletionsResource{cfg: cfg, keel: k}
	return &Client{
		Chat: ChatNamespace{Completions: comp},
		cfg:  cfg,
		keel: k,
	}
}

func defaultSubject() keel.PermitSubject {
	return keel.PermitSubject{Type: "service", ID: "default"}
}

// Create performs a chat completion through Keel governance via Meta.
func (c *CompletionsResource) Create(ctx context.Context, params ChatCompletionParams) (*ChatCompletionResponse, error) {
	subject := defaultSubject()
	if c.cfg.KeelSubject != nil {
		subject = *c.cfg.KeelSubject
	}

	payload := map[string]any{
		"model":    params.Model,
		"messages": params.Messages,
		"stream":   false,
	}
	for key, value := range params.Extra {
		payload[key] = value
	}
	if params.MaxTokens != nil {
		payload["max_tokens"] = *params.MaxTokens
	}
	if params.Temperature != nil {
		payload["temperature"] = *params.Temperature
	}
	if params.TopP != nil {
		payload["top_p"] = *params.TopP
	}

	idempotencyKey := fmt.Sprintf("meta-%s-%d", params.Model, time.Now().UnixNano())
	managedPayload := keel.BuildManagedProxyPayload(payload, c.cfg.KeelProjectID, &subject, params.KeelParentPermitID, params.KeelSessionID)
	raw, err := c.keel.Proxy.MetaWithHeaders(ctx, managedPayload, keel.ManagedProxyHeaders(idempotencyKey))
	if err != nil {
		return nil, fmt.Errorf("meta: proxy request failed: %w", err)
	}

	data, _ := json.Marshal(raw)
	var resp ChatCompletionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("meta: decode response: %w", err)
	}

	return &resp, nil
}

// CreateStream performs a streaming chat completion through Keel governance via Meta.
func (c *CompletionsResource) CreateStream(ctx context.Context, params ChatCompletionParams) (<-chan ChatCompletionChunk, <-chan error) {
	chunks := make(chan ChatCompletionChunk)
	errc := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errc)

		subject := defaultSubject()
		if c.cfg.KeelSubject != nil {
			subject = *c.cfg.KeelSubject
		}

		payload := map[string]any{
			"model":    params.Model,
			"messages": params.Messages,
			"stream":   true,
		}
		for key, value := range params.Extra {
			payload[key] = value
		}
		if params.MaxTokens != nil {
			payload["max_tokens"] = *params.MaxTokens
		}
		if params.Temperature != nil {
			payload["temperature"] = *params.Temperature
		}
		if params.TopP != nil {
			payload["top_p"] = *params.TopP
		}

		idempotencyKey := fmt.Sprintf("meta-stream-%s-%d", params.Model, time.Now().UnixNano())
		managedPayload := keel.BuildManagedProxyPayload(payload, c.cfg.KeelProjectID, &subject, params.KeelParentPermitID, params.KeelSessionID)
		events, sseErrc := c.keel.Proxy.MetaStreamWithHeaders(ctx, managedPayload, keel.ManagedProxyHeaders(idempotencyKey))
		for sse := range events {
			var chunk ChatCompletionChunk
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
