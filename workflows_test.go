package keel

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewClient_WorkflowsClient(t *testing.T) {
	c := NewClient(ClientConfig{
		BaseURL: "https://api.keelapi.com",
		APIKey:  "test",
	})
	if c.Workflows == nil {
		t.Fatal("Workflows client is nil")
	}
}

func TestWorkflowsClient_Declare(t *testing.T) {
	now := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/workflows" {
			t.Errorf("expected /v1/workflows, got %s", r.URL.Path)
		}

		var req WorkflowDeclareRequest
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.WorkflowID != "wf_123" {
			t.Errorf("expected workflow_id wf_123, got %s", req.WorkflowID)
		}
		if req.Intent.ExpectedCalls == nil || *req.Intent.ExpectedCalls != 2 {
			t.Errorf("expected expected_calls 2, got %v", req.Intent.ExpectedCalls)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"workflow_id":                "wf_123",
			"decision":                   "accepted",
			"status":                     "active",
			"id":                         "4be0df33-9aa3-4786-ab27-56d7ba3f4db6",
			"version":                    1,
			"expected_calls":             2,
			"max_calls":                  4,
			"cached_actual_calls":        0,
			"declaration_canonical_hash": "hash_123",
			"declaration_signature_b64":  "sig_123",
			"declared_at":                now.Format(time.RFC3339),
			"projected_cost": map[string]any{
				"amount_micros": 22500000,
				"currency":      "USD",
				"methodology": map[string]any{
					"basis":      "caller_declared_workflow_x_point_pricing",
					"provenance": "caller_declared_workflow",
					"quality":    "authoritative",
				},
			},
		})
	})

	expectedCalls := 2
	maxCalls := 4
	resp, err := c.Workflows.Declare(context.Background(), WorkflowDeclareRequest{
		WorkflowID: "wf_123",
		Intent: WorkflowIntent{
			ExpectedCalls: &expectedCalls,
			MaxCalls:      &maxCalls,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.WorkflowID != "wf_123" {
		t.Errorf("expected wf_123, got %s", resp.WorkflowID)
	}
	if resp.DeclarationSignatureB64 == nil || *resp.DeclarationSignatureB64 != "sig_123" {
		t.Fatalf("expected signed declaration, got %v", resp.DeclarationSignatureB64)
	}
	if resp.DeclarationCanonicalHash == nil || *resp.DeclarationCanonicalHash != "hash_123" {
		t.Fatalf("expected canonical hash, got %v", resp.DeclarationCanonicalHash)
	}
}

func TestWorkflowsClient_Amend_VersionConflict(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/workflows/wf_123/amend" {
			t.Errorf("expected amend path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":            ReasonWorkflowAmendmentVersionConflict,
				"message":         "workflow declaration version does not match",
				"current_version": 2,
			},
		})
	})

	_, err := c.Workflows.Amend(context.Background(), "wf_123", WorkflowAmendRequest{
		IfMatchVersion: 1,
		NewMaxCalls:    intPtrForTest(6),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrWorkflowAmendmentVersionConflict) {
		t.Fatalf("expected amendment version conflict, got %v", err)
	}
	var ke *KeelError
	if !errors.As(err, &ke) {
		t.Fatalf("expected KeelError, got %T", err)
	}
	if ke.Code != ReasonWorkflowAmendmentVersionConflict {
		t.Errorf("expected code %s, got %s", ReasonWorkflowAmendmentVersionConflict, ke.Code)
	}
}

func TestWorkflowsClient_Complete(t *testing.T) {
	completedAt := time.Date(2026, 5, 12, 11, 0, 0, 0, time.UTC)
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/workflows/wf_123/complete" {
			t.Errorf("expected complete path, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"workflow_id":                "wf_123",
			"status":                     "completed",
			"cached_actual_calls":        3,
			"authoritative_actual_calls": 3,
			"counter_reconciled":         true,
			"completed_at":               completedAt.Format(time.RFC3339),
		})
	})

	resp, err := c.Workflows.Complete(context.Background(), "wf_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.WorkflowID != "wf_123" {
		t.Errorf("expected wf_123, got %s", resp.WorkflowID)
	}
	if resp.Status != WorkflowStatusCompleted {
		t.Errorf("expected completed, got %s", resp.Status)
	}
	if resp.AuthoritativeActualCalls != 3 {
		t.Errorf("expected 3 calls, got %d", resp.AuthoritativeActualCalls)
	}
}

func TestWithWorkflow_PropagatesViaContext(t *testing.T) {
	ctx := WithWorkflow(context.Background(), " wf_123 ")
	workflowID, ok := WorkflowFromContext(ctx)
	if !ok {
		t.Fatal("expected workflow ID in context")
	}
	if workflowID != "wf_123" {
		t.Errorf("expected wf_123, got %s", workflowID)
	}
}

func TestRunInWorkflow_AutoInjectsHeader(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(workflowIDHeader); got != "wf_123" {
			t.Errorf("expected workflow header wf_123, got %q", got)
		}
		json.NewEncoder(w).Encode(PermitAuditItem{PermitID: "pmt_123"})
	})

	err := RunInWorkflow(context.Background(), "wf_123", func(ctx context.Context) error {
		_, err := c.Permits.Get(ctx, "pmt_123")
		return err
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInWorkflow_HeaderAbsentOutside(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(workflowIDHeader); got != "" {
			t.Errorf("expected no workflow header, got %q", got)
		}
		json.NewEncoder(w).Encode(PermitAuditItem{PermitID: "pmt_123"})
	})

	_, err := c.Permits.Get(context.Background(), "pmt_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInWorkflow_PropagatesAcrossGoroutines(t *testing.T) {
	err := RunInWorkflow(context.Background(), "wf_goroutine", func(ctx context.Context) error {
		result := make(chan string, 1)
		go func(ctx context.Context) {
			workflowID, _ := WorkflowFromContext(ctx)
			result <- workflowID
		}(ctx)

		select {
		case workflowID := <-result:
			if workflowID != "wf_goroutine" {
				t.Errorf("expected wf_goroutine, got %s", workflowID)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for goroutine")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkflowMaxCallsExceeded_TypedError(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    ReasonWorkflowMaxCallsExceeded,
				"message": "Workflow max_calls has been exceeded.",
			},
		})
	})

	_, err := c.Permits.Get(context.Background(), "pmt_123")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrWorkflowMaxCallsExceeded) {
		t.Fatalf("expected max calls typed error, got %v", err)
	}
}

func TestHTTPHeader_NotForwardedToProvider(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/proxy/openai" {
			t.Errorf("expected /v1/proxy/openai, got %s", r.URL.Path)
		}
		if got := r.Header.Get(workflowIDHeader); got != "wf_proxy" {
			t.Errorf("expected SDK-to-Keel workflow header wf_proxy, got %q", got)
		}

		var payload map[string]any
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if containsWorkflowHeaderKey(payload) {
			t.Fatalf("workflow header leaked into provider payload: %v", payload)
		}

		json.NewEncoder(w).Encode(map[string]any{"id": "chatcmpl_123"})
	})

	err := RunInWorkflow(context.Background(), "wf_proxy", func(ctx context.Context) error {
		_, err := c.Proxy.OpenAI(ctx, map[string]any{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "hello"},
			},
		})
		return err
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func intPtrForTest(v int) *int {
	return &v
}

func containsWorkflowHeaderKey(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			if strings.EqualFold(key, workflowIDHeader) {
				return true
			}
			if containsWorkflowHeaderKey(item) {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if containsWorkflowHeaderKey(item) {
				return true
			}
		}
	}
	return false
}
