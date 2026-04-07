package keel

import (
	"encoding/json"
	"time"
)

// Decision represents the permit decision outcome.
type Decision string

const (
	DecisionAllow     Decision = "allow"
	DecisionDeny      Decision = "deny"
	DecisionChallenge Decision = "challenge"
)

// RoutingProvider represents a supported AI provider.
type RoutingProvider string

const (
	ProviderOpenAI    RoutingProvider = "openai"
	ProviderAnthropic RoutingProvider = "anthropic"
	ProviderGoogle    RoutingProvider = "google"
	ProviderXAI       RoutingProvider = "xai"
	ProviderMeta      RoutingProvider = "meta"
)

// ExecutionMode represents sync or stream execution.
type ExecutionMode string

const (
	ModeSync   ExecutionMode = "sync"
	ModeStream ExecutionMode = "stream"
)

// CapabilityOperation represents the type of AI operation.
type CapabilityOperation string

const (
	OpGenerateText    CapabilityOperation = "generate.text"
	OpEmbedText       CapabilityOperation = "embed.text"
	OpGenerateImage   CapabilityOperation = "generate.image"
	OpEditImage       CapabilityOperation = "edit.image"
	OpUnderstandImage CapabilityOperation = "understand.image"
	OpGenerateAudio   CapabilityOperation = "generate.audio"
	OpTranscribeAudio CapabilityOperation = "transcribe.audio"
	OpGenerateVideo   CapabilityOperation = "generate.video"
	OpUnderstandVideo CapabilityOperation = "understand.video"
	OpRunBatch        CapabilityOperation = "run.batch"
	OpRunAsync        CapabilityOperation = "run.async"
	OpRealtimeSession CapabilityOperation = "realtime.session"
	OpCallTools       CapabilityOperation = "call.tools"
)

// ExecutionOperation is an alias for CapabilityOperation for backward compatibility.
type ExecutionOperation = CapabilityOperation

// JobStatus represents the state of a batch job.
type JobStatus string

const (
	JobSubmitted  JobStatus = "submitted"
	JobQueued     JobStatus = "queued"
	JobProcessing JobStatus = "processing"
	JobCompleted  JobStatus = "completed"
	JobFailed     JobStatus = "failed"
)

// --- Request types ---

// Subject identifies who is making the request.
type Subject struct {
	Type       string         `json:"type"`
	ID         string         `json:"id"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// PermitSubject is an alias for Subject for backward compatibility.
type PermitSubject = Subject

// Action describes the requested action.
type Action struct {
	Name       string         `json:"name"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// PermitAction is an alias for Action for backward compatibility.
type PermitAction = Action

// PermitInput represents an input part in a permit resource.
type PermitInput struct {
	Type    string         `json:"type"`
	Content map[string]any `json:"content,omitempty"`
}

// ResourceAttributes describes the attributes of a resource in a permit request.
type ResourceAttributes struct {
	Provider               string              `json:"provider"`
	Model                  string              `json:"model"`
	Operation              CapabilityOperation `json:"operation"`
	Modality               string              `json:"modality,omitempty"`
	ExecutionMode          ExecutionMode       `json:"execution_mode,omitempty"`
	EstimatedInputTokens   int                 `json:"estimated_input_tokens"`
	EstimatedOutputTokens  int                 `json:"estimated_output_tokens"`
	MaxOutputTokensRequested *int              `json:"max_output_tokens_requested,omitempty"`
	Inputs                 []PermitInput       `json:"inputs,omitempty"`
	Routing                *RoutingPreferences `json:"routing,omitempty"`
	CallbackURL            *string             `json:"callback_url,omitempty"`
}

// Resource describes the target resource.
type Resource struct {
	Type       string             `json:"type"`
	ID         string             `json:"id"`
	Attributes ResourceAttributes `json:"attributes"`
}

// PermitResource is an alias for Resource for backward compatibility.
type PermitResource = Resource

// Context provides additional context for the permit decision.
type Context struct {
	Timestamp *time.Time `json:"timestamp,omitempty"`
	IP        *string    `json:"ip,omitempty"`
	UserAgent *string    `json:"user_agent,omitempty"`
}

// PermitContext is an alias for Context for backward compatibility.
type PermitContext = Context

// PermitConditionRule specifies a single condition rule.
type PermitConditionRule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value,omitempty"`
}

