package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type workflowIDContextKey struct{}

// WorkflowsClient provides access to the workflow intent API.
type WorkflowsClient struct {
	t *httpTransport
}

// WithWorkflow returns a child context carrying the workflow ID.
func WithWorkflow(ctx context.Context, workflowID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, workflowIDContextKey{}, strings.TrimSpace(workflowID))
}

// WorkflowFromContext returns the workflow ID carried by ctx, if present.
func WorkflowFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	workflowID, ok := ctx.Value(workflowIDContextKey{}).(string)
	if !ok {
		return "", false
	}
	workflowID = strings.TrimSpace(workflowID)
	if workflowID == "" {
		return "", false
	}
	return workflowID, true
}

// RunInWorkflow invokes fn with a child context carrying the workflow ID.
func RunInWorkflow(ctx context.Context, workflowID string, fn func(context.Context) error) error {
	if fn == nil {
		return fmt.Errorf("keel: workflow function is nil")
	}
	return fn(WithWorkflow(ctx, workflowID))
}

// Declare declares a workflow intent before the workflow runs.
func (c *WorkflowsClient) Declare(ctx context.Context, req WorkflowDeclareRequest) (*WorkflowDeclarationResponse, error) {
	body, err := c.t.post(ctx, "/v1/workflows", req, nil)
	if err != nil {
		return nil, err
	}
	var resp WorkflowDeclarationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Amend appends an amendment to a workflow declaration.
func (c *WorkflowsClient) Amend(ctx context.Context, workflowID string, req WorkflowAmendRequest) (*WorkflowDeclarationResponse, error) {
	path := "/v1/workflows/" + url.PathEscape(workflowID) + "/amend"
	body, err := c.t.post(ctx, path, req, nil)
	if err != nil {
		return nil, err
	}
	var resp WorkflowDeclarationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Complete marks a workflow as complete and reconciles its final call count.
func (c *WorkflowsClient) Complete(ctx context.Context, workflowID string) (*WorkflowCompleteResponse, error) {
	path := "/v1/workflows/" + url.PathEscape(workflowID) + "/complete"
	body, err := c.t.post(ctx, path, struct{}{}, nil)
	if err != nil {
		return nil, err
	}
	var resp WorkflowCompleteResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Get retrieves a workflow declaration by caller-supplied workflow ID.
func (c *WorkflowsClient) Get(ctx context.Context, workflowID string) (*WorkflowDeclarationResponse, error) {
	path := "/v1/workflows/" + url.PathEscape(workflowID)
	body, err := c.t.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var resp WorkflowDeclarationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// List returns workflow declarations using offset pagination.
func (c *WorkflowsClient) List(ctx context.Context, params WorkflowListParams) (*WorkflowsListResponse, error) {
	q := url.Values{}
	if params.Status != nil {
		q.Set("status", *params.Status)
	}
	if params.CreatedAtFrom != nil {
		q.Set("created_at_from", params.CreatedAtFrom.Format(time.RFC3339))
	}
	if params.CreatedAtTo != nil {
		q.Set("created_at_to", params.CreatedAtTo.Format(time.RFC3339))
	}
	if params.Limit != nil {
		q.Set("limit", strconv.Itoa(*params.Limit))
	}
	if params.Offset != nil {
		q.Set("offset", strconv.Itoa(*params.Offset))
	}

	path := "/v1/workflows"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.t.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var resp WorkflowsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
