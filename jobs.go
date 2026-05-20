package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// JobsClient provides access to the batch jobs API.
type JobsClient struct {
	t *httpTransport
}

// Create submits a new batch job.
func (c *JobsClient) Create(ctx context.Context, req JobSubmitRequest) (*JobCreateResponse, error) {
	body, err := c.t.post(ctx, "/v1/jobs", req, nil)
	if err != nil {
		return nil, err
	}
	var resp JobCreateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Get retrieves the status of a batch job.
func (c *JobsClient) Get(ctx context.Context, jobID string) (*JobStatusResponse, error) {
	body, err := c.t.get(ctx, "/v1/jobs/"+url.PathEscape(jobID), nil)
	if err != nil {
		return nil, err
	}
	var resp JobStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
