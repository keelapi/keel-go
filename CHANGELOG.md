# Changelog

keel-go follows Go module versioning. While the major version is 0, a minor release (v0.x.0) can contain breaking changes.

## v0.4.0 (2026-09-23)

Regenerates the client from Keel's current OpenAPI document and fixes the generator pipeline. **This release breaks code written against v0.3.x**; see [Migrating from v0.3.x](#migrating-from-v03x).

The client is generated from `docs/public-artifacts/openapi.json` in keelapi/keel-api at `bb391aecd0896185c069a35904cecfa0d7d6aed9` (SHA-256 `84aeabc1c5013023dac5188ad411ad64f97145d50eec7815c723aeb5f0d92ad7`) with oapi-codegen v2.7.0, as recorded in `pkg/keelclient/SPEC_SOURCE`. v0.3.x was generated from the document as of 2026-05-20.

### Added

- **451 operations**, up from 289. New runtime operations include:
  - Permits: `POST /v1/permits/{permit_id}/revoke`, `/close`, `/co-signature`, `/counter-signature`, `/operator-approval` and `/audit-attestation`; `GET /v1/permits/{permit_id}/timeline`, `/receipt`, `/proof-pack`, `/artifact`, `/artifact/download` and `/mcp-tool-calls`; `/v1/cost-permits`; `/v1/permit-v2/*-keys`.
  - Execution and work: `POST /v1/execute/{execution_id}/resume`, `/v1/executions/{execution_id}/provider-attestation`, `/v1/executions/{execution_id}/reconcile`, `/v1/work/actions` and `/v1/work/actions/{permit_id}/resume`.
  - MCP: `GET /v1/mcp/decision-records/{decision_id}`, connector and tool contracts, and tool-contract discovery.
  - Verification: `GET /v1/verification-profile`, compliance export proof packs, and integrity checkpoint consistency and lookup.
  - Sessions and voice (`/v1/sessions`, `/v1/voice`), authority grants and graph (`/v1/authority`), `POST /v1/agents/spawn`, `/v1/agent-credentials`, `/v1/principals`, `GET /v1/identity/whoami` and `GET /v1/authoring-context`.
- **Typed refusals.** Permit create and dry-run parse 403 into `JSON403 *PermitStructuralRefusalResponse`. MCP `:call` parses 200 into `*McpCallPermittedResponse`, 403 into a union of `McpCallActionReviewResponse`, `McpCallTerminalDenyResponse` and `McpCallPreconditionRefusalResponse`, 409 into `*McpCallStructuralHoldResponse` and 429 into `*McpCallThrottledResponse`.
- **New fields on existing types**, for example `PermitResponse.IssuedAt`, `ExpiresAt`, `Record` and `GovernanceRecordId`, and `UnifiedExecuteRequest.ActionVerb` and `BudgetEnvelopeId`. v0.3.x silently dropped these on decode.
- **Published schemas no operation references** (for example `AuditExportRecord`, `ChainEntry`, `ReplayEvidenceSummary` and `PermitCoSignatureEvidenceV2`) are generated instead of pruned.

### Changed (breaking)

