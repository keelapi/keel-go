package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := NewClient(ClientConfig{
		BaseURL: ts.URL,
		APIKey:  "test-key",
		RetryConfig: &RetryConfig{
			MaxRetries:           0,
			InitialDelay:         time.Millisecond,
			MaxDelay:             time.Millisecond,
			BackoffMultiplier:    1,
			RetryableStatusCodes: map[int]bool{},
		},
	})
	return c, ts
}

func TestNewClient(t *testing.T) {
	c := NewClient(ClientConfig{
		BaseURL: "https://api.keelapi.com",
		APIKey:  "test",
	})
	if c.Permits == nil {
		t.Fatal("Permits client is nil")
	}
	if c.Executions == nil {
		t.Fatal("Executions client is nil")
	}
	if c.Execute == nil {
		t.Fatal("Execute client is nil")
	}
	if c.Proxy == nil {
		t.Fatal("Proxy client is nil")
	}
	if c.Jobs == nil {
		t.Fatal("Jobs client is nil")
	}
	if c.ApiKeys == nil {
		t.Fatal("ApiKeys client is nil")
	}
	if c.Requests == nil {
		t.Fatal("Requests client is nil")
	}
}

func TestPermitsCreate(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/permits" {
			t.Errorf("expected /v1/permits, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing auth header")
		}

		var req PermitRequest
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)

		if req.Subject.ID != "user-1" {
			t.Errorf("expected subject ID user-1, got %s", req.Subject.ID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PermitResponse{
			PermitID: "pmt_123",
			Decision: DecisionAllow,
		})
	})

	resp, err := c.Permits.Create(context.Background(), PermitRequest{
		ProjectID:      "proj-1",
		IdempotencyKey: "test-1",
		Subject:        Subject{Type: "user", ID: "user-1"},
		Action:         Action{Name: string(OpGenerateText)},
		Resource:       Resource{Type: "ai_model", ID: "gpt-4", Attributes: ResourceAttributes{Provider: "openai", Model: "gpt-4", Operation: OpGenerateText}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PermitID != "pmt_123" {
		t.Errorf("expected pmt_123, got %s", resp.PermitID)
	}
	if resp.Decision != DecisionAllow {
		t.Errorf("expected allow, got %s", resp.Decision)
	}
}

func TestPermitsDryRun(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/permits/dry-run" {
			t.Errorf("expected /v1/permits/dry-run, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(PermitDryRunResponse{
			Decision: DecisionDeny,
			Reason:   "budget exceeded",
		})
	})

	resp, err := c.Permits.DryRun(context.Background(), PermitRequest{
		ProjectID:      "proj-1",
		IdempotencyKey: "test-dry-1",
		Subject:        Subject{Type: "user", ID: "user-1"},
		Action:         Action{Name: string(OpGenerateText)},
		Resource:       Resource{Type: "ai_model", ID: "gpt-4", Attributes: ResourceAttributes{Provider: "openai", Model: "gpt-4", Operation: OpGenerateText}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Decision != DecisionDeny {
		t.Errorf("expected deny, got %s", resp.Decision)
	}
}

func TestPermitsList(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		json.NewEncoder(w).Encode(PermitsListResponse{
			Items: []PermitAuditItem{{PermitID: "pmt_1"}},
		})
	})

	limit := 10
	resp, err := c.Permits.List(context.Background(), PermitListParams{Limit: &limit})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestPermitsGet(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/permits/pmt_123" {
			t.Errorf("expected /v1/permits/pmt_123, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(PermitAuditItem{PermitID: "pmt_123"})
	})

	resp, err := c.Permits.Get(context.Background(), "pmt_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PermitID != "pmt_123" {
		t.Errorf("expected pmt_123, got %s", resp.PermitID)
	}
}

