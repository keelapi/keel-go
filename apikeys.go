package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ApiKeysClient provides access to the API keys management API.
type ApiKeysClient struct {
	t *httpTransport
}

// Create creates a new API key.
func (c *ApiKeysClient) Create(ctx context.Context, req *ApiKeyCreateRequest) (*ApiKeyCreateResponse, error) {
	body, err := c.t.post(ctx, "/v1/api-keys", req, nil)
	if err != nil {
		return nil, err
	}
	var resp ApiKeyCreateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// List returns a paginated list of API keys.
func (c *ApiKeysClient) List(ctx context.Context, params ApiKeyListParams) (*ApiKeyListResponse, error) {
	q := url.Values{}
	if params.Limit != nil {
		q.Set("limit", strconv.Itoa(*params.Limit))
	}
	if params.Offset != nil {
		q.Set("offset", strconv.Itoa(*params.Offset))
	}
	if params.Status != nil {
		q.Set("status", *params.Status)
	}

	path := "/v1/api-keys"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.t.get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var resp ApiKeyListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
// Revoke revokes an API key.
func (c *ApiKeysClient) Revoke(ctx context.Context, keyID string) error {
	_, err := c.t.post(ctx, "/v1/api-keys/"+keyID+"/revoke", nil, nil)
	return err
}
