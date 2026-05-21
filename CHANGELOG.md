# Changelog

## Unreleased

### Added

- **Generated OpenAPI client**: `pkg/keelclient/client.gen.go` is generated from the canonical Keel OpenAPI document with `oapi-codegen`.
- **Regeneration workflow**: `make generate` regenerates the client from `../keel-api/docs/public-artifacts/openapi.json`; CI can open a regeneration PR when output changes.
- **Surface positioning**: README now documents Go as an official generated/reference client, not a first-class runtime surface.

### Changed

- **Runtime surface**: Removed the hand-maintained client implementation in favor of the generated OpenAPI client.

### Removed

- Provider-specific wrappers, provider-swap examples, streaming helpers, retry/backoff behavior, custom error types, compatibility shims, and legacy convenience methods.
