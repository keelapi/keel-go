package keel

import (
	"encoding/json"
	"strings"
	"time"
)

// Decision represents the permit decision outcome.
type Decision string

const (
	DecisionAllow     Decision = "allow"
	DecisionDeny      Decision = "deny"
	DecisionChallenge Decision = "challenge"

	// DecisionThrottled is retained for compatibility with older throttle payloads.
	// Current permit responses use decision="deny" with display_decision="throttle".
	DecisionThrottled Decision = "throttled"
)

// RoutingProvider represents a supported AI provider.
type RoutingProvider string

const (
	ProviderOpenAI    RoutingProvider = "openai"
	ProviderAnthropic RoutingProvider = "anthropic"
	ProviderGoogle    RoutingProvider = "google"
	ProviderXAI       RoutingProvider = "xai"
	ProviderMeta      RoutingProvider = "meta"
	ProviderMCP       RoutingProvider = "mcp"
)

type RoutingPriority string

const (
	RoutingPriorityCheap    RoutingPriority = "cheap"
	RoutingPriorityBalanced RoutingPriority = "balanced"
	RoutingPriorityQuality  RoutingPriority = "quality"
)

// ExecutionMode represents public execution endpoint modes.
type ExecutionMode string

const (
	ModeSync   ExecutionMode = "sync"
	ModeStream ExecutionMode = "stream"
)

// ResourceExecutionMode represents permit resource execution modes.
type ResourceExecutionMode = ExecutionMode

