package keel

import (
	"context"
	"encoding/json"
	"fmt"
)

// IntegrityClient provides access to public integrity verification keys.
type IntegrityClient struct {
	t *httpTransport
}

type CheckpointPublicKeyResponse struct {
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	KeyID     string `json:"key_id"`
	Scope     string `json:"scope"`
}

type PermitBindingPublicKey struct {
	KeyID        string  `json:"key_id"`
	PublicKeyB64 string  `json:"public_key_b64"`
	ActiveFrom   *string `json:"active_from,omitempty"`
	ActiveTo     *string `json:"active_to,omitempty"`
}

type PermitBindingPublicKeysResponse struct {
	Purpose string                   `json:"purpose"`
	Keys    []PermitBindingPublicKey `json:"keys"`
}

// CheckpointPublicKey returns the public key used to verify integrity checkpoints.
func (c *IntegrityClient) CheckpointPublicKey(ctx context.Context) (*CheckpointPublicKeyResponse, error) {
	body, err := c.t.get(ctx, "/v1/integrity/checkpoint-public-key", nil)
	if err != nil {
		return nil, err
	}
	var resp CheckpointPublicKeyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}

// PermitBindingPublicKeys returns permit-binding public key history.
func (c *IntegrityClient) PermitBindingPublicKeys(ctx context.Context) (*PermitBindingPublicKeysResponse, error) {
	body, err := c.t.get(ctx, "/v1/integrity/permit-binding-public-keys", nil)
	if err != nil {
		return nil, err
	}
	var resp PermitBindingPublicKeysResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("keel: decode response: %w", err)
	}
	return &resp, nil
}
