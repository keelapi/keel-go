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
	return c.proxySyncWithHeaders(ctx, provider, payload, nil)
}

func (c *ProxyClient) proxySyncWithHeaders(ctx context.Context, provider string, payload any, headers map[string]string) (map[string]any, error) {
	body, err := c.t.post(ctx, "/v1/proxy/"+provider, payload, &RequestOptions{Headers: headers})
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
	return c.proxyStreamWithHeaders(ctx, provider, payload, nil)
}

func (c *ProxyClient) proxyStreamWithHeaders(ctx context.Context, provider string, payload any, headers map[string]string) (<-chan SSEEvent, <-chan error) {
	events := make(chan SSEEvent)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)

		rc, err := c.t.postStream(ctx, "/v1/proxy/"+provider, payload, &RequestOptions{Headers: headers})
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

// OpenAIWithHeaders proxies a request to OpenAI with additional request headers.
func (c *ProxyClient) OpenAIWithHeaders(ctx context.Context, payload any, headers map[string]string) (map[string]any, error) {
	return c.proxySyncWithHeaders(ctx, "openai", payload, headers)
}

// Anthropic proxies a request to Anthropic.
func (c *ProxyClient) Anthropic(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "anthropic", payload)
}

// AnthropicWithHeaders proxies a request to Anthropic with additional request headers.
func (c *ProxyClient) AnthropicWithHeaders(ctx context.Context, payload any, headers map[string]string) (map[string]any, error) {
	return c.proxySyncWithHeaders(ctx, "anthropic", payload, headers)
}

// Google proxies a request to Google.
func (c *ProxyClient) Google(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "google", payload)
}

// GoogleWithHeaders proxies a request to Google with additional request headers.
func (c *ProxyClient) GoogleWithHeaders(ctx context.Context, payload any, headers map[string]string) (map[string]any, error) {
	return c.proxySyncWithHeaders(ctx, "google", payload, headers)
}

// XAI proxies a request to xAI.
func (c *ProxyClient) XAI(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "xai", payload)
}

// XAIWithHeaders proxies a request to xAI with additional request headers.
func (c *ProxyClient) XAIWithHeaders(ctx context.Context, payload any, headers map[string]string) (map[string]any, error) {
	return c.proxySyncWithHeaders(ctx, "xai", payload, headers)
}

// Meta proxies a request to Meta.
func (c *ProxyClient) Meta(ctx context.Context, payload any) (map[string]any, error) {
	return c.proxySync(ctx, "meta", payload)
}

// MetaWithHeaders proxies a request to Meta with additional request headers.
func (c *ProxyClient) MetaWithHeaders(ctx context.Context, payload any, headers map[string]string) (map[string]any, error) {
	return c.proxySyncWithHeaders(ctx, "meta", payload, headers)
}

// OpenAIStream proxies a streaming request to OpenAI.
func (c *ProxyClient) OpenAIStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "openai", payload)
}

// OpenAIStreamWithHeaders proxies a streaming request to OpenAI with additional request headers.
func (c *ProxyClient) OpenAIStreamWithHeaders(ctx context.Context, payload any, headers map[string]string) (<-chan SSEEvent, <-chan error) {
	return c.proxyStreamWithHeaders(ctx, "openai", payload, headers)
}

// AnthropicStream proxies a streaming request to Anthropic.
func (c *ProxyClient) AnthropicStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "anthropic", payload)
}

// AnthropicStreamWithHeaders proxies a streaming request to Anthropic with additional request headers.
func (c *ProxyClient) AnthropicStreamWithHeaders(ctx context.Context, payload any, headers map[string]string) (<-chan SSEEvent, <-chan error) {
	return c.proxyStreamWithHeaders(ctx, "anthropic", payload, headers)
}

// GoogleStream proxies a streaming request to Google.
func (c *ProxyClient) GoogleStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "google", payload)
}

// GoogleStreamWithHeaders proxies a streaming request to Google with additional request headers.
func (c *ProxyClient) GoogleStreamWithHeaders(ctx context.Context, payload any, headers map[string]string) (<-chan SSEEvent, <-chan error) {
	return c.proxyStreamWithHeaders(ctx, "google", payload, headers)
}

// XAIStream proxies a streaming request to xAI.
func (c *ProxyClient) XAIStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "xai", payload)
}

// XAIStreamWithHeaders proxies a streaming request to xAI with additional request headers.
func (c *ProxyClient) XAIStreamWithHeaders(ctx context.Context, payload any, headers map[string]string) (<-chan SSEEvent, <-chan error) {
	return c.proxyStreamWithHeaders(ctx, "xai", payload, headers)
}

// MetaStream proxies a streaming request to Meta.
func (c *ProxyClient) MetaStream(ctx context.Context, payload any) (<-chan SSEEvent, <-chan error) {
	return c.proxyStream(ctx, "meta", payload)
}

// MetaStreamWithHeaders proxies a streaming request to Meta with additional request headers.
func (c *ProxyClient) MetaStreamWithHeaders(ctx context.Context, payload any, headers map[string]string) (<-chan SSEEvent, <-chan error) {
	return c.proxyStreamWithHeaders(ctx, "meta", payload, headers)
}
