package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// RequestsClient provides access to the request timeline API.
type RequestsClient struct {
	t *httpTransport
}

// Timeline retrieves the timeline of events for a request.
func (c *RequestsClient) Timeline(ctx context.Context, requestID string) (*RequestTimelineResponse, error) {
	body, err := c.t.get(ctx, "/v1/requests/"+url.PathEscape(requestID)+"/timeline", nil)
	if err != nil {
		return nil, err
	}
	var resp RequestTimelineResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