- **Optional values are plain pointers.** Optional or nullable fields that were union wrapper types (for example `*PermitRequest_ParentPermitId`, read with `AsPermitRequestParentPermitId0()`) are now `*T` (here `*openapi_types.UUID`). The wrapper types, their numbered branch types and their `As…`, `From…` and `Merge…` methods are removed. A field that is required but may be `null` and refers to another schema, such as `PermitResponse.Record`, is a plain struct, and `null` decodes to its zero value. Fields whose value can be more than one non-null type, such as `ResourceAttributes.Operation`, are still unions.
- **Optional parameters can be set.** Optional header and query parameters such as `X-API-Key` were anonymous `*struct{ union json.RawMessage }` values that code outside the package could not construct. They are now plain pointers such as `*string`.
- **Enum constants are always named `<TypeName><Value>`.** v0.3.x used bare names such as `Agent` while they were unique. Examples: `Agent` is now `IdentitySubjectTypeAgent`, `Text` is `PermitInputTypeText`, `Pending` is `PermitUsageVerificationSummaryStatusPending`, `AttestationDrift` is `ReplayResultDriftTypeAttestationDrift`, and `ResourceAttributesOperationGenerateText` is `ResourceAttributesOperation0GenerateText`. Enum types that were branches of a nullable union lose their numeric suffix: `IdentitySubjectType0` is now `IdentitySubjectType`.
- **Decisions are `allow`, `deny`, `review` and `throttle`.** The API no longer returns `challenge` in permit, envelope, dry-run and replay decision fields or accepts it as a permit list filter. The 19 `…Challenge` constants on those enums and the `Challenge` counts on `PermitDecisionCounts` and `PermitDecisionByProjectItem` are removed. `review` means approval is required; `throttle` means retry later. MCP decision records (`MCPDecisionRecord.Decision`) still use `allow`, `deny` and `challenge`.
- **`PermitDryRunResponse` and `PermitAuditBundle` are `map[string]interface{}`**, because the OpenAPI document now publishes both as untyped objects. Their typed fields and enums (`PermitDryRunResponseDecision`, `PermitAuditBundleFormat` and others) are removed. `GET /v1/permits/{permit_id}/bundle` returns a union: use `AsPermitAuditBundle`, `AsWorkChainPackV1`, `AsWorkChainPackV2` or `AsSelfAttestingEvidenceBundle`.
- **MCP `:call` and `:prepare` take no `params` argument.** `POST /v1/mcp/servers/{server_ref}/tools/{tool_name}:call` and `:prepare` no longer accept an `X-API-Key` header, so the `PostMcpToolCall…Params` and `PostMcpToolPrepare…Params` types are removed. Both authenticate only with `Authorization: Bearer <client API key bound to a live agent principal>`. The API answers 401 `unauthorized` without a Bearer key, 403 `insufficient_scope` for an admin key and 403 `execution_principal_required` for a client key that is not bound to a live agent principal.
- **MCP request and decision fields are typed.** `RequestedAction` on `McpDecisionRequest`, `McpExecutionCallRequest` and `McpExecutionPrepareRequest` is an enum; the only value `:call` and `:prepare` accept is `mcp.tool.call`. `MatchedPolicyIds` must be null or empty. `Decision` on `McpDecisionResponse`, `MCPServerDecisionItem`, `RequestInspectorDecisionInfo` and `SessionTrailRequestItem` is an enum type instead of `string`.
- **Other type changes:**
  - `LatestCheckpointResponse.PublishedAt` is `*time.Time`.
  - `ListProjectPermits…Response.JSON200` is a union of `PermitsListResponse` and `UnifiedPermitEnvelopeListResponse`.
  - `WorkflowCompleteResponse`, `WorkflowDashboardDetailResponse`, `WorkflowDeclarationResponse` and `WorkflowListResponse` are structs instead of maps.
  - Request body types such as `CreatePermitV1PermitsPostJSONBody` are defined types instead of aliases.

### Removed

