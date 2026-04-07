package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// PermitsClient provides access to the permits API.
type PermitsClient struct {
	t *httpTransport
}

// Create creates a new permit.
func (c *PermitsClient) Create(ctx context.Context, req PermitRequest) (*PermitResponse, error) {
	body, err := c.t.post(ctx, "/v1/permits", req, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// DryRun evaluates a permit without creating it.
func (c *PermitsClient) DryRun(ctx context.Context, req PermitRequest) (*PermitDryRunResponse, error) {
	body, err := c.t.post(ctx, "/v1/permits/dry-run", req, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitDryRunResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// List returns a paginated list of permits using cursor pagination.
func (c *PermitsClient) List(ctx context.Context, params PermitListParams) (*PermitsListResponse, error) {
	q := url.Values{}
	if params.Cursor != nil {
		q.Set("cursor", *params.Cursor)
	}
	if params.Limit != nil {
		q.Set("limit", strconv.Itoa(*params.Limit))
	}
	if params.ProjectID != nil {
		q.Set("project_id", *params.ProjectID)
	}
	if params.Decision != nil {
		q.Set("decision", *params.Decision)
	}

	path := "/v1/permits"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.t.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Get retrieves a single permit by ID.
func (c *PermitsClient) Get(ctx context.Context, permitID string) (*PermitAuditItem, error) {
	body, err := c.t.get(ctx, "/v1/permits/"+permitID, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitAuditItem
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Export exports permits in the given format.
func (c *PermitsClient) Export(ctx context.Context, from, to string, format string) (*PermitExportResponse, error) {
	q := url.Values{}
	q.Set("from", from)
	q.Set("to", to)
	q.Set("format", format)

	body, err := c.t.get(ctx, "/v1/permits/export?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var resp PermitExportResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// ReportUsage reports token/cost usage against a permit.
func (c *PermitsClient) ReportUsage(ctx context.Context, permitID string, req PermitUsageReportRequest) (*PermitUsageReportResponse, error) {
	body, err := c.t.post(ctx, "/v1/permits/"+permitID+"/usage", req, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitUsageReportResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Attest records an attestation on a permit.
func (c *PermitsClient) Attest(ctx context.Context, permitID string, req PermitAttestationRequest) (*PermitResponse, error) {
	body, err := c.t.post(ctx, "/v1/permits/"+permitID+"/attest", req, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// AddEvidence adds evidence to a permit.
func (c *PermitsClient) AddEvidence(ctx context.Context, permitID string, req PermitEvidenceCreateRequest) (*PermitEvidenceOut, error) {
	body, err := c.t.post(ctx, "/v1/permits/"+permitID+"/evidence", req, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitEvidenceOut
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// ListEvidence lists all evidence attached to a permit.
func (c *PermitsClient) ListEvidence(ctx context.Context, permitID string) (*PermitEvidenceListResponse, error) {
	body, err := c.t.get(ctx, "/v1/permits/"+permitID+"/evidence", nil)
	if err != nil {
		return nil, err
	}
	var resp PermitEvidenceListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Lineage returns the lineage graph for a permit.
func (c *PermitsClient) Lineage(ctx context.Context, permitID string) (*PermitLineageResponse, error) {
	body, err := c.t.get(ctx, "/v1/permits/"+permitID+"/lineage", nil)
	if err != nil {
		return nil, err
	}
	var resp PermitLineageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Bundle returns the full audit bundle for a permit.
func (c *PermitsClient) Bundle(ctx context.Context, permitID string) (*PermitAuditBundle, error) {
	body, err := c.t.get(ctx, "/v1/permits/"+permitID+"/bundle", nil)
	if err != nil {
		return nil, err
	}
	var resp PermitAuditBundle
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
