package keel

import "time"

// WorkflowDecision is the declaration decision returned by the workflow API.
type WorkflowDecision string

const (
	WorkflowDecisionAccepted WorkflowDecision = "accepted"
	WorkflowDecisionRejected WorkflowDecision = "rejected"
)

// WorkflowStatus is the lifecycle status of a workflow declaration.
type WorkflowStatus string

const (
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusExpired   WorkflowStatus = "expired"
	WorkflowStatusRejected  WorkflowStatus = "rejected"
)

// WorkflowIntent describes the caller-declared expected workflow shape.
type WorkflowIntent struct {
	ExpectedCalls               *int    `json:"expected_calls,omitempty"`
	MaxCalls                    *int    `json:"max_calls,omitempty"`
	ExpectedModel               *string `json:"expected_model,omitempty"`
	ExpectedInputTokensPerCall  *int    `json:"expected_input_tokens_per_call,omitempty"`
	ExpectedOutputTokensPerCall *int    `json:"expected_output_tokens_per_call,omitempty"`
	MaxDurationSeconds          *int    `json:"max_duration_seconds,omitempty"`
}

// WorkflowDeclareRequest is the request body for declaring a workflow.
type WorkflowDeclareRequest struct {
	WorkflowID       string         `json:"workflow_id"`
	Intent           WorkflowIntent `json:"intent"`
	BudgetEnvelopeID *string        `json:"budget_envelope_id,omitempty"`
}

// ProjectedCostMethodology records how projected workflow cost was computed.
type ProjectedCostMethodology struct {
	Basis                        string  `json:"basis,omitempty"`
	Provenance                   string  `json:"provenance,omitempty"`
	ExpectedCalls                *int    `json:"expected_calls,omitempty"`
	InputTokensPerCallEstimated  *int    `json:"input_tokens_per_call_estimated,omitempty"`
	OutputTokensPerCallEstimated *int    `json:"output_tokens_per_call_estimated,omitempty"`
	PricingTableID               *string `json:"pricing_table_id,omitempty"`
	Tokenizer                    *string `json:"tokenizer,omitempty"`
	Quality                      *string `json:"quality,omitempty"`
}

// ProjectedCost is the projected total workflow cost.
type ProjectedCost struct {
	AmountMicros *int64                    `json:"amount_micros,omitempty"`
	Currency     string                    `json:"currency,omitempty"`
	Methodology  *ProjectedCostMethodology `json:"methodology,omitempty"`
}

// WorkflowPrincipal identifies who declared a workflow.
type WorkflowPrincipal struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// WorkflowDeclaredVia identifies the SDK metadata claimed by the declarer.
type WorkflowDeclaredVia struct {
	SDK        *string `json:"sdk,omitempty"`
	SDKVersion *string `json:"sdk_version,omitempty"`
}

// WorkflowAmendRequest is the request body for amending a workflow declaration.
type WorkflowAmendRequest struct {
	IfMatchVersion   int     `json:"if_match_version"`
	NewMaxCalls      *int    `json:"new_max_calls,omitempty"`
	NewExpectedCalls *int    `json:"new_expected_calls,omitempty"`
	ReasonProvided   *string `json:"reason_provided,omitempty"`
}

// WorkflowAmendmentResponse is a signed workflow amendment record.
type WorkflowAmendmentResponse struct {
	ID                     string    `json:"id"`
	WorkflowDeclarationID  string    `json:"workflow_declaration_id"`
	AppliedAgainstVersion  int       `json:"applied_against_version"`
	PreviousMaxCalls       *int      `json:"previous_max_calls,omitempty"`
	NewMaxCalls            *int      `json:"new_max_calls,omitempty"`
	PreviousExpectedCalls  *int      `json:"previous_expected_calls,omitempty"`
	NewExpectedCalls       *int      `json:"new_expected_calls,omitempty"`
	ReasonProvided         *string   `json:"reason_provided,omitempty"`
	AmendmentCanonicalHash *string   `json:"amendment_canonical_hash,omitempty"`
	AmendmentSignatureB64  *string   `json:"amendment_signature_b64,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}

// WorkflowDeclarationResponse is returned by declare, amend, get, and list APIs.
type WorkflowDeclarationResponse struct {
	WorkflowID               string                      `json:"workflow_id"`
	Decision                 WorkflowDecision            `json:"decision"`
	Status                   *WorkflowStatus             `json:"status,omitempty"`
	ID                       *string                     `json:"id,omitempty"`
	Version                  *int                        `json:"version,omitempty"`
	ExpectedCalls            *int                        `json:"expected_calls,omitempty"`
	MaxCalls                 *int                        `json:"max_calls,omitempty"`
	CachedActualCalls        *int                        `json:"cached_actual_calls,omitempty"`
	ProjectedCost            *ProjectedCost              `json:"projected_cost,omitempty"`
	ReasonCode               *string                     `json:"reason_code,omitempty"`
	DecisionDetails          map[string]any              `json:"decision_details,omitempty"`
	DeclaredBy               *WorkflowPrincipal          `json:"declared_by,omitempty"`
	DeclaredVia              *WorkflowDeclaredVia        `json:"declared_via,omitempty"`
	DeclarationCanonicalHash *string                     `json:"declaration_canonical_hash,omitempty"`
	DeclarationSignatureB64  *string                     `json:"declaration_signature_b64,omitempty"`
	DeclaredAt               *time.Time                  `json:"declared_at,omitempty"`
	ExpiresAt                *time.Time                  `json:"expires_at,omitempty"`
	CompletedAt              *time.Time                  `json:"completed_at,omitempty"`
	Amendments               []WorkflowAmendmentResponse `json:"amendments,omitempty"`
}

// WorkflowCompleteResponse is returned after explicitly completing a workflow.
type WorkflowCompleteResponse struct {
	WorkflowID               string         `json:"workflow_id"`
	Status                   WorkflowStatus `json:"status"`
	CachedActualCalls        int            `json:"cached_actual_calls"`
	AuthoritativeActualCalls int            `json:"authoritative_actual_calls"`
	CounterReconciled        bool           `json:"counter_reconciled"`
	CompletedAt              time.Time      `json:"completed_at"`
}

// WorkflowListParams holds query parameters for listing workflow declarations.
type WorkflowListParams struct {
	Status        *string    `json:"status,omitempty"`
	CreatedAtFrom *time.Time `json:"created_at_from,omitempty"`
	CreatedAtTo   *time.Time `json:"created_at_to,omitempty"`
	Limit         *int       `json:"limit,omitempty"`
	Offset        *int       `json:"offset,omitempty"`
}

// WorkflowsListResponse is the paginated list of workflow declarations.
type WorkflowsListResponse struct {
	Items  []WorkflowDeclarationResponse `json:"items"`
	Limit  int                           `json:"limit"`
	Offset int                           `json:"offset"`
	Total  int                           `json:"total"`
}
