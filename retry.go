package keel

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"time"
)

// RetryConfig configures retry behavior for failed requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts. Defaults to 3.
	MaxRetries int

	// InitialDelay is the delay before the first retry. Defaults to 500ms.
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries. Defaults to 30s.
	MaxDelay time.Duration

	// BackoffMultiplier is the exponential backoff factor. Defaults to 2.0.
	BackoffMultiplier float64

	// RetryableStatusCodes defines which HTTP status codes are retryable.
	// Defaults to {408, 429, 500, 502, 503, 504}.
	RetryableStatusCodes map[int]bool
}

func defaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      500 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableStatusCodes: map[int]bool{
			408: true, 429: true, 500: true, 502: true, 503: true, 504: true,
		},
	}
}

func (rc *RetryConfig) isRetryable(statusCode int) bool {
	return rc.RetryableStatusCodes[statusCode]
}

func (rc *RetryConfig) delay(attempt int) time.Duration {
	delay := float64(rc.InitialDelay) * math.Pow(rc.BackoffMultiplier, float64(attempt))
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}
	// Add 0-25% jitter
	jitter := jitterFraction(0.25)
	delay = delay * (1.0 + jitter)
	return time.Duration(delay)
}

func jitterFraction(max float64) float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return 0
	}
	return max * float64(n.Int64()) / 1000.0
}

type retryableFunc func() (*httpResponse, error)

type httpResponse struct {
	statusCode int
	body       []byte
	retryAfter time.Duration
}

func retryWithBackoff(ctx context.Context, rc *RetryConfig, fn retryableFunc) (*httpResponse, error) {
	var lastResp *httpResponse
	var lastErr error

	for attempt := 0; attempt <= rc.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rc.delay(attempt - 1)
			if lastResp != nil && lastResp.retryAfter > 0 {
				delay = lastResp.retryAfter
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := fn()
		if err != nil {
			lastErr = err
			continue
		}

		if !rc.isRetryable(resp.statusCode) {
			return resp, nil
		}

		lastResp = resp
		lastErr = nil
	}

	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}
