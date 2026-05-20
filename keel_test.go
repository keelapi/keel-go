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
	if c.Compliance == nil {
		t.Fatal("Compliance client is nil")
	}
	if c.Integrity == nil {
		t.Fatal("Integrity client is nil")
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
			Items: []PermitAuditItem{{ID: "pmt_1"}},
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
		json.NewEncoder(w).Encode(PermitAuditItem{ID: "pmt_123"})
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
			ID:           "exec_req_123",
			Object:       "execution",
			Status:       "completed",
			StatusCode:   200,
			OutputAssets: []ExecutionOutputAsset{},
			Routing: &ExecutionRoutingResult{
				SelectedProvider: ProviderOpenAI,
				SelectedModel:    "gpt-4",
				ReasonCode:       "explicit_request",
			},
		})
	})

	resp, err := c.Executions.Create(context.Background(), ExecutionCreateRequest{
		Operation: string(OpGenerateText),
		Messages:  []MessageInput{{Role: "user", Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "exec_req_123" {
		t.Errorf("expected exec_req_123, got %s", resp.ID)
	}
}

func TestExecuteRun(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/execute" {
			t.Errorf("expected /v1/execute, got %s", r.URL.Path)
		}
		var req ExecuteRequest
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		if req.Model != "gpt-4" {
			t.Errorf("expected model gpt-4, got %s", req.Model)
		}
		json.NewEncoder(w).Encode(ExecutionResponse{
			ID:           "exec_req_456",
			Object:       "execution",
			Status:       "completed",
			StatusCode:   200,
			OutputAssets: []ExecutionOutputAsset{},
			Routing: &ExecutionRoutingResult{
				SelectedProvider: ProviderOpenAI,
				SelectedModel:    "gpt-4",
				ReasonCode:       "explicit_request",
			},
		})
	})

	provider := ProviderOpenAI
	resp, err := c.Execute.Run(context.Background(), ExecuteRequest{
		Provider: &provider,
		Model:    "gpt-4",
		Input: map[string]any{
			"messages": []map[string]any{{"role": "user", "content": "Hello"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "exec_req_456" {
		t.Errorf("expected exec_req_456, got %s", resp.ID)
	}
}

func TestExecuteRequestMarshalLegacyFields(t *testing.T) {
	data, err := json.Marshal(ExecuteRequest{
		Subject:  Subject{Type: "user", ID: "user_123"},
		Action:   Action{Name: string(OpGenerateText)},
		Resource: Resource{Type: "ai_model", ID: "gpt-4", Attributes: ResourceAttributes{Provider: string(ProviderOpenAI), Model: "gpt-4"}},
		Messages: []MessageInput{{Role: "user", Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("marshal legacy execute request: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["provider"] != string(ProviderOpenAI) {
		t.Fatalf("expected provider openai, got %v", payload["provider"])
	}
	if payload["model"] != "gpt-4" {
		t.Fatalf("expected model gpt-4, got %v", payload["model"])
	}
	input, ok := payload["input"].(map[string]any)
	if !ok {
		t.Fatalf("expected input object, got %T", payload["input"])
	}
	if _, ok := input["messages"]; !ok {
		t.Fatalf("expected messages in input, got %+v", input)
	}
	if _, ok := input["subject"]; !ok {
		t.Fatalf("expected subject in input, got %+v", input)
	}
}

func TestExecutionResponseDecodesEnvelopeCompatibilityFields(t *testing.T) {
	payload := []byte(`{
		"id":"exec_req_789",
		"object":"execution",
		"status":"completed",
		"status_code":200,
		"output_assets":[],
		"routing":{
			"selected_provider":"openai",
			"selected_model":"gpt-4o-mini",
			"reason_code":"explicit_request",
			"fallback_occurred":false
		},
		"usage":{
			"input_tokens":12,
			"output_tokens":5,
			"total_tokens":17,
			"estimated_final":false
		},
		"timing":{
			"started_at":"2026-05-20T00:00:00Z",
			"completed_at":"2026-05-20T00:00:00.012Z",
			"duration_ms":12.5
		}
	}`)

	var resp ExecutionResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("decode execution envelope: %v", err)
	}
	if resp.Provider != ProviderOpenAI || resp.Model != "gpt-4o-mini" {
		t.Fatalf("legacy provider/model aliases not populated: %+v", resp)
	}
	if resp.Routing == nil || resp.Routing.Provider != ProviderOpenAI || resp.Routing.Model != "gpt-4o-mini" {
		t.Fatalf("routing aliases not populated: %+v", resp.Routing)
	}
	if resp.Usage == nil || resp.Usage.InputTokens != 12 || resp.Usage.TotalTokens != 17 {
		t.Fatalf("usage not decoded as scalars: %+v", resp.Usage)
	}
	if resp.Timing == nil || resp.Timing.DurationMS == nil || *resp.Timing.DurationMS != 12.5 || resp.Timing.TotalMS != 12 {
		t.Fatalf("timing not decoded with compatibility total: %+v", resp.Timing)
	}
}

func TestLegacyContentMapsAvoidForbiddenTopLevelExtras(t *testing.T) {
	inputData, err := json.Marshal(PermitInput{
		Type: "image",
		Content: map[string]any{
			"url":             "https://example.com/image.png",
			"classification":  "confidential",
			"metadata":        map[string]any{"source": "legacy"},
			"unsupported_key": "moved",
		},
	})
	if err != nil {
		t.Fatalf("marshal permit input: %v", err)
	}
	var input map[string]any
	if err := json.Unmarshal(inputData, &input); err != nil {
		t.Fatalf("decode permit input: %v", err)
	}
	if input["url"] != "https://example.com/image.png" {
		t.Fatalf("expected known field promoted, got %+v", input)
	}
	if _, ok := input["classification"]; ok {
		t.Fatalf("unexpected top-level legacy field: %+v", input)
	}
	metadata, ok := input["metadata"].(map[string]any)
	if !ok || metadata["classification"] != "confidential" || metadata["unsupported_key"] != "moved" {
		t.Fatalf("expected unknown legacy fields in metadata, got %+v", input)
	}

	partData, err := json.Marshal(InputPart{
		Type: "text",
		Content: map[string]any{
			"text":       "hello",
			"trace_hint": "legacy",
		},
	})
	if err != nil {
		t.Fatalf("marshal input part: %v", err)
	}
	var part map[string]any
	if err := json.Unmarshal(partData, &part); err != nil {
		t.Fatalf("decode input part: %v", err)
	}
	if part["text"] != "hello" {
		t.Fatalf("expected text promoted, got %+v", part)
	}
	if _, ok := part["trace_hint"]; ok {
		t.Fatalf("unexpected top-level legacy field: %+v", part)
	}
}

func TestBuildManagedProxyPayload(t *testing.T) {
	parent := "permit_parent"
	session := "session_123"
	subject := PermitSubject{Type: "api_key", ID: "key_123"}
	payload := BuildManagedProxyPayload(
		map[string]any{
			"model":    "gpt-4o-mini",
			"messages": []any{},
			"metadata": map[string]any{"caller": "test"},
		},
		"proj_123",
		&subject,
		&parent,
		&session,
	)

	if payload["project_id"] != "proj_123" {
		t.Fatalf("expected project_id, got %+v", payload)
	}
	metadata, ok := payload["metadata"].(map[string]any)
	if !ok || metadata["caller"] != "test" || metadata["keel_parent_permit_id"] != parent || metadata["keel_session_id"] != session {
		t.Fatalf("unexpected metadata: %+v", payload["metadata"])
	}
	identity, ok := payload["identity"].(*Identity)
	if !ok || identity.Credential == nil || identity.Credential.ID == nil || *identity.Credential.ID != "key_123" {
		t.Fatalf("unexpected identity: %+v", payload["identity"])
	}
}

func TestExecutionsCreateDecodesNon2xxEnvelope(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{
			"id":            "exec_req_denied",
			"object":        "execution",
			"status":        "denied",
			"status_code":   403,
			"output":        nil,
			"output_assets": []any{},
			"routing": map[string]any{
				"selected_provider": "openai",
				"selected_model":    "gpt-4",
				"reason_code":       "explicit_request",
				"fallback_occurred": false,
			},
			"timing": map[string]any{},
			"error": map[string]any{
				"stage":       "permit",
				"code":        "policy_denied",
				"message":     "blocked",
				"status_code": 403,
			},
		})
	})

	resp, err := c.Executions.Create(context.Background(), ExecutionCreateRequest{
		Operation: string(OpGenerateText),
		Messages:  []MessageInput{{Role: "user", Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "denied" || resp.StatusCode != 403 {
		t.Fatalf("unexpected envelope: %+v", resp)
	}
	if resp.Error == nil || resp.Error.Code != "policy_denied" {
		t.Fatalf("expected policy_denied error, got %+v", resp.Error)
	}
}

func TestExecuteRunDecodesNon2xxEnvelope(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]any{
			"id":            "exec_req_failed",
			"object":        "execution",
			"status":        "failed",
			"status_code":   502,
			"output":        nil,
			"output_assets": []any{},
			"routing": map[string]any{
				"selected_provider": "openai",
				"selected_model":    "gpt-4",
				"reason_code":       "explicit_request",
				"fallback_occurred": false,
			},
			"timing": map[string]any{},
			"error": map[string]any{
				"stage":       "dispatch",
				"code":        "upstream_error",
				"message":     "timeout",
				"status_code": 502,
			},
			"resolved": map[string]any{"provider": "openai", "model": "gpt-4"},
		})
	})

	provider := ProviderOpenAI
	resp, err := c.Execute.Run(context.Background(), ExecuteRequest{
		Provider: &provider,
		Model:    "gpt-4",
		Input:    map[string]any{"messages": []map[string]any{{"role": "user", "content": "Hello"}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "failed" || resp.StatusCode != 502 {
		t.Fatalf("unexpected envelope: %+v", resp)
	}
	if resp.Resolved == nil || resp.Resolved.Provider == nil || *resp.Resolved.Provider != "openai" {
		t.Fatalf("expected resolved provider, got %+v", resp.Resolved)
	}
}

func TestPermitExportsCompatibility(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/permits/export" {
			t.Errorf("expected /v1/permits/export, got %s", r.URL.Path)
		}
		switch r.URL.Query().Get("format") {
		case "json":
			json.NewEncoder(w).Encode(map[string]any{
				"bundle_type":      "audit_export",
				"schema_version":   2,
				"format":           "json",
				"record_count":     1,
				"records":          []any{},
				"generated_at":     "2026-05-20T00:00:00Z",
				"project_id":       "proj_1",
				"time_range_basis": "created_at",
				"from":             "2026-05-01T00:00:00Z",
				"to":               "2026-05-20T00:00:00Z",
			})
		case "csv":
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write([]byte("permit_id,decision\npermit_1,allow\n"))
		default:
			t.Errorf("unexpected format %s", r.URL.Query().Get("format"))
		}
	})

	legacy, err := c.Permits.Export(context.Background(), "2026-05-01T00:00:00Z", "2026-05-20T00:00:00Z", "json")
	if err != nil {
		t.Fatalf("legacy export: %v", err)
	}
	if legacy.Count != 1 || !strings.Contains(legacy.Data, `"bundle_type"`) {
		t.Fatalf("unexpected legacy export wrapper: %+v", legacy)
	}

	bundle, err := c.Permits.ExportBundle(context.Background(), "2026-05-01T00:00:00Z", "2026-05-20T00:00:00Z")
	if err != nil {
		t.Fatalf("bundle export: %v", err)
	}
	if bundle.RecordCount != 1 || bundle.Format != "json" {
		t.Fatalf("unexpected bundle: %+v", bundle)
	}

	csv, err := c.Permits.ExportCSV(context.Background(), "2026-05-01T00:00:00Z", "2026-05-20T00:00:00Z")
	if err != nil {
		t.Fatalf("csv export: %v", err)
	}
	if !strings.Contains(csv, "permit_id,decision") {
		t.Fatalf("unexpected csv: %q", csv)
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

func TestApiKeysListAndRevoke(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/api-keys":
			if r.URL.Query().Get("cursor") != "cur_1" {
				t.Errorf("expected cursor=cur_1, got %s", r.URL.Query().Get("cursor"))
			}
			if r.URL.Query().Get("scope") != "client" {
				t.Errorf("expected scope=client, got %s", r.URL.Query().Get("scope"))
			}
			next := "cur_2"
			json.NewEncoder(w).Encode(ApiKeyListResponse{
				Items:      []ApiKeyRecord{{ID: "key_1", Prefix: "keel_sk_", Scope: "client"}},
				NextCursor: &next,
			})
		case "/v1/api-keys/key_1/revoke":
			revokedAt := "2026-01-01T00:00:00Z"
			json.NewEncoder(w).Encode(ApiKeyRecord{
				ID:        "key_1",
				Prefix:    "keel_sk_",
				Scope:     "client",
				RevokedAt: &revokedAt,
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	cursor := "cur_1"
	scope := "client"
	list, err := c.ApiKeys.List(context.Background(), ApiKeyListParams{Cursor: &cursor, Scope: &scope})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if list.NextCursor == nil || *list.NextCursor != "cur_2" {
		t.Fatalf("expected next cursor cur_2, got %+v", list.NextCursor)
	}

	if err := c.ApiKeys.Revoke(context.Background(), "key_1"); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	revoked, err := c.ApiKeys.RevokeRecord(context.Background(), "key_1")
	if err != nil {
		t.Fatalf("revoke record: %v", err)
	}
	if revoked.RevokedAt == nil {
		t.Fatal("expected revoked_at")
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

func TestComplianceAndIntegrityClients(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/compliance/exports":
			if r.Method == http.MethodPost {
				if r.URL.Query().Get("include_chain_entries") != "true" {
					t.Errorf("expected include_chain_entries=true, got %s", r.URL.Query().Get("include_chain_entries"))
				}
				json.NewEncoder(w).Encode(ComplianceExportResponse{
					ExportID:   "exp_1",
					ProjectID:  "proj_1",
					ExportType: "full_audit",
					Status:     "queued",
					Filters:    map[string]any{},
				})
				return
			}
			json.NewEncoder(w).Encode(ComplianceExportListResponse{Items: []ComplianceExportResponse{{ExportID: "exp_1"}}})
		case "/v1/compliance/keys":
			json.NewEncoder(w).Encode(map[string]any{"keys": []any{}})
		case "/v1/integrity/checkpoint-public-key":
			json.NewEncoder(w).Encode(CheckpointPublicKeyResponse{
				Algorithm: "ed25519",
				PublicKey: "ed25519:test",
				KeyID:     "key_1",
				Scope:     "integrity_checkpoints",
			})
		case "/v1/integrity/permit-binding-public-keys":
			json.NewEncoder(w).Encode(PermitBindingPublicKeysResponse{Purpose: "permit_binding_signing"})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	created, err := c.Compliance.CreateExport(context.Background(), ComplianceExportCreateRequest{ExportType: "full_audit"}, true)
	if err != nil {
		t.Fatalf("create export: %v", err)
	}
	if created.ExportID != "exp_1" {
		t.Fatalf("expected exp_1, got %s", created.ExportID)
	}
	if _, err := c.Compliance.ListExports(context.Background()); err != nil {
		t.Fatalf("list exports: %v", err)
	}
	if _, err := c.Compliance.Keys(context.Background()); err != nil {
		t.Fatalf("keys: %v", err)
	}
	if _, err := c.Integrity.CheckpointPublicKey(context.Background()); err != nil {
		t.Fatalf("checkpoint key: %v", err)
	}
	if _, err := c.Integrity.PermitBindingPublicKeys(context.Background()); err != nil {
		t.Fatalf("permit binding keys: %v", err)
	}
}

func TestErrorParsing(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{
			"permit": map[string]any{
				"decision":    "throttled",
				"reason_code": "budget.rate_limit_throttled",
				"message":     "too many requests",
			},
		})
	})

	_, err := c.Permits.Get(context.Background(), "pmt_123")
	if err == nil {
		t.Fatal("expected error")
	}
	te, ok := err.(*ThrottledError)
	if !ok {
		t.Fatalf("expected ThrottledError, got %T", err)
	}
	if te.RetryAfterSeconds != 30 {
		t.Errorf("expected retry_after 30, got %d", te.RetryAfterSeconds)
	}
	if te.ReasonCode != "budget.rate_limit_throttled" {
		t.Errorf("expected budget.rate_limit_throttled, got %s", te.ReasonCode)
	}
	if !te.IsRetryable() {
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
		json.NewEncoder(w).Encode(PermitAuditItem{ID: "pmt_123"})
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

func TestThrottled429ReturnsThrottledError(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]any{
			"permit": map[string]any{
				"permit_id":   "pmt_abc",
				"decision":    "throttled",
				"reason_code": "budget.rate_limit_throttled",
				"message":     "Rate limit throttled.",
				"outcome_detail": map[string]any{
					"retry_after_seconds": 30,
				},
			},
		})
	})

	_, err := c.Permits.Get(context.Background(), "pmt_abc")
	if err == nil {
		t.Fatal("expected error")
	}
	te, ok := err.(*ThrottledError)
	if !ok {
		t.Fatalf("expected ThrottledError, got %T: %v", err, err)
	}
	if te.PermitID != "pmt_abc" {
		t.Errorf("expected permit_id pmt_abc, got %s", te.PermitID)
	}
	if te.ReasonCode != "budget.rate_limit_throttled" {
		t.Errorf("expected reason_code budget.rate_limit_throttled, got %s", te.ReasonCode)
	}
	if te.RetryAfterSeconds != 30 {
		t.Errorf("expected retry_after_seconds 30, got %d", te.RetryAfterSeconds)
	}
	if !te.IsRetryable() {
		t.Error("expected retryable")
	}
}

func TestThrottled429BodyFallback(t *testing.T) {
	// No Retry-After header; falls back to body's retry_after_seconds
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]any{
			"permit": map[string]any{
				"decision":    "throttled",
				"reason_code": "budget.rate_limit_throttled",
				"outcome_detail": map[string]any{
					"retry_after_seconds": 15,
				},
			},
		})
	})

	_, err := c.Permits.Get(context.Background(), "pmt_abc")
	te, ok := err.(*ThrottledError)
	if !ok {
		t.Fatalf("expected ThrottledError, got %T", err)
	}
	if te.RetryAfterSeconds != 15 {
		t.Errorf("expected 15 from body fallback, got %d", te.RetryAfterSeconds)
	}
}

func TestThrottled429RetryThenExhaust(t *testing.T) {
	attempts := 0
	rc := &RetryConfig{
		MaxRetries:           1,
		InitialDelay:         time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		BackoffMultiplier:    1,
		RetryableStatusCodes: map[int]bool{429: true},
	}

	_, err := retryWithBackoff(context.Background(), rc, func() (*httpResponse, error) {
		attempts++
		return &httpResponse{
			statusCode: 429,
			body:       []byte(`{"permit":{"decision":"throttled","reason_code":"budget.rate_limit_throttled","outcome_detail":{"retry_after_seconds":1}}}`),
			retryAfter: time.Millisecond,
		}, nil
	})

	// After exhaustion, retryWithBackoff returns the last response (status 429)
	// The caller (doWithRetry) then converts it to ThrottledError
	if err != nil {
		t.Fatalf("expected nil err from retryWithBackoff (returns resp), got %v", err)
	}
	if attempts != 2 { // initial + 1 retry
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func Test403NoRetry(t *testing.T) {
	attempts := 0
	rc := &RetryConfig{
		MaxRetries:           2,
		InitialDelay:         time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		BackoffMultiplier:    1,
		RetryableStatusCodes: map[int]bool{429: true},
	}

	resp, err := retryWithBackoff(context.Background(), rc, func() (*httpResponse, error) {
		attempts++
		return &httpResponse{statusCode: 403, body: []byte(`{"error":{"code":"denied"}}`)}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 403 {
		t.Errorf("expected 403, got %d", resp.statusCode)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry for 403), got %d", attempts)
	}
}

func TestPermitResponseShapeDFields(t *testing.T) {
	data := `{
		"permit_id": "pmt_123",
		"project_id": "proj_1",
		"decision": "deny",
		"reason": "Budget exceeded",
		"reason_code": "budget.daily_cap_exceeded",
		"reason_detail": {"category": "budget", "kind": "daily_cap", "outcome": "deny"},
		"outcome_detail": {"cap": 3000000, "current_spend": 3100000},
		"message": "Daily budget cap exceeded.",
		"budgets": {"schema_version": 1, "currency_unit": "usd_micros", "daily": {"current_spend": 3100000, "cap": 3000000}},
		"constraints": {"schema_version": 1},
		"actions": []
	}`

	var resp PermitResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if resp.Decision != DecisionDeny {
		t.Errorf("expected deny, got %s", resp.Decision)
	}
	if resp.ReasonCode == nil || *resp.ReasonCode != "budget.daily_cap_exceeded" {
		t.Errorf("expected reason_code budget.daily_cap_exceeded, got %v", resp.ReasonCode)
	}
	if resp.ReasonDetail == nil || resp.ReasonDetail["category"] != "budget" {
		t.Errorf("unexpected reason_detail: %v", resp.ReasonDetail)
	}
	if resp.OutcomeDetail == nil {
		t.Error("expected outcome_detail")
	}
	if resp.Message == nil || *resp.Message != "Daily budget cap exceeded." {
		t.Errorf("unexpected message: %v", resp.Message)
	}
	// budgets_json remains opaque map — verify schema_version passes through
	if resp.Budgets["schema_version"] != float64(1) {
		t.Errorf("expected budgets schema_version 1, got %v", resp.Budgets["schema_version"])
	}
}

func TestProxyStream(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/proxy/openai" {
			t.Errorf("expected /v1/proxy/openai, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\"}\n\n")
	})

	events, errc := c.Proxy.OpenAIStream(context.Background(), map[string]any{"model": "gpt-4", "stream": true})
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

func TestAnthropicProxyStream(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/proxy/anthropic" {
			t.Errorf("expected /v1/proxy/anthropic, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\"}\n\n")
	})

	events, errc := c.Proxy.AnthropicStream(context.Background(), map[string]any{"model": "claude-3", "stream": true})
	var received []SSEEvent
	for range events {
		received = append(received, SSEEvent{})
	}
	if err := <-errc; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
}
