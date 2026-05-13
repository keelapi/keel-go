// Package anthropic provides a drop-in Anthropic-compatible client that routes through Keel.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	keel "github.com/keelapi/keel-go"
)

// Config configures the Anthropic-compatible Keel client.
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

// Client is an Anthropic-compatible client that routes through Keel.
type Client struct {
	Messages *MessagesResource
	cfg      Config
	keel     *keel.Client
}

// MessagesResource provides message creation methods.
type MessagesResource struct {
	cfg  Config
	keel *keel.Client
}

// NewClient creates a new Anthropic-compatible Keel client.
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

	msgs := &MessagesResource{cfg: cfg, keel: k}
	return &Client{
		Messages: msgs,
		cfg:      cfg,
		keel:     k,
	}
}

// MessageCreateParams are the parameters for creating a message.
type MessageCreateParams struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	Messages  []MessageParam `json:"messages"`
	System    *string        `json:"system,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"`
}

// MessageParam represents a message in the conversation.
type MessageParam struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ContentBlock represents a content block in a message response.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// MessageUsage represents token usage in a message response.
type MessageUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// MessageResponse is the response from creating a message.
type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   *string        `json:"stop_reason,omitempty"`
	StopSequence *string        `json:"stop_sequence,omitempty"`
	Usage        MessageUsage   `json:"usage"`
}

// MessageStreamEvent represents a single event in a streaming message response.
type MessageStreamEvent struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data,omitempty"`
	Index *int            `json:"index,omitempty"`
	Delta *ContentBlock   `json:"delta,omitempty"`
}

func estimateTokens(messages []MessageParam) int {
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

// Create performs a message creation through Keel governance.
func (m *MessagesResource) Create(ctx context.Context, params MessageCreateParams) (*MessageResponse, error) {
	subject := defaultSubject()
	if m.cfg.KeelSubject != nil {
		subject = *m.cfg.KeelSubject
	}

	model := params.Model
	tokens := estimateTokens(params.Messages)
	permit, err := m.keel.Permits.Create(ctx, keel.PermitRequest{
		ProjectID:      m.cfg.KeelProjectID,
		IdempotencyKey: fmt.Sprintf("anthropic-%s-%d", model, time.Now().UnixNano()),
		Subject:        subject,
		Action:         keel.Action{Name: string(keel.OpGenerateText)},
		Resource: keel.Resource{
			Type: "ai_model",
			ID:   model,
			Attributes: keel.ResourceAttributes{
				Provider:              string(keel.ProviderAnthropic),
				Model:                 model,
				Operation:             keel.OpGenerateText,
				EstimatedInputTokens:  tokens,
				EstimatedOutputTokens: tokens,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic: permit creation failed: %w", err)
	}

	if permit.Decision != keel.DecisionAllow {
		return nil, fmt.Errorf("anthropic: permit denied: %s", permit.Decision)
	}

	payload := map[string]any{
		"model":      params.Model,
		"max_tokens": params.MaxTokens,
		"messages":   params.Messages,
	}
	if params.System != nil {
		payload["system"] = *params.System
	}

	raw, err := m.keel.Proxy.Anthropic(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("anthropic: proxy request failed: %w", err)
	}

	data, _ := json.Marshal(raw)
	var resp MessageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("anthropic: decode response: %w", err)
	}

	inputTokens := resp.Usage.InputTokens
	outputTokens := resp.Usage.OutputTokens
	totalTokens := inputTokens + outputTokens
	_, _ = m.keel.Permits.ReportUsage(ctx, permit.PermitID, keel.PermitUsageReportRequest{
		ActualInputTokens:  &inputTokens,
		ActualOutputTokens: &outputTokens,
		ActualTotalTokens:  &totalTokens,
	})

	return &resp, nil
}

// CreateStream performs a streaming message creation through Keel governance.
func (m *MessagesResource) CreateStream(ctx context.Context, params MessageCreateParams) (<-chan MessageStreamEvent, <-chan error) {
	events := make(chan MessageStreamEvent)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)

		subject := defaultSubject()
		if m.cfg.KeelSubject != nil {
			subject = *m.cfg.KeelSubject
		}

		model := params.Model
		tokens := estimateTokens(params.Messages)
		permit, err := m.keel.Permits.Create(ctx, keel.PermitRequest{
			ProjectID:      m.cfg.KeelProjectID,
			IdempotencyKey: fmt.Sprintf("anthropic-stream-%s-%d", model, time.Now().UnixNano()),
			Subject:        subject,
			Action:         keel.Action{Name: string(keel.OpGenerateText)},
			Resource: keel.Resource{
				Type: "ai_model",
				ID:   model,
				Attributes: keel.ResourceAttributes{
					Provider:              string(keel.ProviderAnthropic),
					Model:                 model,
					Operation:             keel.OpGenerateText,
					ExecutionMode:         keel.ModeStream,
					EstimatedInputTokens:  tokens,
					EstimatedOutputTokens: tokens,
				},
			},
		})
		if err != nil {
			errc <- fmt.Errorf("anthropic: permit creation failed: %w", err)
			return
		}

		if permit.Decision != keel.DecisionAllow {
			errc <- fmt.Errorf("anthropic: permit denied: %s", permit.Decision)
			return
		}

		payload := map[string]any{
			"model":      params.Model,
			"max_tokens": params.MaxTokens,
			"messages":   params.Messages,
			"stream":     true,
		}
		if params.System != nil {
			payload["system"] = *params.System
		}

		sseEvents, sseErrc := m.keel.Proxy.AnthropicStream(ctx, payload)
		for sse := range sseEvents {
			var evt MessageStreamEvent
			if json.Unmarshal(sse.Data, &evt) == nil {
				events <- evt
			}
		}
		if err := <-sseErrc; err != nil {
			errc <- err
		}
	}()

	return events, errc
}
