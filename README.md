# keel-go

## Surface position

> The OpenAPI specification is the canonical integration contract for all Keel surfaces.
>
> **First-class runtime SDKs:** Python and TypeScript. Release-gated and kept in semantic lockstep with the runtime.
>
> **Infrastructure surfaces:** Terraform is the official policy-as-code surface. MCP governance is exposed through `/v1/mcp/*`. Keel should not be described as a generic MCP server or submitted to MCP registries.
>
> **Generated/reference client:** Go is published as an official generated/reference client for infrastructure teams. It is not a first-class runtime SDK.
>
> **Other languages:** Clients can be generated from the OpenAPI specification. They are not maintained as official Keel SDKs.

The official generated/reference client for [Keel](https://keelapi.com).

Keel is built and published by Keel API, Inc.

> **Keel is live with self-serve access.** Create an account and project API key in the dashboard. Production deployment assistance is available from Keel.
> [Get started](https://dashboard.keelapi.com/signup)

## Installation

```bash
go get github.com/keelapi/keel-go
```

## Package

The generated OpenAPI client lives in:

```go
import keelclient "github.com/keelapi/keel-go/pkg/keelclient"
```

The repository root package is documentation-only. Runtime API calls should use `pkg/keelclient`.

## Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"

    keelclient "github.com/keelapi/keel-go/pkg/keelclient"
)

func main() {
    client, err := keelclient.NewClient(
        os.Getenv("KEEL_BASE_URL"),
        keelclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
            req.Header.Set("Authorization", "Bearer "+os.Getenv("KEEL_API_KEY"))
            return nil
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.ExecuteV1ExecutePost(
        context.Background(),
        keelclient.UnifiedExecuteRequest{
            Model: "gpt-4o-mini",
            Input: map[string]interface{}{
                "messages": []map[string]string{
                    {"role": "user", "content": "Reply with exactly: keel works"},
                },
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    fmt.Println(resp.Status)
}
```

The generated client returns standard `*http.Response` values. Callers own status handling, response decoding, retry policy, and transport configuration.

## Regeneration

The client is generated from `api/openapi.json`, a structural copy of Keel's OpenAPI document (`docs/public-artifacts/openapi.json` in the keel-api repository). The vendored copy omits descriptions, summaries, examples, tags, contact details, terms and the API license metadata; it retains paths, parameters, responses and schemas required to generate the client. The API's commercial license remains distinct from this repository's MIT license. `make generate` regenerates the client from the structural copy, and `make check-generated`, which CI runs, fails if the committed client does not match it. With a keel-api checkout next to this repository, update the copy and regenerate with:

```bash
make regenerate
```

`make regenerate` redacts the spec committed at the checkout's `HEAD` into `api/openapi.json`, regenerates from it, and records both the original source SHA-256 and the redacted spec SHA-256 in `pkg/keelclient/SPEC_SOURCE`; the recorded commit changes only when the source spec does. Use `KEEL_API_DIR=/path/to/keel-api make regenerate` for a checkout elsewhere. To try another spec file without recording a source:

```bash
OPENAPI_SPEC=/path/to/openapi.json make generate
```

The `Regenerate OpenAPI client` workflow runs `make regenerate` weekly against keel-api's default branch, checks that the vendored spec remains redacted, runs the build, vet, race tests, gofmt and README checks on the result, and opens a pull request when the generated output changes. It needs the two repository secrets described in `.github/workflows/regenerate.yml`: without the keel-api read token it skips with a notice, and without the pull request token it regenerates and verifies but opens no pull request. The generated code still exposes the API's structural operations and schemas; the redaction only removes the extra prose and metadata from the vendored document.

Generation writes `pkg/keelclient/client.gen.go`. The Makefile first runs a local normalizer (`cmd/normalize-openapi`) that adapts the OpenAPI 3.1 document for oapi-codegen's OpenAPI 3.0 parser without modifying the source specification.

## Generated Surface

`pkg/keelclient/client.gen.go` is generated and is the package's entire API; the only hand-written Go file in the package is a test. `pkg/keelclient/SPEC_SOURCE` records what the client was generated from: the keel-api commit, the SHA-256 of the source and redacted OpenAPI documents, the oapi-codegen version, and the generator settings.

This repository does not maintain provider-specific wrappers, permit-chain helpers, streaming helpers, custom retry/backoff, or custom error types. Any language-specific surface beyond generated OpenAPI bindings is out of scope for this Tier 4 client.

## License

MIT
