// Package openai provides a drop-in OpenAI-compatible client that routes through Keel.
package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	keel "github.com/keelapi/keel-go"
)

// Config configures the OpenAI-compatible Keel client.
type Config struct {
	// APIKey is ignored — Keel manages provider keys.
	APIKey string

	// KeelBaseURL is the Keel API base URL. Falls back to KEEL_BASE_URL env var.
	KeelBaseURL string

	// KeelAPIKey is the Keel Bearer token. Falls back to KEEL_API_KEY env var.
	KeelAPIKey string

	// KeelProjectID is the Keel project ID. Falls back to KEEL_PROJECT_ID env var.
	KeelProjectID string

	// KeelSubject identifies the caller for permit creation.
	KeelSubject *keel.PermitSubject

	// Timeout for HTTP requests. Defaults to 30s.
	Timeout time.Duration

	// MaxRetries for failed requests. Defaults to 3.
	MaxRetries int
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

// Client is an OpenAI-compatible client that routes through Keel.
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

// NewClient creates a new OpenAI-compatible Keel client.
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
	Tools              []any                   `json:"tools,omitempty"`
	ToolChoice         any                     `json:"tool_choice,omitempty"`
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

// ChatCompletionUsage represents token usage in a chat completion response.
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

// ChatCompletionChunk represents a streamed chunk of a chat completion.
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
	Usage *ChatCompletionUsage `json:"usage,omitempty"`
}

func defaultSubject() keel.PermitSubject {
	return keel.PermitSubject{Type: "service", ID: "default"}
}

// Create performs a chat completion through Keel governance.
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
	if len(params.Tools) > 0 {
		payload["tools"] = params.Tools
	}
	if params.ToolChoice != nil {
		payload["tool_choice"] = params.ToolChoice
	}

	idempotencyKey := fmt.Sprintf("openai-%s-%d", params.Model, time.Now().UnixNano())
	managedPayload := keel.BuildManagedProxyPayload(payload, c.cfg.KeelProjectID, &subject, params.KeelParentPermitID, params.KeelSessionID)
	raw, err := c.keel.Proxy.OpenAIWithHeaders(ctx, managedPayload, keel.ManagedProxyHeaders(idempotencyKey))
	if err != nil {
		return nil, fmt.Errorf("openai: proxy request failed: %w", err)
	}

	data, _ := json.Marshal(raw)
	var resp ChatCompletionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}

	return &resp, nil
}

// CreateStream performs a streaming chat completion through Keel governance.
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
		if len(params.Tools) > 0 {
			payload["tools"] = params.Tools
		}
		if params.ToolChoice != nil {
			payload["tool_choice"] = params.ToolChoice
		}

		idempotencyKey := fmt.Sprintf("openai-stream-%s-%d", params.Model, time.Now().UnixNano())
		managedPayload := keel.BuildManagedProxyPayload(payload, c.cfg.KeelProjectID, &subject, params.KeelParentPermitID, params.KeelSessionID)
		events, sseErrc := c.keel.Proxy.OpenAIStreamWithHeaders(ctx, managedPayload, keel.ManagedProxyHeaders(idempotencyKey))
		for sse := range events {
			var chunk ChatCompletionChunk
			if json.Unmarshal(sse.Data, &chunk) == nil {
				chunks <- chunk
			}
		}
		if err := <-sseErrc; err != nil {
			errc <- err
			return
		}
	}()

	return chunks, errc
}
