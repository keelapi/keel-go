package keel

import (
	"fmt"
	"time"
)

// KeelError represents an error response from the Keel API.
type KeelError struct {
	Status     int           `json:"status"`
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	Field      string        `json:"field,omitempty"`
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