const (
	ResourceModeSync     ResourceExecutionMode = "sync"
	ResourceModeAsync    ResourceExecutionMode = "async"
	ResourceModeRealtime ResourceExecutionMode = "realtime"
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

type Confidence string

const (
	ConfidenceExact    Confidence = "exact"
	ConfidenceInferred Confidence = "inferred"
	ConfidenceMissing  Confidence = "missing"
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

type IdentityTenant struct {
	ID         *string    `json:"id,omitempty"`
	Name       *string    `json:"name,omitempty"`
	ExternalID *string    `json:"external_id,omitempty"`
	Confidence Confidence `json:"confidence,omitempty"`
}

type IdentitySubject struct {
	Type       *string    `json:"type,omitempty"`
	ID         *string    `json:"id,omitempty"`
	Name       *string    `json:"name,omitempty"`
	Email      *string    `json:"email,omitempty"`
	ExternalID *string    `json:"external_id,omitempty"`
	Confidence Confidence `json:"confidence,omitempty"`
}

type IdentityCredential struct {
	Type       *string    `json:"type,omitempty"`
	ID         *string    `json:"id,omitempty"`
	Name       *string    `json:"name,omitempty"`
	ExternalID *string    `json:"external_id,omitempty"`
	Confidence Confidence `json:"confidence,omitempty"`
}

type IdentityEnvironment struct {
	Name       *string    `json:"name,omitempty"`
	ExternalID *string    `json:"external_id,omitempty"`
	Confidence Confidence `json:"confidence,omitempty"`
}

type Identity struct {
	Tenant      *IdentityTenant      `json:"tenant,omitempty"`
	Subject     *IdentitySubject     `json:"subject,omitempty"`
	Credential  *IdentityCredential  `json:"credential,omitempty"`
	Environment *IdentityEnvironment `json:"environment,omitempty"`
}

// PermitInput represents an input descriptor in a permit resource.
type PermitInput struct {
	Type      string         `json:"type"`
	Purpose   *string        `json:"purpose,omitempty"`
	Source    *string        `json:"source,omitempty"`
	URL       *string        `json:"url,omitempty"`
	FileID    *string        `json:"file_id,omitempty"`
	Filename  *string        `json:"filename,omitempty"`
	MimeType  *string        `json:"mime_type,omitempty"`
	SizeBytes *int64         `json:"size_bytes,omitempty"`
	SHA256    *string        `json:"sha256,omitempty"`
	AssetID   *string        `json:"asset_id,omitempty"`
	Base64    *string        `json:"base64,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`

	// Content is kept for source compatibility with older SDK versions. When
	// marshaled, its keys are merged into the descriptor without overriding typed fields.
	Content map[string]any `json:"-"`
}

func (p PermitInput) MarshalJSON() ([]byte, error) {
	type alias PermitInput
	return marshalInputDescriptorWithContent(alias(p), p.Content, map[string]bool{
		"type":       true,
		"purpose":    true,
		"source":     true,
		"url":        true,
		"file_id":    true,
		"filename":   true,
		"mime_type":  true,
		"size_bytes": true,
		"sha256":     true,
		"asset_id":   true,
		"base64":     true,
		"metadata":   true,
	})
}

type ImageRequestAttributes struct {
	ModelID           *string              `json:"model_id,omitempty"`
	Operation         *CapabilityOperation `json:"operation,omitempty"`
	Modality          *string              `json:"modality,omitempty"`
	Size              *string              `json:"size,omitempty"`
	Quality           *string              `json:"quality,omitempty"`
	N                 *int                 `json:"n,omitempty"`
	Background        *string              `json:"background,omitempty"`
	OutputFormat      *string              `json:"output_format,omitempty"`
	OutputCompression *int                 `json:"output_compression,omitempty"`
	Moderation        *string              `json:"moderation,omitempty"`
	InputFidelity     *string              `json:"input_fidelity,omitempty"`
	Stream            *bool                `json:"stream,omitempty"`
	PartialImages     *int                 `json:"partial_images,omitempty"`
	HasImageInputs    *bool                `json:"has_image_inputs,omitempty"`
	ImageInputCount   *int                 `json:"image_input_count,omitempty"`
	HasMask           *bool                `json:"has_mask,omitempty"`
}

// ResourceAttributes describes the attributes of a resource in a permit request.
type ResourceAttributes struct {
	Provider                 string                  `json:"provider"`
	Model                    string                  `json:"model"`
	Operation                CapabilityOperation     `json:"operation,omitempty"`
	Modality                 string                  `json:"modality,omitempty"`
	ExecutionMode            ResourceExecutionMode   `json:"execution_mode,omitempty"`
	EstimatedInputTokens     int                     `json:"estimated_input_tokens"`
	EstimatedOutputTokens    int                     `json:"estimated_output_tokens"`
	MaxOutputTokensRequested *int                    `json:"max_output_tokens_requested,omitempty"`
	Inputs                   []PermitInput           `json:"inputs,omitempty"`
	AssetSummary             map[string]any          `json:"asset_summary,omitempty"`
	ImageRequest             *ImageRequestAttributes `json:"image_request,omitempty"`
	PromptContentHash        *string                 `json:"prompt_content_hash,omitempty"`
	Routing                  *RoutingPreferences     `json:"routing,omitempty"`
	CallbackURL              *string                 `json:"callback_url,omitempty"`
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
	Timestamp *time.Time     `json:"timestamp,omitempty"`
	IP        *string        `json:"ip,omitempty"`
	UserAgent *string        `json:"user_agent,omitempty"`
	UserID    *string        `json:"user_id,omitempty"`
	OrgID     *string        `json:"org_id,omitempty"`
	Role      *string        `json:"role,omitempty"`
	Extra     map[string]any `json:"-"`
}

func (c Context) MarshalJSON() ([]byte, error) {
	type alias Context
	return marshalWithExtra(alias(c), c.Extra)
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
	Match string                `json:"match,omitempty"`
	Rules []PermitConditionRule `json:"rules"`
}

type ActorRef struct {
	Issuer       string  `json:"issuer"`
	Subject      string  `json:"subject"`
	DisplayLabel *string `json:"display_label,omitempty"`
}

type AuthorityEnvelope struct {
	Actions     []string   `json:"actions,omitempty"`
	Tools       []string   `json:"tools,omitempty"`
	Providers   []string   `json:"providers,omitempty"`
	Models      []string   `json:"models,omitempty"`
	DataClasses []string   `json:"data_classes,omitempty"`
	Regions     []string   `json:"regions,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type UsageLimits struct {
	MaxCalls             *int     `json:"max_calls,omitempty"`
	MaxTokens            *int     `json:"max_tokens,omitempty"`
	EstimatedCostCeiling *float64 `json:"estimated_cost_ceiling,omitempty"`
}

// PermitRequest is the request body for creating a permit.
type PermitRequest struct {
	ProjectID                string             `json:"project_id"`
	ParentPermitID           *string            `json:"parent_permit_id,omitempty"`
	BudgetEnvelopeID         *string            `json:"budget_envelope_id,omitempty"`
	IdempotencyKey           string             `json:"idempotency_key"`
	SessionID                *string            `json:"session_id,omitempty"`
	Subject                  Subject            `json:"subject"`
	Action                   Action             `json:"action"`
	Resource                 Resource           `json:"resource"`
	Context                  *Context           `json:"context,omitempty"`
	Conditions               *PermitConditions  `json:"conditions,omitempty"`
	Identity                 *Identity          `json:"identity,omitempty"`
	ActorRef                 *ActorRef          `json:"actor_ref,omitempty"`
	AuthorityEnvelope        *AuthorityEnvelope `json:"authority_envelope,omitempty"`
	AuthorityEnvelopeVersion *string            `json:"authority_envelope_version,omitempty"`
	UsageLimits              *UsageLimits       `json:"usage_limits,omitempty"`
	CustomMetadata           map[string]any     `json:"custom_metadata,omitempty"`
}

// --- Response types ---

// PermitActionResponse is an action returned in a permit response.
type PermitActionResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ProjectedCostMethodology struct {
	Basis                 *string `json:"basis,omitempty"`
	Provenance            *string `json:"provenance,omitempty"`
	InputTokensEstimated  *int    `json:"input_tokens_estimated,omitempty"`
	OutputTokensEstimated *int    `json:"output_tokens_estimated,omitempty"`
	Quality               *string `json:"quality,omitempty"`
	PricingTableID        *string `json:"pricing_table_id,omitempty"`
	Tokenizer             *string `json:"tokenizer,omitempty"`
}

type ProjectedCost struct {
	AmountMicros *int64                    `json:"amount_micros,omitempty"`
	Currency     string                    `json:"currency,omitempty"`
	Methodology  *ProjectedCostMethodology `json:"methodology,omitempty"`
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
	Decision string         `json:"decision,omitempty"`
	Code     string         `json:"code,omitempty"`
	Reason   string         `json:"reason,omitempty"`
	Extra    map[string]any `json:"-"`
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
	PermitID         string                 `json:"permit_id"`
	ProjectID        string                 `json:"project_id"`
	ParentPermitID   *string                `json:"parent_permit_id,omitempty"`
	DelegationDepth  int                    `json:"delegation_depth"`
	Decision         Decision               `json:"decision"`
	Reason           string                 `json:"reason"`
	ReasonCode       *string                `json:"reason_code,omitempty"`
	ReasonDetail     map[string]any         `json:"reason_detail,omitempty"`
	OutcomeDetail    map[string]any         `json:"outcome_detail,omitempty"`
	Message          *string                `json:"message,omitempty"`
	Status           *string                `json:"status,omitempty"`
	Constraints      map[string]any         `json:"constraints,omitempty"`
	Budgets          map[string]any         `json:"budgets,omitempty"`
	Actions          []PermitActionResponse `json:"actions"`
	DecisionDetails  *PermitDecisionDetails `json:"decision_details,omitempty"`
	Routing          *RoutingDecision       `json:"routing,omitempty"`
	Metadata         PermitMetadata         `json:"metadata"`
	DisplayDecision  string                 `json:"display_decision,omitempty"`
	DisplayReason    string                 `json:"display_reason,omitempty"`
	DecisionSource   string                 `json:"decision_source,omitempty"`
	EstimatedCostUSD *float64               `json:"estimated_cost_usd,omitempty"`
	ProjectedCost    *ProjectedCost         `json:"projected_cost,omitempty"`
}

type PermitConditionEvaluation struct {
	Outcome         string         `json:"outcome"`
	EvaluatedAt     string         `json:"evaluated_at"`
	Match           string         `json:"match"`
	RuleCount       int            `json:"rule_count"`
	FailedRuleIndex *int           `json:"failed_rule_index,omitempty"`
	Code            *string        `json:"code,omitempty"`
	Message         *string        `json:"message,omitempty"`
	Field           *string        `json:"field,omitempty"`
	Operator        *string        `json:"operator,omitempty"`
	ExpectedValue   any            `json:"expected_value,omitempty"`
	ActualValue     any            `json:"actual_value,omitempty"`
	Extra           map[string]any `json:"-"`
}

type BudgetEnvelopeSummary struct {
	EnvelopeID               string `json:"envelope_id"`
	ReservationStatus        string `json:"reservation_status,omitempty"`
	EstimatedCostUSDMicros   *int64 `json:"estimated_cost_usd_micros,omitempty"`
	TotalBudgetUSDMicros     *int64 `json:"total_budget_usd_micros,omitempty"`
	ReservedBudgetUSDMicros  *int64 `json:"reserved_budget_usd_micros,omitempty"`
	SpentBudgetUSDMicros     *int64 `json:"spent_budget_usd_micros,omitempty"`
	RemainingBudgetUSDMicros *int64 `json:"remaining_budget_usd_micros,omitempty"`
	CorrectionUSDMicros      *int64 `json:"correction_usd_micros,omitempty"`
}

// PermitDryRunResponse is the response from a dry-run permit evaluation.
type PermitDryRunResponse struct {
	DryRun              bool                       `json:"dry_run"`
	ProjectID           string                     `json:"project_id"`
	ExistingPermitID    *string                    `json:"existing_permit_id,omitempty"`
	ReplayOutcome       string                     `json:"replay_outcome,omitempty"`
	ParentPermitID      *string                    `json:"parent_permit_id,omitempty"`
	DelegationDepth     int                        `json:"delegation_depth"`
	Decision            Decision                   `json:"decision"`
	Reason              string                     `json:"reason"`
	ReasonCode          *string                    `json:"reason_code,omitempty"`
	ReasonDetail        map[string]any             `json:"reason_detail,omitempty"`
	Message             *string                    `json:"message,omitempty"`
	Status              *string                    `json:"status,omitempty"`
	Constraints         map[string]any             `json:"constraints,omitempty"`
	Budgets             map[string]any             `json:"budgets,omitempty"`
	Actions             []PermitActionResponse     `json:"actions,omitempty"`
	DecisionDetails     *PermitDecisionDetails     `json:"decision_details,omitempty"`
	Routing             *RoutingDecision           `json:"routing,omitempty"`
	Metadata            PermitMetadata             `json:"metadata"`
	DisplayDecision     string                     `json:"display_decision,omitempty"`
	DisplayReason       string                     `json:"display_reason,omitempty"`
	DecisionSource      string                     `json:"decision_source,omitempty"`
	EstimatedCostUSD    *float64                   `json:"estimated_cost_usd,omitempty"`
	ProjectedCost       *ProjectedCost             `json:"projected_cost,omitempty"`
	BudgetEnvelope      *BudgetEnvelopeSummary     `json:"budget_envelope,omitempty"`
	ConditionEvaluation *PermitConditionEvaluation `json:"condition_evaluation,omitempty"`
}

// PermitAuditItem represents a permit in the audit log.
type PermitAuditItem struct {
	ID                         string                     `json:"id"`
	PermitID                   string                     `json:"permit_id,omitempty"`
	ProjectID                  string                     `json:"project_id"`
	ParentPermitID             *string                    `json:"parent_permit_id,omitempty"`
	DelegationDepth            int                        `json:"delegation_depth"`
	SubjectType                string                     `json:"subject_type"`
	SubjectID                  string                     `json:"subject_id"`
	ActionName                 string                     `json:"action_name"`
	ResourceProvider           string                     `json:"resource_provider"`
	ResourceModel              string                     `json:"resource_model"`
	ResourceOperation          *string                    `json:"resource_operation,omitempty"`
	ResourceModality           *string                    `json:"resource_modality,omitempty"`
	ResourceInputsJSON         []map[string]any           `json:"resource_inputs_json,omitempty"`
	ResourceAttributesJSON     map[string]any             `json:"resource_attributes_json,omitempty"`
	ConnectorID                *string                    `json:"connector_id,omitempty"`
	ConnectorType              *string                    `json:"connector_type,omitempty"`
	ConnectorDisplayName       *string                    `json:"connector_display_name,omitempty"`
	EstimatedInputTokens       int                        `json:"estimated_input_tokens"`
	EstimatedOutputTokens      int                        `json:"estimated_output_tokens"`
	MaxOutputTokensRequested   *int                       `json:"max_output_tokens_requested,omitempty"`
	Decision                   Decision                   `json:"decision"`
	Reason                     string                     `json:"reason"`
	Status                     *string                    `json:"status,omitempty"`
	ConstraintsJSON            map[string]any             `json:"constraints_json,omitempty"`
	BudgetsJSON                map[string]any             `json:"budgets_json,omitempty"`
	ProjectedCost              *ProjectedCost             `json:"projected_cost,omitempty"`
	BudgetEnvelope             *BudgetEnvelopeSummary     `json:"budget_envelope,omitempty"`
	ActionsJSON                []map[string]any           `json:"actions_json,omitempty"`
	Conditions                 *PermitConditions          `json:"conditions,omitempty"`
	ConditionEvaluation        *PermitConditionEvaluation `json:"condition_evaluation,omitempty"`
	DecisionDetails            *PermitDecisionDetails     `json:"decision_details,omitempty"`
	Routing                    *RoutingDecision           `json:"routing,omitempty"`
	ActualUsageJSON            map[string]any             `json:"actual_usage_json,omitempty"`
	AccountingDisposition      *string                    `json:"accounting_disposition,omitempty"`
	UsageVerificationMethod    *string                    `json:"usage_verification_method,omitempty"`
	UsageVerificationStatus    *string                    `json:"usage_verification_status,omitempty"`
	UsageVerificationUpdatedAt *string                    `json:"usage_verification_updated_at,omitempty"`
	PolicyID                   string                     `json:"policy_id"`
	PolicyVersion              string                     `json:"policy_version"`
	Metadata                   PermitMetadata             `json:"metadata"`
	IdempotencyKey             string                     `json:"idempotency_key"`
	RequestFingerprint         string                     `json:"request_fingerprint"`
	ExpiresAt                  *string                    `json:"expires_at,omitempty"`
	CreatedAt                  string                     `json:"created_at"`
	ProviderResponseDigestV1   *string                    `json:"provider_response_digest_v1,omitempty"`
	ClientResponseDigestV1     *string                    `json:"client_response_digest_v1,omitempty"`
	ClosureStatus              *string                    `json:"closure_status,omitempty"`
	ClosureSignatureB64        *string                    `json:"closure_signature_b64,omitempty"`
	ClosureCanonicalHash       *string                    `json:"closure_canonical_hash,omitempty"`
	ClosureSignedAt            *string                    `json:"closure_signed_at,omitempty"`
	Identity                   *Identity                  `json:"identity,omitempty"`
}

func (p *PermitAuditItem) UnmarshalJSON(data []byte) error {
	type alias PermitAuditItem
	var item alias
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*p = PermitAuditItem(item)
	if p.PermitID == "" {
		p.PermitID = p.ID
	}
	return nil
}

// PermitAttestation records attestation details for a used permit.
type PermitAttestation struct {
	Status     string `json:"status"`
	AttestedAt string `json:"attested_at"`
	AttestedBy string `json:"attested_by,omitempty"`
}

type GovernanceEventItem map[string]any

// PermitAuditBundle contains a full audit bundle for a permit.
type PermitAuditBundle struct {
	BundleType       string                `json:"bundle_type"`
	SchemaVersion    int                   `json:"schema_version"`
	Format           string                `json:"format"`
	ProjectID        string                `json:"project_id"`
	PermitID         string                `json:"permit_id"`
	GeneratedAt      string                `json:"generated_at"`
	Record           AuditExportRecord     `json:"record"`
	GovernanceEvents []GovernanceEventItem `json:"governance_events,omitempty"`

	// Legacy fields are retained for callers compiled against older versions.
	Permit    *PermitAuditItem       `json:"permit,omitempty"`
	Execution *ExecutionResponse     `json:"execution,omitempty"`
	Timeline  []RequestTimelineEvent `json:"timeline,omitempty"`
	Evidence  []PermitEvidenceOut    `json:"evidence,omitempty"`
}

// PermitListParams holds query parameters for listing permits.
type PermitListParams struct {
	Before      *string `json:"before,omitempty"`
	Cursor      *string `json:"cursor,omitempty"` // Deprecated: use Before.
	Limit       *int    `json:"limit,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
	View        *string `json:"view,omitempty"`
	Type        *string `json:"type,omitempty"`
	SubjectID   *string `json:"subject_id,omitempty"`
	SubjectType *string `json:"subject_type,omitempty"`
	ActionName  *string `json:"action_name,omitempty"`
	PolicyID    *string `json:"policy_id,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	Decision    *string `json:"decision,omitempty"`
}

// PermitsListResponse is the paginated list of permit audit items.
type PermitsListResponse struct {
	Items      []PermitAuditItem `json:"items"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

// PermitAuditListResponse is an alias for backward compatibility.
type PermitAuditListResponse = PermitsListResponse

type PermitExportFormat string

const (
	PermitExportJSON PermitExportFormat = "json"
	PermitExportCSV  PermitExportFormat = "csv"
)

// PermitExportResponse is the legacy response wrapper for permit exports.
type PermitExportResponse struct {
	Data   string `json:"data"`
	Format string `json:"format"`
	Count  int    `json:"count"`
}

// PermitAttestationRequest is the request body for attesting a permit.
type PermitAttestationRequest struct {
	Attestor        string  `json:"attestor,omitempty"`
	AttestationType string  `json:"attestation_type"`
	Rationale       *string `json:"rationale,omitempty"`
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
	ID            string `json:"id"`
	EvidenceID    string `json:"evidence_id,omitempty"`
	PermitID      string `json:"permit_id"`
	ProjectID     string `json:"project_id"`
	EvidenceType  string `json:"evidence_type"`
	EvidenceValue string `json:"evidence_value"`
	Label         string `json:"label"`
	AttachedBy    string `json:"attached_by"`
	AttachedAt    string `json:"attached_at"`

	// Legacy response fields.
	Type      string         `json:"type,omitempty"`
	Content   map[string]any `json:"content,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
}

func (p *PermitEvidenceOut) UnmarshalJSON(data []byte) error {
	type alias PermitEvidenceOut
	var item alias
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*p = PermitEvidenceOut(item)
	if p.EvidenceID == "" {
		p.EvidenceID = p.ID
	}
	return nil
}

// PermitEvidenceListResponse is the response from listing evidence.
type PermitEvidenceListResponse struct {
	Items []PermitEvidenceOut `json:"items"`
}

// PermitLineageNodeSummary represents a node summary in a permit lineage graph.
type PermitLineageNodeSummary struct {
	PermitID             string   `json:"permit_id"`
	ParentPermitID       *string  `json:"parent_permit_id,omitempty"`
	ParentID             *string  `json:"parent_id,omitempty"` // Deprecated alias.
	DelegationDepth      int      `json:"delegation_depth"`
	Decision             Decision `json:"decision"`
	Status               *string  `json:"status,omitempty"`
	Reason               string   `json:"reason"`
	ActionName           string   `json:"action_name"`
	ResourceProvider     string   `json:"resource_provider"`
	ResourceModel        string   `json:"resource_model"`
	ConnectorID          *string  `json:"connector_id,omitempty"`
	ConnectorType        *string  `json:"connector_type,omitempty"`
	ConnectorDisplayName *string  `json:"connector_display_name,omitempty"`
	CreatedAt            string   `json:"created_at"`
	IssuedAt             string   `json:"issued_at,omitempty"` // Deprecated alias.
	ExpiresAt            *string  `json:"expires_at,omitempty"`
	UsageReportedAt      *string  `json:"usage_reported_at,omitempty"`
	ActualTotalTokens    *int     `json:"actual_total_tokens,omitempty"`
	ActualCostUSDMicros  *int64   `json:"actual_cost_usd_micros,omitempty"`
}

// PermitLineageNode represents a node in a permit's lineage graph.
type PermitLineageNode struct {
	PermitLineageNodeSummary
	Children []PermitLineageNode `json:"children,omitempty"`
}

// PermitLineageResponse is the response from querying permit lineage.
type PermitLineageResponse struct {
	ProjectID       string                     `json:"project_id"`
	RootPermitID    string                     `json:"root_permit_id"`
	CurrentPermitID string                     `json:"current_permit_id"`
	Parent          *PermitLineageNodeSummary  `json:"parent,omitempty"`
	Ancestors       []PermitLineageNodeSummary `json:"ancestors,omitempty"`
	Current         PermitLineageNode          `json:"current"`
}

type PermitUsageVerificationInput map[string]any
type PermitUsageVerificationSummary map[string]any

// PermitUsageReportRequest is the request body for reporting usage on a permit.
type PermitUsageReportRequest struct {
	Provider            *string                      `json:"provider,omitempty"`
	Model               *string                      `json:"model,omitempty"`
	ActualInputTokens   *int                         `json:"actual_input_tokens,omitempty"`
	ActualOutputTokens  *int                         `json:"actual_output_tokens,omitempty"`
	ActualTotalTokens   *int                         `json:"actual_total_tokens,omitempty"`
	CostUSD             *string                      `json:"cost_usd,omitempty"`
	CostUSDMicros       *int64                       `json:"cost_usd_micros,omitempty"`
	UsageIdempotencyKey *string                      `json:"usage_idempotency_key,omitempty"`
	NormalizedUsage     map[string]any               `json:"normalized_usage,omitempty"`
	UsageMetrics        []map[string]any             `json:"usage_metrics,omitempty"`
	Metadata            map[string]any               `json:"metadata,omitempty"`
	Outcome             *string                      `json:"outcome,omitempty"`
	ErrorClass          *string                      `json:"error_class,omitempty"`
	OperationMetadata   map[string]any               `json:"operation_metadata,omitempty"`
	Verification        PermitUsageVerificationInput `json:"verification,omitempty"`
}

// PermitUsageReportResponse is the response from reporting usage.
type PermitUsageReportResponse struct {
	PermitID            string                         `json:"permit_id"`
	ProjectID           string                         `json:"project_id"`
	UsageReportedAt     *string                        `json:"usage_reported_at,omitempty"`
	ActualInputTokens   *int                           `json:"actual_input_tokens,omitempty"`
	ActualOutputTokens  *int                           `json:"actual_output_tokens,omitempty"`
	ActualTotalTokens   *int                           `json:"actual_total_tokens,omitempty"`
	ActualCostUSDMicros *int64                         `json:"actual_cost_usd_micros,omitempty"`
	ActualUsageJSON     map[string]any                 `json:"actual_usage_json,omitempty"`
	UsageSource         *string                        `json:"usage_source,omitempty"`
	UsageVerification   PermitUsageVerificationSummary `json:"usage_verification,omitempty"`
	Status              *string                        `json:"status,omitempty"`
}

// --- Routing types ---

// RoutingTarget identifies a specific provider and model.
type RoutingTarget struct {
	Provider RoutingProvider `json:"provider"`
	Model    string          `json:"model"`
}

// RoutingPreferences express caller preferences for routing.
type RoutingPreferences struct {
	Priority                   RoutingPriority `json:"priority,omitempty"`
	TaskType                   *string         `json:"task_type,omitempty"`
	LatencyTargetMS            *int            `json:"latency_target_ms,omitempty"`
	AllowCrossProviderFallback bool            `json:"allow_cross_provider_fallback,omitempty"`
	FallbackChain              []RoutingTarget `json:"fallback_chain,omitempty"`

	// Legacy fields are ignored by JSON marshaling because the API forbids them.
	PreferredProviders []RoutingProvider `json:"-"`
	PreferredModels    []string          `json:"-"`
	Fallback           *bool             `json:"-"`
}

func (r RoutingPreferences) MarshalJSON() ([]byte, error) {
	allowFallback := r.AllowCrossProviderFallback
	if r.Fallback != nil {
		allowFallback = *r.Fallback
	}
	type wire struct {
		Priority                   RoutingPriority `json:"priority,omitempty"`
		TaskType                   *string         `json:"task_type,omitempty"`
		LatencyTargetMS            *int            `json:"latency_target_ms,omitempty"`
		AllowCrossProviderFallback bool            `json:"allow_cross_provider_fallback,omitempty"`
		FallbackChain              []RoutingTarget `json:"fallback_chain,omitempty"`
	}
	return json.Marshal(wire{
		Priority:                   r.Priority,
		TaskType:                   r.TaskType,
		LatencyTargetMS:            r.LatencyTargetMS,
		AllowCrossProviderFallback: allowFallback,
		FallbackChain:              r.FallbackChain,
	})
}

// RoutingDecision is the routing result embedded in a permit or execution.
type RoutingDecision struct {
	DecisionID             string          `json:"decision_id,omitempty"`
	DecidedAt              string          `json:"decided_at,omitempty"`
	RequestedProvider      RoutingProvider `json:"requested_provider"`
	RequestedModel         string          `json:"requested_model"`
	SelectedProvider       RoutingProvider `json:"selected_provider"`
	SelectedModel          string          `json:"selected_model"`
	ReasonCode             string          `json:"reason_code"`
	FallbackOccurred       bool            `json:"fallback_occurred"`
	PolicyID               *string         `json:"policy_id,omitempty"`
	PolicyVersion          *string         `json:"policy_version,omitempty"`
	EstimatedCostUSDMicros *int64          `json:"estimated_cost_usd_micros,omitempty"`
	Priority               RoutingPriority `json:"priority,omitempty"`
	TaskType               *string         `json:"task_type,omitempty"`
	LatencyTargetMS        *int            `json:"latency_target_ms,omitempty"`
	FallbackChain          []RoutingTarget `json:"fallback_chain,omitempty"`
	ReasonMetadata         map[string]any  `json:"reason_metadata,omitempty"`

	// Legacy aliases for older responses.
	Provider RoutingProvider `json:"provider,omitempty"`
	Model    string          `json:"model,omitempty"`
	Region   *string         `json:"region,omitempty"`
}

// --- Execution types ---

// MessageInput represents a message in an execution request.
type MessageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AssetInput describes an asset supplied to an execution request.
type AssetInput struct {
	Source    string         `json:"source"`
	URL       *string        `json:"url,omitempty"`
	FileID    *string        `json:"file_id,omitempty"`
	AssetID   *string        `json:"asset_id,omitempty"`
	Base64    *string        `json:"base64,omitempty"`
	MimeType  *string        `json:"mime_type,omitempty"`
	Filename  *string        `json:"filename,omitempty"`
	SizeBytes *int64         `json:"size_bytes,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// InputPart represents an input part in an execution request.
type InputPart struct {
	Type     string         `json:"type"`
	Role     *string        `json:"role,omitempty"`
	Text     *string        `json:"text,omitempty"`
	JSON     map[string]any `json:"json,omitempty"`
	Asset    *AssetInput    `json:"asset,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`

	// Content is kept for source compatibility with older SDK versions. It is not
	// sent because the current API accepts typed fields.
	Content map[string]any `json:"-"`
}

func (p InputPart) MarshalJSON() ([]byte, error) {
	type alias InputPart
	return marshalInputDescriptorWithContent(alias(p), p.Content, map[string]bool{
		"type":     true,
		"role":     true,
		"text":     true,
		"json":     true,
		"asset":    true,
		"metadata": true,
	})
}

// RoutingInput represents routing preferences in an execution request.
type RoutingInput struct {
	Provider                   RoutingProvider `json:"provider,omitempty"`
	Model                      *string         `json:"model,omitempty"`
	Priority                   RoutingPriority `json:"priority,omitempty"`
	TaskType                   *string         `json:"task_type,omitempty"`
	LatencyTargetMS            *int            `json:"latency_target_ms,omitempty"`
	AllowCrossProviderFallback bool            `json:"allow_cross_provider_fallback,omitempty"`
	FallbackChain              []RoutingTarget `json:"fallback_chain,omitempty"`

	// Legacy fields are ignored by JSON marshaling because the API forbids them.
	PreferredProviders []RoutingProvider `json:"-"`
	PreferredModels    []string          `json:"-"`
	Fallback           *bool             `json:"-"`
}

func (r RoutingInput) MarshalJSON() ([]byte, error) {
	provider := r.Provider
	if provider == "" && len(r.PreferredProviders) == 1 {
		provider = r.PreferredProviders[0]
	}
	model := r.Model
	if model == nil && len(r.PreferredModels) == 1 {
		modelValue := r.PreferredModels[0]
		model = &modelValue
	}
	allowFallback := r.AllowCrossProviderFallback
	if r.Fallback != nil {
		allowFallback = *r.Fallback
	}
	type wire struct {
		Provider                   RoutingProvider `json:"provider,omitempty"`
		Model                      *string         `json:"model,omitempty"`
		Priority                   RoutingPriority `json:"priority,omitempty"`
		TaskType                   *string         `json:"task_type,omitempty"`
		LatencyTargetMS            *int            `json:"latency_target_ms,omitempty"`
		AllowCrossProviderFallback bool            `json:"allow_cross_provider_fallback,omitempty"`
		FallbackChain              []RoutingTarget `json:"fallback_chain,omitempty"`
	}
	return json.Marshal(wire{
		Provider:                   provider,
		Model:                      model,
		Priority:                   r.Priority,
		TaskType:                   r.TaskType,
		LatencyTargetMS:            r.LatencyTargetMS,
		AllowCrossProviderFallback: allowFallback,
		FallbackChain:              r.FallbackChain,
	})
}

// ParametersInput represents parameters for an execution request.
type ParametersInput struct {
	MaxOutputTokens *int     `json:"max_output_tokens,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"top_p,omitempty"`

	// Deprecated: use MaxOutputTokens. When marshaling, MaxTokens is emitted as
	// max_output_tokens only when MaxOutputTokens is nil.
	MaxTokens *int `json:"-"`

	// Deprecated execution parameters from older SDK versions. The API forbids
	// arbitrary fields here, so these are intentionally not sent.
	Extra map[string]any `json:"-"`
}

func (p ParametersInput) MarshalJSON() ([]byte, error) {
	maxOutputTokens := p.MaxOutputTokens
	if maxOutputTokens == nil {
		maxOutputTokens = p.MaxTokens
	}
	type wire struct {
		MaxOutputTokens *int     `json:"max_output_tokens,omitempty"`
		Temperature     *float64 `json:"temperature,omitempty"`
		TopP            *float64 `json:"top_p,omitempty"`
	}
	return json.Marshal(wire{
		MaxOutputTokens: maxOutputTokens,
		Temperature:     p.Temperature,
		TopP:            p.TopP,
	})
}

type ImageOptionsInput struct {
	Size              *string `json:"size,omitempty"`
	Quality           *string `json:"quality,omitempty"`
	N                 *int    `json:"n,omitempty"`
	Background        *string `json:"background,omitempty"`
	OutputFormat      *string `json:"output_format,omitempty"`
	OutputCompression *int    `json:"output_compression,omitempty"`
	Moderation        *string `json:"moderation,omitempty"`
}

// ExecutionCreateRequest is the request body for creating an execution.
type ExecutionCreateRequest struct {
	Operation       string                    `json:"operation"`
	Mode            ExecutionMode             `json:"mode,omitempty"`
	Messages        []MessageInput            `json:"messages,omitempty"`
	Inputs          []InputPart               `json:"inputs,omitempty"`
	Mask            *AssetInput               `json:"mask,omitempty"`
	ImageOptions    *ImageOptionsInput        `json:"image_options,omitempty"`
	Routing         *RoutingInput             `json:"routing,omitempty"`
	Parameters      *ParametersInput          `json:"parameters,omitempty"`
	ProviderOptions map[string]map[string]any `json:"provider_options,omitempty"`
}

// ExecutionMessage is an alias for MessageInput for backward compatibility.
type ExecutionMessage = MessageInput

type ExecutionOutputContentPart struct {
	Type  string         `json:"type"`
	Role  *string        `json:"role,omitempty"`
	Text  *string        `json:"text,omitempty"`
	JSON  map[string]any `json:"json,omitempty"`
	Extra map[string]any `json:"-"`
}

type ExecutionOutput struct {
	Content    []ExecutionOutputContentPart `json:"content,omitempty"`
	Embeddings []any                        `json:"embeddings,omitempty"`
	Raw        map[string]any               `json:"raw,omitempty"`

	// Parts is retained for compatibility with older prerelease responses.
	Parts []ExecutionOutputContentPart `json:"parts,omitempty"`
}

type ExecutionOutputAsset struct {
	AssetID   string         `json:"asset_id"`
	Type      string         `json:"type"`
	MimeType  string         `json:"mime_type"`
	SizeBytes *int64         `json:"size_bytes,omitempty"`
	URI       *string        `json:"uri,omitempty"`
	Status    *string        `json:"status,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ExecutionResponse is the normalized execution envelope.
type ExecutionResponse struct {
	ID           string                  `json:"id"`
	Object       string                  `json:"object"`
	CreatedAt    *string                 `json:"created_at,omitempty"`
	Status       string                  `json:"status"`
	StatusCode   int                     `json:"status_code"`
	Output       *ExecutionOutput        `json:"output,omitempty"`
	OutputAssets []ExecutionOutputAsset  `json:"output_assets"`
	Routing      *ExecutionRoutingResult `json:"routing,omitempty"`
	Governance   *ExecutionGovernance    `json:"governance,omitempty"`
	Usage        *ExecutionUsage         `json:"usage,omitempty"`
	Timing       *ExecutionTiming        `json:"timing,omitempty"`
	Error        *ExecutionErrorPayload  `json:"error,omitempty"`
	Resolved     *ExecutionResolved      `json:"resolved,omitempty"`

	// Legacy fields retained for source compatibility with older SDK versions.
	RequestID string          `json:"request_id,omitempty"`
	PermitID  string          `json:"permit_id,omitempty"`
	Provider  RoutingProvider `json:"provider,omitempty"`
	Model     string          `json:"model,omitempty"`
	Messages  []MessageInput  `json:"messages,omitempty"`
	Raw       map[string]any  `json:"raw,omitempty"`
}

func (e *ExecutionResponse) UnmarshalJSON(data []byte) error {
	type alias ExecutionResponse
	var item alias
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*e = ExecutionResponse(item)
	if e.RequestID == "" && strings.HasPrefix(e.ID, "exec_") {
		e.RequestID = strings.TrimPrefix(e.ID, "exec_")
	}
	if e.Provider == "" && e.Routing != nil {
		e.Provider = e.Routing.SelectedProvider
	}
	if e.Provider == "" && e.Routing != nil {
		e.Provider = e.Routing.Provider
	}
	if e.Model == "" && e.Routing != nil {
		e.Model = e.Routing.SelectedModel
	}
	if e.Model == "" && e.Routing != nil {
		e.Model = e.Routing.Model
	}
	if e.PermitID == "" && e.Governance != nil {
		e.PermitID = e.Governance.PermitID
	}
	return nil
}

// ExecutionUsage records token usage for an execution.
type ExecutionUsage struct {
	InputTokens    int              `json:"input_tokens,omitempty"`
	OutputTokens   int              `json:"output_tokens,omitempty"`
	TotalTokens    int              `json:"total_tokens,omitempty"`
	CostUSDMicros  *int64           `json:"cost_usd_micros,omitempty"`
	CostUSD        *float64         `json:"cost_usd,omitempty"`
	EstimatedFinal bool             `json:"estimated_final,omitempty"`
	Metrics        []map[string]any `json:"metrics,omitempty"`
}

// ExecutionTiming records timing information for an execution.
type ExecutionTiming struct {
	StartedAt   *string  `json:"started_at,omitempty"`
	CompletedAt *string  `json:"completed_at,omitempty"`
	DurationMS  *float64 `json:"duration_ms,omitempty"`

	// Legacy timing fields retained for older responses.
	TotalMS    int `json:"total_ms,omitempty"`
	PermitMS   int `json:"permit_ms,omitempty"`
	RoutingMS  int `json:"routing_ms,omitempty"`
	DispatchMS int `json:"dispatch_ms,omitempty"`
	ProviderMS int `json:"provider_ms,omitempty"`
}

func (t *ExecutionTiming) UnmarshalJSON(data []byte) error {
	type alias ExecutionTiming
	var item alias
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*t = ExecutionTiming(item)
	if t.TotalMS == 0 && t.DurationMS != nil {
		t.TotalMS = int(*t.DurationMS)
	}
	return nil
}

// ExecutionRoutingResult records which provider and model were selected.
type ExecutionRoutingResult struct {
	RequestedProvider *RoutingProvider `json:"requested_provider,omitempty"`
	RequestedModel    *string          `json:"requested_model,omitempty"`
	SelectedProvider  RoutingProvider  `json:"selected_provider"`
	SelectedModel     string           `json:"selected_model"`
	ReasonCode        string           `json:"reason_code"`
	FallbackOccurred  bool             `json:"fallback_occurred"`

	// Legacy aliases.
	Provider RoutingProvider `json:"provider,omitempty"`
	Model    string          `json:"model,omitempty"`
	Region   *string         `json:"region,omitempty"`
}

func (r *ExecutionRoutingResult) UnmarshalJSON(data []byte) error {
	type alias ExecutionRoutingResult
	var item alias
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*r = ExecutionRoutingResult(item)
	if r.SelectedProvider == "" {
		r.SelectedProvider = r.Provider
	}
	if r.Provider == "" {
		r.Provider = r.SelectedProvider
	}
	if r.SelectedModel == "" {
		r.SelectedModel = r.Model
	}
	if r.Model == "" {
		r.Model = r.SelectedModel
	}
	return nil
}

// ExecutionGovernance records governance details for an execution.
type ExecutionGovernance struct {
	Decision    Decision               `json:"decision"`
	Reason      string                 `json:"reason,omitempty"`
	Actions     []PermitActionResponse `json:"actions,omitempty"`
	Constraints map[string]any         `json:"constraints,omitempty"`
	Budgets     map[string]any         `json:"budgets,omitempty"`

	// Legacy fields.
	PermitID      string   `json:"permit_id,omitempty"`
	PolicyIDs     []string `json:"policy_ids,omitempty"`
	PolicyID      string   `json:"policy_id,omitempty"`
	PolicyVersion string   `json:"policy_version,omitempty"`
}

type ExecutionErrorPayload struct {
	Stage        *string        `json:"stage,omitempty"`
	Code         string         `json:"code"`
	Message      string         `json:"message"`
	Details      map[string]any `json:"details,omitempty"`
	RetryClass   *string        `json:"retry_class,omitempty"`
	StatusCode   *int           `json:"status_code,omitempty"`
	ProviderCode *string        `json:"provider_code,omitempty"`
}

type ExecutionResolved struct {
	Provider *string        `json:"provider,omitempty"`
	Model    *string        `json:"model,omitempty"`
	Alias    *string        `json:"alias,omitempty"`
	Policy   map[string]any `json:"policy,omitempty"`
	Health   map[string]any `json:"health,omitempty"`
	Budget   map[string]any `json:"budget,omitempty"`
}

// ExecutionStreamEvent represents a single SSE event during streaming execution.
type ExecutionStreamEvent struct {
	EventType string          `json:"event_type,omitempty"`
	Data      json.RawMessage `json:"data"`
}

// ExecuteRequest is the request body for the combined execute endpoint.
type ExecuteRequest struct {
	Provider *RoutingProvider `json:"provider,omitempty"`
	Model    string           `json:"model"`
	Input    map[string]any   `json:"input"`

	// Legacy fields retained for source compatibility with older SDK versions.
	// When Input is empty, they are folded into the canonical input object.
	Subject        Subject           `json:"-"`
	Action         Action            `json:"-"`
	Resource       Resource          `json:"-"`
	Context        *Context          `json:"-"`
	Conditions     *PermitConditions `json:"-"`
	Messages       []MessageInput    `json:"-"`
	Parameters     *ParametersInput  `json:"-"`
	CustomMetadata map[string]any    `json:"-"`
}

func (r ExecuteRequest) MarshalJSON() ([]byte, error) {
	provider := r.Provider
	if provider == nil && r.Resource.Attributes.Provider != "" {
		value := RoutingProvider(r.Resource.Attributes.Provider)
		provider = &value
	}

	model := r.Model
	if model == "" {
		model = r.Resource.Attributes.Model
	}

	input := r.Input
	if len(input) == 0 {
		input = map[string]any{}
		if len(r.Messages) > 0 {
			input["messages"] = r.Messages
		}
		if r.Parameters != nil {
			input["parameters"] = r.Parameters
		}
		if r.Subject.Type != "" || r.Subject.ID != "" || len(r.Subject.Attributes) > 0 {
			input["subject"] = r.Subject
		}
		if r.Action.Name != "" || len(r.Action.Attributes) > 0 {
			input["action"] = r.Action
		}
		if r.Resource.Type != "" || r.Resource.ID != "" {
			input["resource"] = r.Resource
		}
		if r.Context != nil {
			input["context"] = r.Context
		}
		if r.Conditions != nil {
			input["conditions"] = r.Conditions
		}
		if len(r.CustomMetadata) > 0 {
			input["custom_metadata"] = r.CustomMetadata
		}
	}

	type wire struct {
		Provider *RoutingProvider `json:"provider,omitempty"`
		Model    string           `json:"model"`
		Input    map[string]any   `json:"input"`
	}
	return json.Marshal(wire{
		Provider: provider,
		Model:    model,
		Input:    input,
	})
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
	JobID         string    `json:"job_id"`
	Status        JobStatus `json:"status"`
	ExecutionMode string    `json:"execution_mode"`
	SubmittedAt   string    `json:"submitted_at"`
	PermitID      *string   `json:"permit_id,omitempty"`

	// Deprecated alias for older responses.
	CreatedAt string `json:"created_at,omitempty"`
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
	EventType string         `json:"event_type"`
	Source    string         `json:"source"`
	Phase     string         `json:"phase"`
	Provider  *string        `json:"provider,omitempty"`
	Model     *string        `json:"model,omitempty"`
	PermitID  *string        `json:"permit_id,omitempty"`
	RequestID *string        `json:"request_id,omitempty"`
	JobID     *string        `json:"job_id,omitempty"`
	SessionID *string        `json:"session_id,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`

	// Deprecated alias for older SDK tests.
	Detail map[string]any `json:"detail,omitempty"`
}

// RequestTimelineResponse is the response from the request timeline endpoint.
type RequestTimelineResponse struct {
	ProjectID  string                 `json:"project_id"`
	RequestID  string                 `json:"request_id"`
	PermitIDs  []string               `json:"permit_ids"`
	JobIDs     []string               `json:"job_ids"`
	SessionIDs []string               `json:"session_ids"`
	Events     []RequestTimelineEvent `json:"events"`
}

// ExecutionTimelineResponse is the timeline response for an execution.
type ExecutionTimelineResponse = RequestTimelineResponse

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
	Cursor *string `json:"cursor,omitempty"`
	Status *string `json:"status,omitempty"`
	Scope  *string `json:"scope,omitempty"`

	// Deprecated: the API no longer accepts offset pagination.
	Offset *int `json:"-"`
}

// ApiKeyListResponse is the paginated list of API keys.
type ApiKeyListResponse struct {
	Items      []ApiKeyRecord `json:"items"`
	NextCursor *string        `json:"next_cursor,omitempty"`

	// Deprecated legacy response fields.
	Total  int `json:"total,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// --- Audit export / verifier types ---

type AuditExportRecord struct {
	Permit            map[string]any      `json:"permit"`
	PermitEvidence    []PermitEvidenceOut `json:"permit_evidence,omitempty"`
	LatestUsageLedger map[string]any      `json:"latest_usage_ledger,omitempty"`
	UsageLog          map[string]any      `json:"usage_log,omitempty"`
	ReplayEvidence    map[string]any      `json:"replay_evidence,omitempty"`
	ChainEntries      []map[string]any    `json:"chain_entries,omitempty"`
}

type AuditExportBundle struct {
	BundleType           string              `json:"bundle_type"`
	SchemaVersion        int                 `json:"schema_version"`
	Format               string              `json:"format"`
	ProjectID            string              `json:"project_id"`
	TimeRangeBasis       string              `json:"time_range_basis"`
	From                 string              `json:"from"`
	To                   string              `json:"to"`
	GeneratedAt          string              `json:"generated_at"`
	RecordCount          int                 `json:"record_count"`
	Records              []AuditExportRecord `json:"records"`
	MCPToolDecisionCount int                 `json:"mcp_tool_decision_count,omitempty"`
	MCPToolDecisions     []map[string]any    `json:"mcp_tool_decisions,omitempty"`
	ChainEntries         []map[string]any    `json:"chain_entries,omitempty"`
	IncludeChainEntries  bool                `json:"include_chain_entries,omitempty"`
}

func marshalWithExtra(base any, extra map[string]any) ([]byte, error) {
	data, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	if len(extra) == 0 {
		return data, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	for key, value := range extra {
		if key == "" || value == nil {
			continue
		}
		if _, exists := payload[key]; !exists {
			payload[key] = value
		}
	}
	return json.Marshal(payload)
}

func marshalInputDescriptorWithContent(base any, content map[string]any, knownFields map[string]bool) ([]byte, error) {
	data, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return data, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	metadata := map[string]any{}
	if existing, ok := payload["metadata"].(map[string]any); ok {
		for key, value := range existing {
			metadata[key] = value
		}
	}
	if contentMetadata, ok := content["metadata"].(map[string]any); ok {
		for key, value := range contentMetadata {
			if key != "" && value != nil {
				metadata[key] = value
			}
		}
	}

	for key, value := range content {
		if key == "" || value == nil || key == "metadata" {
			continue
		}
		if knownFields[key] {
			if _, exists := payload[key]; !exists {
				payload[key] = value
			}
			continue
		}
		metadata[key] = value
	}

	if len(metadata) > 0 {
		payload["metadata"] = metadata
	}
	return json.Marshal(payload)
}

// SSEEvent represents a parsed Server-Sent Event.
type SSEEvent struct {
	EventType string          `json:"event_type,omitempty"`
	Data      json.RawMessage `json:"data"`
}
