package keelclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type permitLifecycleDoer func(*http.Request) (*http.Response, error)

func (do permitLifecycleDoer) Do(request *http.Request) (*http.Response, error) {
	return do(request)
}

func TestGetPermitLifecycleWithResponse(t *testing.T) {
	doer := permitLifecycleDoer(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/v1/permits/permit_123/timeline" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "keel_sk_test" {
			t.Fatalf("X-API-Key = %q", r.Header.Get("X-API-Key"))
		}
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
			"generated_at":"2026-07-29T23:00:00Z",
			"permit":{
				"permit_id":"2b951e04-53b1-4a5b-a154-4965991021b8",
				"project_id":"aef0df9a-6551-4f11-8274-c074ba1bc260",
				"decision":"allow",
				"status":"active",
				"action_name":"payment.execute",
				"issued_at":"2026-07-29T22:55:00Z",
				"expires_at":"2026-07-29T23:05:00Z",
				"expiry_limiting_source":"policy",
				"target":{"resource_id":"endpoint_123","binding":"request_bound"}
			},
			"request_id":"request_123",
			"related_permit_ids":["2b951e04-53b1-4a5b-a154-4965991021b8"],
			"events":[]
		}`)),
		}, nil
	})

	client, err := NewClientWithResponses("https://api.keel.test", WithHTTPClient(doer))
	if err != nil {
		t.Fatal(err)
	}
	apiKey := "keel_sk_test"
	response, err := client.GetPermitLifecycleV1PermitsPermitIdTimelineGetWithResponse(
		context.Background(),
		"permit_123",
		&GetPermitLifecycleV1PermitsPermitIdTimelineGetParams{XAPIKey: &apiKey},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil {
		t.Fatal("expected JSON200")
	}
	if response.JSON200.Permit.Status != "active" {
		t.Fatalf("status = %q", response.JSON200.Permit.Status)
	}
	if response.JSON200.Permit.ExpiryLimitingSource == nil ||
		*response.JSON200.Permit.ExpiryLimitingSource != "policy" {
		t.Fatalf("expiry limiting source = %#v", response.JSON200.Permit.ExpiryLimitingSource)
	}
}
