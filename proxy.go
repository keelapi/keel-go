package keel

import (
	"context"
	"encoding/json"
	"fmt"
)

// ProxyClient provides pass-through proxy access to AI providers.
type ProxyClient struct {
	t *httpTransport
}

func (c *ProxyClient) proxySync(ctx context.Context, provider string, payload any) (map[string]any, error) {
	body, err := c.t.post(ctx, "/v1/proxy/"+provider, payload, nil)
	if err != nil {
		return nil, err
	}
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return resp, nil
}

func (c *ProxyClient) proxyStream(ctx context.Context, provider string, payload any) (<-chan SSEEvent, <-chan error) {
	events := make(chan SSEEvent)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)

		rc, err := c.t.postStream(ctx, "/v1/proxy/"+provider, payload, nil)
		if err != nil {
			errc <- err
			return
		}

		sseEvents, sseErrc := parseSSEStream(rc)
		for sse := range sseEvents {
			events <- sse
		}
		if err := <-sseErrc; err != nil {
			errc <- err
		}
	}()

	return events, errc
}

// OpenAI proxies a request to OpenAI.
func (c *ProxyClient) OpenAI(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "openai", payload)
}

// Anthropic proxies a request to Anthropic.
func (c *ProxyClient) Anthropic(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "anthropic", payload)
}

// Google proxies a request to Google.
func (c *ProxyClient) Google(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "google", payload)
}

// XAI proxies a request to xAI.
func (c *ProxyClient) XAI(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "xai", payload)
}

// Meta proxies a request to Meta.
func (c *ProxyClient) Meta(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "meta", payload)
}

// OpenAIStream proxies a streaming request to OpenAI.
func (c *ProxyClient) OpenAIStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "openai", payload)
}

// AnthropicStream proxies a streaming request to Anthropic.
func (c *ProxyClient) AnthropicStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "anthropic", payload)
}

// GoogleStream proxies a streaming request to Google.
func (c *ProxyClient) GoogleStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "google", payload)
}

// XAIStream proxies a streaming request to xAI.
func (c *ProxyClient) XAIStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "xai", payload)
}

// MetaStream proxies a streaming request to Meta.
func (c *ProxyClient) MetaStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "meta", payload)
}
