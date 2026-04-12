# Changelog

## Unreleased

### Added

- **`ThrottledError` type**: HTTP 429 responses now return a `*ThrottledError` instead of a generic `*KeelError`. The new error type exposes `RetryAfterSeconds`, `PermitID`, `ReasonCode`, and `Message` for programmatic handling of rate-limit throttles.
- **`DecisionThrottled` constant**: New `Decision` value `"throttled"` for permit responses that indicate a rate-limit throttle rather than a hard deny.
- **Shape D reason code constants**: 11 semantic reason code constants (`ReasonBudgetDailyCapExceeded`, `ReasonPolicyModelNotAllowed`, etc.) for programmatic matching against permit denial and throttle reason codes.
- **New permit response fields**: `ReasonCode`, `ReasonDetail`, `OutcomeDetail`, and `Message` fields on `PermitResponse`, `PermitDryRunResponse`, and `PermitAuditItem` to support the V1.16.0 policy engine response shape.
- **Rate limiting and retries documentation**: README now documents retry behavior, `Retry-After` handling, `ThrottledError` usage, and `RetryConfig` options.

### Changed

- **429 body parsing**: The HTTP client now parses the 429 response body for throttle-specific fields (`permit.permit_id`, `permit.reason_code`, `permit.message`, `permit.outcome_detail.retry_after_seconds`). The `Retry-After` header takes precedence; body values are used as fallback.
- **`TestErrorParsing`**: Updated to expect `ThrottledError` for 429 responses (was `KeelError`).

### Notes

- `Budgets` and `Constraints` remain `map[string]any` (opaque). No typed model changes needed — `schema_version` passes through naturally.
- All legacy reason codes continue to work. Shape D codes are additive.