// PermitConditions specifies constraints for the permit.
type PermitConditions struct {
	Match string                `json:"match"` // "all" or "any"
	Rules []PermitConditionRule `json:"rules"`
}

// PermitRequest is the request body for creating a permit.
type PermitRequest struct {
	ProjectID        string            `json:"project_id"`
	ParentPermitID   *string           `json:"parent_permit_id,omitempty"`
	BudgetEnvelopeID *string           `json:"budget_envelope_id,omitempty"`
	IdempotencyKey   string            `json:"idempotency_key"`
	SessionID        *string           `json:"session_id,omitempty"`
	Subject          Subject           `json:"subject"`
	Action           Action            `json:"action"`
	Resource         Resource          `json:"resource"`
	Context          *Context          `json:"context,omitempty"`
	Conditions       *PermitConditions `json:"conditions,omitempty"`
	CustomMetadata   map[string]any    `json:"custom_metadata,omitempty"`
}

// --- Response types ---

// PermitActionResponse is an action returned in a permit response.
type PermitActionResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// PermitMetadataSubject identifies the subject in permit metadata.
type PermitMetadataSubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// PermitMetadataAction identifies the action in permit metadata.
type PermitMetadataAction struct {
	Name string `json:"name"`
}

// PermitDecisionDetails holds additional detail about how a permit decision was made.
type PermitDecisionDetails struct {
	PolicyResults []map[string]any `json:"policy_results,omitempty"`
	Extra         map[string]any   `json:"extra,omitempty"`
}

// PermitMetadata holds metadata about a permit decision.
type PermitMetadata struct {
	PolicyID      string                 `json:"policy_id"`
	PolicyVersion string                 `json:"policy_version"`
	EvaluatedAt   time.Time              `json:"evaluated_at"`
	Subject       *PermitMetadataSubject `json:"subject,omitempty"`
	Action        *PermitMetadataAction  `json:"action,omitempty"`
	SessionID     *string                `json:"session_id,omitempty"`
}

// PermitResponse is the response from creating a permit.
type PermitResponse struct {
	PermitID        string                 `json:"permit_id"`
	ProjectID       string                 `json:"project_id"`
	ParentPermitID  *string                `json:"parent_permit_id,omitempty"`
	DelegationDepth int                    `json:"delegation_depth"`
	Decision        Decision               `json:"decision"`
	Reason          string                 `json:"reason"`
	Status          *string                `json:"status,omitempty"`
	Constraints     map[string]any         `json:"constraints,omitempty"`
	Budgets         map[string]any         `json:"budgets,omitempty"`
	Actions         []PermitActionResponse `json:"actions"`
	DecisionDetails *PermitDecisionDetails `json:"decision_details,omitempty"`
	Routing         *RoutingDecision       `json:"routing,omitempty"`
	Metadata        PermitMetadata         `json:"metadata"`
}

// PermitDryRunResponse is the response from a dry-run permit evaluation.
type PermitDryRunResponse struct {
	Decision Decision       `json:"decision"`
	Reason   string         `json:"reason"`
	Routing  *RoutingDecision `json:"routing,omitempty"`
}

// PermitAuditItem represents a permit in the audit log.
type PermitAuditItem struct {
	PermitID        string                 `json:"permit_id"`
	ProjectID       string                 `json:"project_id"`
	ParentPermitID  *string                `json:"parent_permit_id,omitempty"`
	DelegationDepth int                    `json:"delegation_depth"`
	Decision        Decision               `json:"decision"`
	Reason          string                 `json:"reason"`
	Status          *string                `json:"status,omitempty"`
	Constraints     map[string]any         `json:"constraints,omitempty"`
	Budgets         map[string]any         `json:"budgets,omitempty"`
	Actions         []PermitActionResponse `json:"actions"`
	DecisionDetails *PermitDecisionDetails `json:"decision_details,omitempty"`
	Routing         *RoutingDecision       `json:"routing,omitempty"`
	Metadata        PermitMetadata         `json:"metadata"`
}

