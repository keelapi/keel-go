package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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

// List returns a paginated list of permits using the API's before cursor.
func (c *PermitsClient) List(ctx context.Context, params PermitListParams) (*PermitsListResponse, error) {
	q := url.Values{}
	if params.Before != nil {
		q.Set("before", *params.Before)
	} else if params.Cursor != nil {
		q.Set("before", *params.Cursor)
	}
	if params.Limit != nil {
		q.Set("limit", strconv.Itoa(*params.Limit))
	}
	if params.ProjectID != nil {
		q.Set("project_id", *params.ProjectID)
	}
	if params.View != nil {
		q.Set("view", *params.View)
	}
	if params.Type != nil {
		q.Set("type", *params.Type)
	}
	if params.SubjectID != nil {
		q.Set("subject_id", *params.SubjectID)
	}
	if params.SubjectType != nil {
		q.Set("subject_type", *params.SubjectType)
	}
	if params.ActionName != nil {
		q.Set("action_name", *params.ActionName)
	}
	if params.PolicyID != nil {
		q.Set("policy_id", *params.PolicyID)
	}
	if params.StartDate != nil {
		q.Set("start_date", *params.StartDate)
	}
	if params.EndDate != nil {
		q.Set("end_date", *params.EndDate)
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
	body, err := c.t.get(ctx, "/v1/permits/"+url.PathEscape(permitID), nil)
	if err != nil {
		return nil, err
	}
	var resp PermitAuditItem
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Export exports permits in the requested format. JSON responses are returned
// in Data as raw JSON so older callers keep working with the current bundle API.
func (c *PermitsClient) Export(ctx context.Context, from, to string, format string) (*PermitExportResponse, error) {
	q := url.Values{}
	q.Set("from", from)
	q.Set("to", to)
	q.Set("format", format)

	body, err := c.t.get(ctx, "/v1/permits/export?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp := PermitExportResponse{
		Data:   string(body),
		Format: format,
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		if data, ok := payload["data"].(string); ok {
			resp.Data = data
		}
		if responseFormat, ok := payload["format"].(string); ok {
			resp.Format = responseFormat
		}
		if count, ok := payload["count"].(float64); ok {
			resp.Count = int(count)
		} else if count, ok := payload["record_count"].(float64); ok {
			resp.Count = int(count)
		}
	}
	return &resp, nil
}

// ExportBundle exports permits as the canonical JSON audit export bundle.
func (c *PermitsClient) ExportBundle(ctx context.Context, from, to string) (*AuditExportBundle, error) {
	q := url.Values{}
	q.Set("from", from)
	q.Set("to", to)
	q.Set("format", string(PermitExportJSON))

	body, err := c.t.get(ctx, "/v1/permits/export?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var resp AuditExportBundle
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// ExportCSV exports permits as CSV text.
func (c *PermitsClient) ExportCSV(ctx context.Context, from, to string) (string, error) {
	q := url.Values{}
	q.Set("from", from)
	q.Set("to", to)
	q.Set("format", string(PermitExportCSV))

	resp, err := c.Export(ctx, from, to, string(PermitExportCSV))
	if err != nil {
		return "", err
	}
	return resp.Data, nil
}

// ReportUsage reports token/cost usage against a permit.
func (c *PermitsClient) ReportUsage(ctx context.Context, permitID string, req PermitUsageReportRequest) (*PermitUsageReportResponse, error) {
	body, err := c.t.post(ctx, "/v1/permits/"+url.PathEscape(permitID)+"/usage", req, nil)
	if err != nil {
		return nil, err
	}
	var resp PermitUsageReportResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// VerifyUsage verifies previously reported usage for a permit.
func (c *PermitsClient) VerifyUsage(ctx context.Context, permitID string, req PermitUsageVerificationInput) (map[string]any, error) {
	body, err := c.t.post(ctx, "/v1/permits/"+url.PathEscape(permitID)+"/usage/verify", req, nil)
	if err != nil {
		return nil, err
	}
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return resp, nil
}

// Attest records an attestation on a permit.
func (c *PermitsClient) Attest(ctx context.Context, permitID string, req PermitAttestationRequest) (*PermitResponse, error) {
	body, err := c.t.post(ctx, "/v1/permits/"+url.PathEscape(permitID)+"/attest", req, nil)
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
	body, err := c.t.post(ctx, "/v1/permits/"+url.PathEscape(permitID)+"/evidence", req, nil)
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
	body, err := c.t.get(ctx, "/v1/permits/"+url.PathEscape(permitID)+"/evidence", nil)
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
	body, err := c.t.get(ctx, "/v1/permits/"+url.PathEscape(permitID)+"/lineage", nil)
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
	body, err := c.t.get(ctx, "/v1/permits/"+url.PathEscape(strings.TrimSpace(permitID))+"/bundle", nil)
	if err != nil {
		return nil, err
	}
	var resp PermitAuditBundle
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
