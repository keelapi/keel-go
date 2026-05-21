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

> **Note:** Keel is currently in private beta. You'll need a Keel account and API key to use this client.
> [Sign up for early access](https://dashboard.keelapi.com/signup)

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
    client, err := keelclient.NewKeelHTTPClient(
        os.Getenv("KEEL_BASE_URL"),
        keelclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
            req.Header.Set("Authorization", "Bearer "+os.Getenv("KEEL_API_KEY"))
            return nil
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.CreatePermitV1PermitsPost(
        context.Background(),
        nil,
        map[string]interface{}{
            "project_id":      os.Getenv("KEEL_PROJECT_ID"),
            "idempotency_key": "example-1",
            "subject": map[string]interface{}{
                "type": "user",
                "id":   "user-123",
            },
            "action": map[string]interface{}{
                "name": "generate.text",
            },
            "resource": map[string]interface{}{
                "type": "ai_model",
                "id":   "gpt-4",
                "attributes": map[string]interface{}{
                    "provider":                "openai",
                    "model":                   "gpt-4",
                    "operation":               "generate.text",
                    "estimated_input_tokens":  100,
                    "estimated_output_tokens": 200,
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

The canonical OpenAPI source lives in `../keel-api/docs/public-artifacts/openapi.json`.

Regenerate the client with:

```bash
make generate
```

To point at a different local spec path:

```bash
OPENAPI_SPEC=/path/to/openapi.json make generate
```

Generation writes `pkg/keelclient/client.gen.go`. The Makefile runs a local normalizer that adapts the OpenAPI 3.1 document for the current `oapi-codegen` parser without modifying the source specification.

## Generated Surface

This repository does not maintain provider-specific wrappers, permit-chain helpers, streaming helpers, custom retry/backoff, or custom error types. Any language-specific surface beyond generated OpenAPI bindings is out of scope for this Tier 4 client.

## License

MIT
