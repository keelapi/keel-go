// Code generated from the scoped Permit lifecycle OpenAPI surface; DO NOT EDIT.
package keelclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// PermitLifecycleSummary is the current state and bounded authority shown at
// the top of a Permit lifecycle.
type PermitLifecycleSummary struct {
	ActionName           string                 `json:"action_name"`
	Decision             string                 `json:"decision"`
	ExpiresAt            *time.Time             `json:"expires_at,omitempty"`
	ExpiryLimitingSource *string                `json:"expiry_limiting_source,omitempty"`
	IssuedAt             time.Time              `json:"issued_at"`
	PermitId             openapi_types.UUID     `json:"permit_id"`
	ProjectId            openapi_types.UUID     `json:"project_id"`
	RevocationReasonCode *string                `json:"revocation_reason_code,omitempty"`
	RevokedAt            *time.Time             `json:"revoked_at,omitempty"`
	RevokedByActorId     *string                `json:"revoked_by_actor_id,omitempty"`
	RevokedByActorKind   *string                `json:"revoked_by_actor_kind,omitempty"`
	Status               string                 `json:"status"`
	Target               map[string]interface{} `json:"target"`
}

// PermitLifecycleResponse is the Permit-centric lifecycle reconstructed from
// persisted evidence.
type PermitLifecycleResponse struct {
	Events           []AppSchemasRequestTimelineRequestTimelineEvent `json:"events"`
	GeneratedAt      time.Time                                       `json:"generated_at"`
	Permit           PermitLifecycleSummary                          `json:"permit"`
	RelatedPermitIds []openapi_types.UUID                            `json:"related_permit_ids"`
	RequestId        *string                                         `json:"request_id,omitempty"`
}

// GetPermitLifecycleV1PermitsPermitIdTimelineGetParams defines optional
// compatibility authentication parameters for the Permit lifecycle request.
type GetPermitLifecycleV1PermitsPermitIdTimelineGetParams struct {
	XAPIKey *string `json:"X-API-Key,omitempty"`
}

// GetPermitLifecycleV1PermitsPermitIdTimelineGet performs the public
// project-scoped Permit lifecycle request.
func (c *KeelHTTPClient) GetPermitLifecycleV1PermitsPermitIdTimelineGet(
	ctx context.Context,
	permitId string,
	params *GetPermitLifecycleV1PermitsPermitIdTimelineGetParams,
	reqEditors ...RequestEditorFn,
) (*http.Response, error) {
	req, err := NewGetPermitLifecycleV1PermitsPermitIdTimelineGetRequest(c.Server, permitId, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetPermitLifecycleV1PermitsPermitIdTimelineGetRequest builds the public
// Permit lifecycle request.
func NewGetPermitLifecycleV1PermitsPermitIdTimelineGetRequest(
	server string,
	permitId string,
	params *GetPermitLifecycleV1PermitsPermitIdTimelineGetParams,
) (*http.Request, error) {
	pathParam, err := runtime.StyleParamWithOptions(
		"simple",
		false,
		"permit_id",
		permitId,
		runtime.StyleParamOptions{
			ParamLocation: runtime.ParamLocationPath,
			Type:          "string",
		},
	)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	operationPath := fmt.Sprintf("/v1/permits/%s/timeline", pathParam)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}
	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, queryURL.String(), nil)
	if err != nil {
		return nil, err
	}
	if params != nil && params.XAPIKey != nil {
		req.Header.Set("X-API-Key", *params.XAPIKey)
	}
	return req, nil
}

// GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse is the typed lifecycle
// response envelope.
type GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *PermitLifecycleResponse
	JSON422      *HTTPValidationError
}

// Status returns HTTPResponse.Status.
func (r GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode.
func (r GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ContentType returns the response Content-Type.
func (r GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse) ContentType() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Header.Get("Content-Type")
	}
	return ""
}

// GetPermitLifecycleV1PermitsPermitIdTimelineGetWithResponse returns a parsed
// Permit lifecycle response.
func (c *ClientWithResponses) GetPermitLifecycleV1PermitsPermitIdTimelineGetWithResponse(
	ctx context.Context,
	permitId string,
	params *GetPermitLifecycleV1PermitsPermitIdTimelineGetParams,
	reqEditors ...RequestEditorFn,
) (*GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse, error) {
	rsp, err := c.GetPermitLifecycleV1PermitsPermitIdTimelineGet(
		ctx,
		permitId,
		params,
		reqEditors...,
	)
	if err != nil {
		return nil, err
	}
	return ParseGetPermitLifecycleV1PermitsPermitIdTimelineGetResponse(rsp)
}

// ParseGetPermitLifecycleV1PermitsPermitIdTimelineGetResponse parses the
// lifecycle response body.
func ParseGetPermitLifecycleV1PermitsPermitIdTimelineGetResponse(
	rsp *http.Response,
) (*GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetPermitLifecycleV1PermitsPermitIdTimelineGetResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}
	if strings.Contains(rsp.Header.Get("Content-Type"), "json") {
		switch rsp.StatusCode {
		case http.StatusOK:
			var dest PermitLifecycleResponse
			if err := json.Unmarshal(bodyBytes, &dest); err != nil {
				return nil, err
			}
			response.JSON200 = &dest
		case http.StatusUnprocessableEntity:
			var dest HTTPValidationError
			if err := json.Unmarshal(bodyBytes, &dest); err != nil {
				return nil, err
			}
			response.JSON422 = &dest
		}
	}
	return response, nil
}
