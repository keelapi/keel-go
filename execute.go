package keel

import (
	"context"
	"encoding/json"
	"fmt"
)

// ExecuteClient provides access to the combined execute endpoint.
type ExecuteClient struct {
	t *httpTransport
}

// Run creates a permit and executes in a single call.
func (c *ExecuteClient) Run(ctx context.Context, req ExecuteRequest) (*ExecutionResponse, error) {
	body, err := c.t.post(ctx, "/v1/execute", req, nil)
	if err != nil {
		if envelope, ok := executionResponseFromError(err); ok {
			return envelope, nil
		}
		return nil, err
	}
	var resp ExecutionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