func TestExecutionsCreate(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/executions" {
			t.Errorf("expected /v1/executions, got %s", r.URL.Path)
		}
		var req ExecutionCreateRequest
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		if req.Mode != ModeSync {
			t.Errorf("expected sync mode, got %s", req.Mode)
		}
		json.NewEncoder(w).Encode(ExecutionResponse{
			RequestID: "req_123",
			PermitID:  "pmt_123",
			Status:    "completed",
		})
	})

	resp, err := c.Executions.Create(context.Background(), ExecutionCreateRequest{
		Operation: string(OpGenerateText),
		Messages:  []MessageInput{{Role: "user", Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RequestID != "req_123" {
		t.Errorf("expected req_123, got %s", resp.RequestID)
	}
}

func TestExecuteRun(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/execute" {
			t.Errorf("expected /v1/execute, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ExecutionResponse{
			RequestID: "req_456",
			Status:    "completed",
		})
	})

	resp, err := c.Execute.Run(context.Background(), ExecuteRequest{
		Subject:  Subject{Type: "user", ID: "user-1"},
		Action:   Action{Name: string(OpGenerateText)},
		Resource: Resource{Type: "ai_model", ID: "gpt-4", Attributes: ResourceAttributes{Provider: "openai", Model: "gpt-4", Operation: OpGenerateText}},
		Messages: []MessageInput{{Role: "user", Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RequestID != "req_456" {
		t.Errorf("expected req_456, got %s", resp.RequestID)
	}
}

func TestProxyOpenAI(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/proxy/openai" {
			t.Errorf("expected /v1/proxy/openai, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "chatcmpl-123"})
	})

	resp, err := c.Proxy.OpenAI(context.Background(), map[string]any{"model": "gpt-4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["id"] != "chatcmpl-123" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestJobsCreateAndGet(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/jobs":
			json.NewEncoder(w).Encode(JobCreateResponse{
				JobID: "job_123", Status: JobSubmitted,
			})
		case "/v1/jobs/job_123":
			json.NewEncoder(w).Encode(JobStatusResponse{
				JobID: "job_123", Status: JobCompleted,
			})
		}
	})

	created, err := c.Jobs.Create(context.Background(), JobSubmitRequest{
		Permit: PermitRequest{
			ProjectID:      "proj-1",
			IdempotencyKey: "job-test-1",
			Subject:        Subject{Type: "user", ID: "user-1"},
			Action:         Action{Name: "run.batch"},
			Resource:       Resource{Type: "ai_model", ID: "gpt-4", Attributes: ResourceAttributes{Provider: "openai", Model: "gpt-4", Operation: OpRunBatch}},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.JobID != "job_123" {
		t.Errorf("expected job_123, got %s", created.JobID)
	}

	status, err := c.Jobs.Get(context.Background(), "job_123")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if status.Status != JobCompleted {
		t.Errorf("expected completed, got %s", status.Status)
	}
}


func TestRequestsTimeline(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/requests/req_123/timeline" {
			t.Errorf("expected /v1/requests/req_123/timeline, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(RequestTimelineResponse{
			RequestID: "req_123",
			Events:    []RequestTimelineEvent{{Phase: "permit", Timestamp: "2026-01-01T00:00:00Z"}},
		})
	})

	resp, err := c.Requests.Timeline(context.Background(), "req_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(resp.Events))
	}
}

func TestErrorParsing(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    "rate_limit",
				"message": "too many requests",
			},
		})
	})

	_, err := c.Permits.Get(context.Background(), "pmt_123")
	if err == nil {
		t.Fatal("expected error")
	}
	ke, ok := err.(*KeelError)
	if !ok {
		t.Fatalf("expected KeelError, got %T", err)
	}
	if ke.Status != 429 {
		t.Errorf("expected 429, got %d", ke.Status)
	}
	if ke.Code != "rate_limit" {
		t.Errorf("expected rate_limit, got %s", ke.Code)
	}
	if !ke.IsRetryable() {
		t.Error("expected retryable")
	}
}

func TestKeelErrorNotRetryable(t *testing.T) {
	tests := []struct {
		status    int
		retryable bool
	}{
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{408, true},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.status), func(t *testing.T) {
			e := &KeelError{Status: tt.status}
			if e.IsRetryable() != tt.retryable {
				t.Errorf("status %d: expected retryable=%v", tt.status, tt.retryable)
			}
		})
	}
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	rc := &RetryConfig{
		MaxRetries:           2,
		InitialDelay:         time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		BackoffMultiplier:    2.0,
		RetryableStatusCodes: map[int]bool{500: true},
	}

	resp, err := retryWithBackoff(context.Background(), rc, func() (*httpResponse, error) {
		attempts++
		if attempts < 3 {
			return &httpResponse{statusCode: 500, body: []byte(`{}`)}, nil
		}
		return &httpResponse{statusCode: 200, body: []byte(`{"ok":true}`)}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 200 {
		t.Errorf("expected 200, got %d", resp.statusCode)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestSSEParsing(t *testing.T) {
	sseData := "event: message_start\ndata: {\"type\":\"message_start\"}\n\n: comment\nevent: content_block_delta\ndata: {\"type\":\"delta\"}\ndata: {\"more\":true}\n\n"

	r := io.NopCloser(strings.NewReader(sseData))
	events, errc := parseSSEStream(r)

	var received []SSEEvent
	for e := range events {
		received = append(received, e)
	}
	if err := <-errc; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}

	if received[0].EventType != "message_start" {
		t.Errorf("expected message_start, got %s", received[0].EventType)
	}
	if received[1].EventType != "content_block_delta" {
		t.Errorf("expected content_block_delta, got %s", received[1].EventType)
	}
	// Multiline data should be joined
	if !strings.Contains(string(received[1].Data), "{\"type\":\"delta\"}") {
		t.Errorf("unexpected data: %s", received[1].Data)
	}
}

func TestRequestFreshness(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Keel-Timestamp") == "" {
			t.Error("missing X-Keel-Timestamp")
		}
		if r.Header.Get("X-Keel-Nonce") == "" {
			t.Error("missing X-Keel-Nonce")
		}
		json.NewEncoder(w).Encode(PermitAuditItem{PermitID: "pmt_123"})
	})
	c.transport.freshness = true

	_, err := c.Permits.Get(context.Background(), "pmt_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecutionsStream(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: execution.started\ndata: {\"request_id\":\"req_1\"}\n\nevent: content.delta\ndata: {\"text\":\"hello\"}\n\nevent: execution.done\ndata: {\"request_id\":\"req_1\"}\n\n")
	})

	events, errc := c.Executions.Stream(context.Background(), ExecutionCreateRequest{
		Operation: string(OpGenerateText),
		Messages:  []MessageInput{{Role: "user", Content: "Hi"}},
	})

	var received []ExecutionStreamEvent
	for e := range events {
		received = append(received, e)
	}
	if err := <-errc; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 3 {
		t.Fatalf("expected 3 events, got %d", len(received))
	}
	if received[0].EventType != "execution.started" {
		t.Errorf("expected execution.started, got %s", received[0].EventType)
	}
}

func TestProxyStream(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/proxy/anthropic" {
			t.Errorf("expected /v1/proxy/anthropic, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\"}\n\n")
	})

	events, errc := c.Proxy.AnthropicStream(context.Background(), map[string]any{"model": "claude-3"})
	var received []SSEEvent
	for e := range events {
		received = append(received, e)
	}
	if err := <-errc; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
}