// PermitAttestation records attestation details for a used permit.
type PermitAttestation struct {
	Status     string `json:"status"`
	AttestedAt string `json:"attested_at"`
	AttestedBy string `json:"attested_by,omitempty"`
}

// PermitAuditBundle contains a full audit bundle for a permit.
type PermitAuditBundle struct {
	Permit    PermitAuditItem        `json:"permit"`
	Execution *ExecutionResponse     `json:"execution,omitempty"`
	Timeline  []RequestTimelineEvent `json:"timeline,omitempty"`
	Evidence  []PermitEvidenceOut    `json:"evidence,omitempty"`
}

// PermitListParams holds query parameters for listing permits.
type PermitListParams struct {
	Cursor    *string `json:"cursor,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
	ProjectID *string `json:"project_id,omitempty"`
	Decision  *string `json:"decision,omitempty"`
}

// PermitsListResponse is the paginated list of permit audit items.
type PermitsListResponse struct {
	Items      []PermitAuditItem `json:"items"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

// PermitAuditListResponse is an alias for backward compatibility.
type PermitAuditListResponse = PermitsListResponse

// PermitExportResponse is the response from exporting permits.
type PermitExportResponse struct {
	Data   string `json:"data"`
	Format string `json:"format"`
	Count  int    `json:"count"`
}

// PermitAttestationRequest is the request body for attesting a permit.
type PermitAttestationRequest struct {
	Attestor        string  `json:"attestor"`
	AttestationType string  `json:"attestation_type"`
	EvidenceURL     *string `json:"evidence_url,omitempty"`
}

// PermitEvidenceCreateRequest is the request body for adding evidence to a permit.
type PermitEvidenceCreateRequest struct {
	EvidenceType  string `json:"evidence_type"`
	EvidenceValue string `json:"evidence_value"`
	Label         string `json:"label"`
	AttachedBy    string `json:"attached_by"`
}

// PermitEvidenceOut represents a piece of evidence attached to a permit.
type PermitEvidenceOut struct {
	EvidenceID string `json:"evidence_id"`
	PermitID   string `json:"permit_id"`
	Type       string `json:"type"`
	Content    map[string]any `json:"content"`
	CreatedAt  string `json:"created_at"`
}

// PermitEvidenceListResponse is the response from listing evidence.
type PermitEvidenceListResponse struct {
	Items []PermitEvidenceOut `json:"items"`
}

// PermitLineageNode represents a node in a permit's lineage graph.
type PermitLineageNode struct {
	PermitID string   `json:"permit_id"`
	ParentID *string  `json:"parent_id,omitempty"`
	Decision Decision `json:"decision"`
	IssuedAt string   `json:"issued_at"`
}

// PermitLineageResponse is the response from querying permit lineage.
type PermitLineageResponse struct {
	ProjectID       string              `json:"project_id"`
	RootPermitID    string              `json:"root_permit_id"`
	CurrentPermitID string              `json:"current_permit_id"`
	Parent          *PermitLineageNode  `json:"parent,omitempty"`
	Ancestors       []PermitLineageNode `json:"ancestors,omitempty"`
	Current         PermitLineageNode   `json:"current"`
}

// PermitUsageReportRequest is the request body for reporting usage on a permit.
type PermitUsageReportRequest struct {
	Provider            *string          `json:"provider,omitempty"`
	Model               *string          `json:"model,omitempty"`
	ActualInputTokens   *int             `json:"actual_input_tokens,omitempty"`
	ActualOutputTokens  *int             `json:"actual_output_tokens,omitempty"`
	ActualTotalTokens   *int             `json:"actual_total_tokens,omitempty"`
	CostUSD             *string          `json:"cost_usd,omitempty"`
	CostUSDMicros       *int64           `json:"cost_usd_micros,omitempty"`
	UsageIdempotencyKey *string          `json:"usage_idempotency_key,omitempty"`
	NormalizedUsage     map[string]any   `json:"normalized_usage,omitempty"`
	UsageMetrics        []map[string]any `json:"usage_metrics,omitempty"`
	Metadata            map[string]any   `json:"metadata,omitempty"`
}

// PermitUsageReportResponse is the response from reporting usage.
type PermitUsageReportResponse struct {
	PermitID string `json:"permit_id"`
	Status   string `json:"status"`
}

// --- Routing types ---

// RoutingTarget identifies a specific provider and model.
type RoutingTarget struct {
	Provider RoutingProvider `json:"provider"`
	Model    string          `json:"model"`
}

// RoutingPreferences express caller preferences for routing.
type RoutingPreferences struct {
	PreferredProviders []RoutingProvider `json:"preferred_providers,omitempty"`
	PreferredModels    []string          `json:"preferred_models,omitempty"`
	Fallback           *bool             `json:"fallback,omitempty"`
}

// RoutingDecision is the routing result embedded in a permit.
type RoutingDecision struct {
	Provider RoutingProvider `json:"provider"`
	Model    string          `json:"model"`
	Region   *string         `json:"region,omitempty"`
}

// --- Execution types ---

// MessageInput represents a message in an execution request.
type MessageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// InputPart represents an input part in an execution request.
type InputPart struct {
	Type    string         `json:"type"`
	Content map[string]any `json:"content,omitempty"`
}

// RoutingInput represents routing preferences in an execution request.
type RoutingInput struct {
	PreferredProviders []RoutingProvider `json:"preferred_providers,omitempty"`
	PreferredModels    []string          `json:"preferred_models,omitempty"`
	Fallback           *bool             `json:"fallback,omitempty"`
}

// ParametersInput represents parameters for an execution request.
type ParametersInput struct {
	MaxTokens   *int               `json:"max_tokens,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
	Extra       map[string]any     `json:"extra,omitempty"`
}

// ExecutionCreateRequest is the request body for creating an execution.
type ExecutionCreateRequest struct {
	Operation       string                        `json:"operation"`
	Mode            ExecutionMode                 `json:"mode,omitempty"`
	Messages        []MessageInput                `json:"messages,omitempty"`
	Inputs          []InputPart                   `json:"inputs,omitempty"`
	Routing         *RoutingInput                 `json:"routing,omitempty"`
	Parameters      *ParametersInput              `json:"parameters,omitempty"`
	ProviderOptions map[string]map[string]any     `json:"provider_options,omitempty"`
}

// ExecutionMessage is an alias for MessageInput for backward compatibility.
type ExecutionMessage = MessageInput

// ExecutionResponse is the response from a synchronous execution.
type ExecutionResponse struct {
	RequestID  string                 `json:"request_id"`
	PermitID   string                 `json:"permit_id"`
	Status     string                 `json:"status"`
	Provider   RoutingProvider        `json:"provider"`
	Model      string                 `json:"model"`
	Messages   []MessageInput         `json:"messages,omitempty"`
	Usage      *ExecutionUsage        `json:"usage,omitempty"`
	Timing     *ExecutionTiming       `json:"timing,omitempty"`
	Routing    *ExecutionRoutingResult `json:"routing,omitempty"`
	Governance *ExecutionGovernance   `json:"governance,omitempty"`
	Raw        map[string]any         `json:"raw,omitempty"`
}

// ExecutionUsage records token usage for an execution.
type ExecutionUsage struct {
	InputTokens  int      `json:"input_tokens"`
	OutputTokens int      `json:"output_tokens"`
	TotalTokens  int      `json:"total_tokens"`
	CostUSD      *float64 `json:"cost_usd,omitempty"`
}

// ExecutionTiming records timing information for an execution.
type ExecutionTiming struct {
	TotalMS    int `json:"total_ms"`
	PermitMS   int `json:"permit_ms,omitempty"`
	ProviderMS int `json:"provider_ms,omitempty"`
}

// ExecutionRoutingResult records which provider was used.
type ExecutionRoutingResult struct {
	Provider RoutingProvider `json:"provider"`
	Model    string          `json:"model"`
	Region   *string         `json:"region,omitempty"`
}

// ExecutionGovernance records governance details for an execution.
type ExecutionGovernance struct {
	PermitID  string   `json:"permit_id"`
	Decision  Decision `json:"decision"`
	PolicyIDs []string `json:"policy_ids,omitempty"`
}

// ExecutionStreamEvent represents a single SSE event during streaming execution.
type ExecutionStreamEvent struct {
	EventType string          `json:"event_type,omitempty"`
	Data      json.RawMessage `json:"data"`
}

// ExecuteRequest is the request body for the combined execute endpoint.
type ExecuteRequest struct {
	Subject        Subject                       `json:"subject"`
	Action         Action                        `json:"action"`
	Resource       Resource                      `json:"resource"`
	Context        *Context                      `json:"context,omitempty"`
	Conditions     *PermitConditions             `json:"conditions,omitempty"`
	Messages       []MessageInput                `json:"messages,omitempty"`
	Parameters     *ParametersInput              `json:"parameters,omitempty"`
	CustomMetadata map[string]any                `json:"custom_metadata,omitempty"`
}

// --- Job types ---

// JobSubmitRequest is the request body for submitting a batch job.
type JobSubmitRequest struct {
	Permit          PermitRequest  `json:"permit"`
	ProviderPayload map[string]any `json:"provider_payload,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// JobCreateResponse is the response from submitting a job.
type JobCreateResponse struct {
	JobID     string    `json:"job_id"`
	Status    JobStatus `json:"status"`
	CreatedAt string    `json:"created_at"`
}

// JobStatusResponse is the response from querying job status.
type JobStatusResponse struct {
	JobID         string           `json:"job_id"`
	RequestID     string           `json:"request_id"`
	PermitID      *string          `json:"permit_id,omitempty"`
	Provider      string           `json:"provider"`
	Model         string           `json:"model"`
	Operation     string           `json:"operation"`
	ExecutionMode string           `json:"execution_mode"`
	Status        JobStatus        `json:"status"`
	Progress      *float64         `json:"progress,omitempty"`
	CallbackURL   *string          `json:"callback_url,omitempty"`
	SubmittedAt   string           `json:"submitted_at"`
	StartedAt     *string          `json:"started_at,omitempty"`
	CompletedAt   *string          `json:"completed_at,omitempty"`
	ResultAssets  []map[string]any `json:"result_assets,omitempty"`
	UsageMetrics  []map[string]any `json:"usage_metrics,omitempty"`
	Error         map[string]any   `json:"error,omitempty"`
	Metadata      map[string]any   `json:"metadata,omitempty"`
}

// --- Request timeline types ---

// RequestTimelineEvent represents an event in the request timeline.
type RequestTimelineEvent struct {
	Timestamp string         `json:"timestamp"`
	Phase     string         `json:"phase"`
	Detail    map[string]any `json:"detail,omitempty"`
}

// RequestTimelineResponse is the response from the request timeline endpoint.
type RequestTimelineResponse struct {
	RequestID string                 `json:"request_id"`
	Events    []RequestTimelineEvent `json:"events"`
}

// ExecutionTimelineResponse is the timeline response for an execution.
type ExecutionTimelineResponse struct {
	RequestID string                 `json:"request_id"`
	Events    []RequestTimelineEvent `json:"events"`
}

// --- API Key types ---

// ApiKeyCreateRequest is the request body for creating an API key.
type ApiKeyCreateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Scope       string  `json:"scope,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

// ApiKeyRecord represents an API key in the system.
type ApiKeyRecord struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	Prefix      string  `json:"prefix"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Scope       string  `json:"scope"`
	CreatedBy   *string `json:"created_by,omitempty"`
	CreatedAt   string  `json:"created_at"`
	RevokedAt   *string `json:"revoked_at,omitempty"`
	LastUsedAt  *string `json:"last_used_at,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

// ApiKeyCreateResponse is the response from creating an API key.
type ApiKeyCreateResponse struct {
	ApiKeyRecord
	RawKey string `json:"raw_key"`
}

// ApiKeyListParams holds query parameters for listing API keys.
type ApiKeyListParams struct {
	Limit  *int    `json:"limit,omitempty"`
	Offset *int    `json:"offset,omitempty"`
	Status *string `json:"status,omitempty"`
}

// ApiKeyListResponse is the paginated list of API keys.
type ApiKeyListResponse struct {
	Items  []ApiKeyRecord `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// --- SSE types ---

// SSEEvent represents a parsed Server-Sent Event.
type SSEEvent struct {
	EventType string          `json:"event_type,omitempty"`
	Data      json.RawMessage `json:"data"`
}
