package keel

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type httpTransport struct {
	baseURL    string
	apiKey     string
	freshness  bool
	httpClient *http.Client
	retryCfg   *RetryConfig
}

func (t *httpTransport) get(ctx context.Context, path string, opts *RequestOptions) ([]byte, error) {
	resp, err := t.doWithRetry(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return nil, err
	}
	return resp.body, nil
}

func (t *httpTransport) post(ctx context.Context, path string, body any, opts *RequestOptions) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("keel: marshal request: %w", err)
	}
	resp, err := t.doWithRetry(ctx, http.MethodPost, path, data, opts)
	if err != nil {
		return nil, err
	}
	return resp.body, nil
}

func (t *httpTransport) postStream(ctx context.Context, path string, body any, opts *RequestOptions) (io.ReadCloser, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("keel: marshal request: %w", err)
	}

	req, err := t.newRequest(ctx, http.MethodPost, path, data, opts)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("keel: request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, parseErrorResponse(resp.StatusCode, respBody, resp.Header)
	}

	return resp.Body, nil
}

func (t *httpTransport) doWithRetry(ctx context.Context, method, path string, body []byte, opts *RequestOptions) (*httpResponse, error) {
	resp, err := retryWithBackoff(ctx, t.retryCfg, func() (*httpResponse, error) {
		req, err := t.newRequest(ctx, method, path, body, opts)
		if err != nil {
			return nil, err
		}

		httpResp, err := t.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("keel: request failed: %w", err)
		}
		defer httpResp.Body.Close()

		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return nil, fmt.Errorf("keel: read response: %w", err)
		}

		var retryAfter time.Duration
		if ra := httpResp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil {
				retryAfter = time.Duration(secs) * time.Second
			}
		}

		return &httpResponse{
			statusCode: httpResp.StatusCode,
			body:       respBody,
			retryAfter: retryAfter,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	if resp.statusCode >= 400 {
		return nil, parseErrorResponse(resp.statusCode, resp.body, nil)
	}

	return resp, nil
}

func (t *httpTransport) newRequest(ctx context.Context, method, path string, body []byte, opts *RequestOptions) (*http.Request, error) {
	url := strings.TrimRight(t.baseURL, "/") + "/" + strings.TrimLeft(path, "/")

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("keel: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	if t.freshness {
		req.Header.Set("X-Keel-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
		nonce := make([]byte, 16)
		if _, err := rand.Read(nonce); err == nil {
			req.Header.Set("X-Keel-Nonce", base64.RawURLEncoding.EncodeToString(nonce))
		}
	}

	if opts != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	return req, nil
}

func parseErrorResponse(status int, body []byte, headers http.Header) error {
	ke := &KeelError{Status: status}

	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Field   string `json:"field"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil {
		ke.Code = errResp.Error.Code
		ke.Message = errResp.Error.Message
		ke.Field = errResp.Error.Field
	}

	if ke.Message == "" {
		ke.Message = string(body)
	}

	if headers != nil {
		if ra := headers.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil {
				ke.RetryAfter = time.Duration(secs) * time.Second
			}
		}
	}

	return ke
}
