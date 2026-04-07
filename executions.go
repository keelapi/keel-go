package keel

import (
	"context"
	"encoding/json"
	"fmt"
)

// ExecutionsClient provides access to the executions API.
type ExecutionsClient struct {
	t *httpTransport
}

// Create runs a synchronous execution.
func (c *ExecutionsClient) Create(ctx context.Context, req ExecutionCreateRequest) (*ExecutionResponse, error) {
	req.Mode = ModeSync
	body, err := c.t.post(ctx, "/v1/executions", req, nil)
	if err != nil {
		return nil, err
	}
	var resp ExecutionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// Stream runs a streaming execution and returns channels for events and errors.
func (c *ExecutionsClient) Stream(ctx context.Context, req ExecutionCreateRequest) (<-chan ExecutionStreamEvent, <-chan error) {
	req.Mode = ModeStream
	events := make(chan ExecutionStreamEvent)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)

		rc, err := c.t.postStream(ctx, "/v1/executions", req, nil)
		if err != nil {
			errc <- err
			return
		}

		sseEvents, sseErrc := parseSSEStream(rc)
		for sse := range sseEvents {
			events <- ExecutionStreamEvent{
				EventType: sse.EventType,
				Data:      sse.Data,
			}
		}
		if err := <-sseErrc; err != nil {
			errc <- err
		}
	}()

	return events, errc
}

// Get retrieves the timeline for an execution.
func (c *ExecutionsClient) Get(ctx context.Context, requestID string) (*ExecutionTimelineResponse, error) {
	body, err := c.t.get(ctx, "/v1/executions/"+requestID, nil)
	if err != nil {
		return nil, err
	}
	var resp ExecutionTimelineResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