- The 12 dashboard Drata and Vanta integration operations and their types. They are no longer part of the API.
- Hand-written code in `pkg/keelclient`, which was merged to main after v0.3.1 but never released: `permit_lifecycle.gen.go` and hand edits inside `client.gen.go` (#4). The generator now emits the timeline operation with the same names and field types, and adds `PermitLifecycleResponse.Record` and `RecordChildren`.

### Fixed

- `make generate` against the current OpenAPI document produced a client that did not compile.
- The normalizer collapsed only one of the document's 2,417 nullable unions. Every other optional value became a union wrapper, including parameters that callers could not set.
- The README example calls `keelclient.NewClient` and leads with managed execution (`POST /v1/execute`) (#5). The v0.3.1 README called `keelclient.NewKeelHTTPClient`, which does not exist.

### Tooling

- `pkg/keelclient/SPEC_SOURCE` records the keel-api commit, the OpenAPI document's SHA-256, the oapi-codegen version and the generator settings. `make regenerate` regenerates from a keel-api checkout and updates it.
- `api/openapi.json` is a prose-redacted structural copy of the OpenAPI document the client is generated from. `SPEC_SOURCE` records both its SHA-256 and the source document's SHA-256. `make check-generated`, which CI runs, checks the redaction, regenerates from the vendored copy and fails if `client.gen.go` or `SPEC_SOURCE` differs from the result.
- oapi-codegen settings: `output-options.skip-prune` and `compatibility.always-prefix-enum-values`.
- The normalizer converts numeric `exclusiveMinimum` and `exclusiveMaximum` for the OpenAPI 3.0 parser (#4).
- CI compiles the README example (`make check-readme`).
- The regeneration workflow has not run successfully since it was added: the default `GITHUB_TOKEN` cannot read the private keel-api repository. It now reads the spec with a `KEEL_API_READ_TOKEN` secret and skips with a notice without it. It verifies the regenerated client before proposing it, and opens the pull request with a `KEEL_GO_PR_TOKEN` secret so that CI runs on it.

### Migrating from v0.3.x

1. Run `go get github.com/keelapi/keel-go@v0.4.0`, then `go build ./...`. The compiler lists every renamed or removed identifier you use.
2. Replace union accessors with pointer checks: `v.AsPermitRequestParentPermitId0()` becomes `*v` after a nil check, and `From…` setters become plain assignments of a pointer.
3. Rename enum constants to `<TypeName><Value>`. The type name is the one in the compile error.
4. Handle `review` and `throttle` wherever you handled `challenge`.
5. Read dry-run results and audit bundles from the map, or unmarshal `resp.Body` into your own struct.
6. For MCP `:call` and `:prepare`, drop the `params` argument and send an agent-bound client key as a Bearer token, for example with `keelclient.WithRequestEditorFn`.

## v0.3.1 (2026-05-21)

The same client as v0.3.0, plus CI (build, race tests, gofmt and vet on the Go version in `go.mod`).

- Reconciled main, where workflow_intent helpers had been merged after v0.2.1 (#1, #2, #3), by removing them under the Tier 4 doctrine. They were never in a tagged release.
- The README example calls `keelclient.NewKeelHTTPClient`, which does not exist. Use `keelclient.NewClient`.
- The release notes said CI opens a regeneration pull request when the OpenAPI spec changes. The workflow could not run; see v0.4.0.

## v0.3.0 (2026-05-21)

**Breaking.** keel-go becomes the official generated/reference client (Tier 4); the first-class runtime SDKs are Python and TypeScript.

- The API is the oapi-codegen client in `pkg/keelclient/client.gen.go`, generated from Keel's OpenAPI document as of 2026-05-20 (289 operations). The root `keel` package is documentation-only.
- `make generate` and the `cmd/normalize-openapi` normalizer regenerate the client, and a scheduled workflow was added to open regeneration pull requests.
- Removed the hand-maintained client: the root-package clients, provider wrappers (`providers/openai`, `anthropic`, `google`, `xai`, `meta`), streaming helpers, retry and backoff, custom error types, compatibility shims and examples. Hand-maintained features merged after v0.2.1 but never tagged, including `ThrottledError` and the reason-code constants, were removed with them.

## v0.2.1 (2026-04-07)

- README: removed the request pipeline line.

## v0.2.0 (2026-04-07)

- Tagged on the same commit as v0.1.0.

## v0.1.0 (2026-04-07)

Initial release: a hand-written Go SDK with no dependencies outside the standard library.

- `keel.NewClient` with clients for permits (create, dry-run, list, attest, evidence, lineage), executions (including streaming), execute, proxy, jobs, API keys and request timelines, with retries and typed errors.
- Drop-in provider wrappers for OpenAI, Anthropic, Google, xAI and Meta, and examples.
