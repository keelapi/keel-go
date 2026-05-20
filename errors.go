package keel

import (
	"fmt"
	"time"
)

// Reason code constants for permit denials and throttles (Shape D).
const (
	ReasonBudgetRequestCapExceeded       = "budget.request_cap_exceeded"
	ReasonBudgetDailyCapExceeded         = "budget.daily_cap_exceeded"
	ReasonBudgetMonthlyCapExceeded       = "budget.monthly_cap_exceeded"
	ReasonBudgetMonthlyThresholdExceeded = "budget.monthly_threshold_exceeded"
	ReasonBudgetDailySpikeDetected       = "budget.daily_spike_detected"
	ReasonBudgetRateLimitExceeded        = "budget.rate_limit_exceeded"
	ReasonBudgetRateLimitThrottled       = "budget.rate_limit_throttled"
	ReasonBudgetPricingUnavailable       = "budget.pricing_unavailable"
	ReasonPolicyModelNotAllowed          = "policy.model_not_allowed"
	ReasonPolicyRuleDenied               = "policy.rule_denied"
	ReasonPolicyReviewRequired           = "policy.review_required"
)

// KeelError represents an error response from the Keel API.
type KeelError struct {
	Status     int           `json:"status"`
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	Field      string        `json:"field,omitempty"`
	Details    any           `json:"details,omitempty"`
	RetryAfter time.Duration `json:"-"`
}

// Error implements the error interface.
func (e *KeelError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("keel: %d %s: %s (field: %s)", e.Status, e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("keel: %d %s: %s", e.Status, e.Code, e.Message)
}

// IsRetryable returns true if the error status code indicates the request can be retried.
func (e *KeelError) IsRetryable() bool {
	switch e.Status {
	case 408, 429, 500, 502, 503, 504:
		return true
	}
	return false
}

// ThrottledError is returned when a permit request is rate-limited (HTTP 429).
// It is raised after all retry attempts are exhausted.
type ThrottledError struct {
	PermitID          string        `json:"permit_id,omitempty"`
	ReasonCode        string        `json:"reason_code,omitempty"`
	RetryAfterSeconds int           `json:"retry_after_seconds"`
	RetryAfter        time.Duration `json:"-"`
	Message           string        `json:"message,omitempty"`
}

// Error implements the error interface.
func (e *ThrottledError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "rate limit throttled"
	}
	return fmt.Sprintf("keel: 429 throttled: %s (retry after %ds)", msg, e.RetryAfterSeconds)
}

// IsRetryable returns true. A throttled error is always retryable after the indicated delay.
func (e *ThrottledError) IsRetryable() bool {
	return true
}
