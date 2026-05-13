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
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	MaxTokens   *int                    `json:"max_tokens,omitempty"`
	Temperature *float64                `json:"temperature,omitempty"`
	TopP        *float64                `json:"top_p,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
	Extra       map[string]any          `json:"extra,omitempty"`
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
}

func estimateTokens(messages []ChatCompletionMessage) int {
	total := 0
	for _, m := range messages {
		total += len(m.Content) / 4
	}
	if total == 0 {
		total = 1
	}
	return total
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

	model := params.Model
	estimatedTokens := estimateTokens(params.Messages)
	permit, err := c.keel.Permits.Create(ctx, keel.PermitRequest{
		ProjectID:      c.cfg.KeelProjectID,
		IdempotencyKey: fmt.Sprintf("openai-%s-%d", model, time.Now().UnixNano()),
		Subject:        subject,
		Action:         keel.Action{Name: string(keel.OpGenerateText)},
		Resource: keel.Resource{
			Type: "ai_model",
			ID:   model,
			Attributes: keel.ResourceAttributes{
				Provider:              string(keel.ProviderOpenAI),
				Model:                 model,
				Operation:             keel.OpGenerateText,
				EstimatedInputTokens:  estimatedTokens,
				EstimatedOutputTokens: estimatedTokens,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai: permit creation failed: %w", err)
	}

	if permit.Decision != keel.DecisionAllow {
		return nil, fmt.Errorf("openai: permit denied: %s", permit.Decision)
	}

	payload := map[string]any{
		"model":    params.Model,
		"messages": params.Messages,
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

	raw, err := c.keel.Proxy.OpenAI(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("openai: proxy request failed: %w", err)
	}

	data, _ := json.Marshal(raw)
	var resp ChatCompletionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}

	_, _ = c.keel.Permits.ReportUsage(ctx, permit.PermitID, keel.PermitUsageReportRequest{
		ActualInputTokens:  &resp.Usage.PromptTokens,
		ActualOutputTokens: &resp.Usage.CompletionTokens,
		ActualTotalTokens:  &resp.Usage.TotalTokens,
	})

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

		model := params.Model
		estimatedTokens := estimateTokens(params.Messages)
		permit, err := c.keel.Permits.Create(ctx, keel.PermitRequest{
			ProjectID:      c.cfg.KeelProjectID,
			IdempotencyKey: fmt.Sprintf("openai-stream-%s-%d", model, time.Now().UnixNano()),
			Subject:        subject,
			Action:         keel.Action{Name: string(keel.OpGenerateText)},
			Resource: keel.Resource{
				Type: "ai_model",
				ID:   model,
				Attributes: keel.ResourceAttributes{
					Provider:              string(keel.ProviderOpenAI),
					Model:                 model,
					Operation:             keel.OpGenerateText,
					ExecutionMode:         keel.ModeStream,
					EstimatedInputTokens:  estimatedTokens,
					EstimatedOutputTokens: estimatedTokens,
				},
			},
		})
		if err != nil {
			errc <- fmt.Errorf("openai: permit creation failed: %w", err)
			return
		}

		if permit.Decision != keel.DecisionAllow {
			errc <- fmt.Errorf("openai: permit denied: %s", permit.Decision)
			return
		}

		payload := map[string]any{
			"model":    params.Model,
			"messages": params.Messages,
			"stream":   true,
		}
		if params.MaxTokens != nil {
			payload["max_tokens"] = *params.MaxTokens
		}
		if params.Temperature != nil {
			payload["temperature"] = *params.Temperature
		}

		events, sseErrc := c.keel.Proxy.OpenAIStream(ctx, payload)
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
