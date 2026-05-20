package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ComplianceClient provides access to signed compliance export artifacts.
type ComplianceClient struct {
	t *httpTransport
}

type ComplianceExportVantaWebhook struct {
	CallbackURL     string `json:"callback_url"`
	IncludeManifest bool   `json:"include_manifest,omitempty"`
}

type ComplianceExportCreateRequest struct {
	ExportType   string                        `json:"export_type"`
	Format       string                        `json:"format,omitempty"`
	CompressGzip bool                          `json:"compress_gzip,omitempty"`
	Filters      map[string]any                `json:"filters,omitempty"`
	VantaWebhook *ComplianceExportVantaWebhook `json:"vanta_webhook,omitempty"`
}

type ComplianceExportResponse struct {
	ExportID     string         `json:"export_id"`
	ProjectID    string         `json:"project_id"`
	ExportType   string         `json:"export_type"`
	Status       string         `json:"status"`
	Filters      map[string]any `json:"filters"`
	FilePath     *string        `json:"file_path,omitempty"`
	CreatedAt    string         `json:"created_at"`
	StartedAt    *string        `json:"started_at,omitempty"`
	CompletedAt  *string        `json:"completed_at,omitempty"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	Manifest     map[string]any `json:"manifest,omitempty"`
}

type ComplianceExportListResponse struct {
	Items []ComplianceExportResponse `json:"items"`
}

// CreateExport starts a signed compliance export job.
func (c *ComplianceClient) CreateExport(ctx context.Context, req ComplianceExportCreateRequest, includeChainEntries bool) (*ComplianceExportResponse, error) {
	path := "/v1/compliance/exports"
	if includeChainEntries {
		q := url.Values{}
		q.Set("include_chain_entries", strconv.FormatBool(includeChainEntries))
		path += "?" + q.Encode()
	}
	body, err := c.t.post(ctx, path, req, nil)
	if err != nil {
		return nil, err
	}
	var resp ComplianceExportResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// ListExports lists compliance export jobs for the authenticated project.
func (c *ComplianceClient) ListExports(ctx context.Context) (*ComplianceExportListResponse, error) {
	body, err := c.t.get(ctx, "/v1/compliance/exports", nil)
	if err != nil {
		return nil, err
	}
	var resp ComplianceExportListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// GetExport retrieves a compliance export job.
func (c *ComplianceClient) GetExport(ctx context.Context, exportID string) (*ComplianceExportResponse, error) {
	body, err := c.t.get(ctx, "/v1/compliance/exports/"+url.PathEscape(exportID), nil)
	if err != nil {
		return nil, err
	}
	var resp ComplianceExportResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Keys returns the public key manifest used by verifier workflows.
func (c *ComplianceClient) Keys(ctx context.Context) (map[string]any, error) {
	body, err := c.t.get(ctx, "/v1/compliance/keys", nil)
	if err != nil {
		return nil, err
	}
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return resp, nil
}
