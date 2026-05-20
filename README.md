# keel-go

The official Go SDK for [Keel](https://keelapi.com) — a permit-first AI governance and execution control plane.

Keel sits between your application and AI providers (OpenAI, Anthropic, Google, xAI, Meta).

Keel is built and published by Keel API, Inc.

> **⚠️ Keel is currently in private beta.** You'll need a Keel account and API key to use this SDK.
> [Sign up for early access →](https://dashboard.keelapi.com/signup)

## Installation

```bash
go get github.com/keelapi/keel-go
```

**Zero external dependencies.** Only Go standard library.

## Quick Start — One-Line Provider Migration

Add Keel governance to existing AI code by swapping the import — no other code changes.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/keelapi/keel-go/providers/openai"
)

func main() {
    // Reads KEEL_BASE_URL, KEEL_API_KEY, KEEL_PROJECT_ID from the environment.
    client := openai.NewClient(openai.Config{})

    resp, err := client.Chat.Completions.Create(context.Background(), openai.ChatCompletionParams{
        Model:    "gpt-4o",
        Messages: []openai.ChatCompletionMessage{{Role: "user", Content: "Hello!"}},
    })
    if err != nil {
        log.Fatal(err) // a denied permit surfaces here as an error
    }
    fmt.Printf("%+v\n", resp)
}
```

Every call automatically requests a governance permit, executes through Keel — policy and budget enforced, audit recorded — and reports usage. Anthropic, Google, xAI, and Meta wrappers work the same way; see [Drop-in Provider Replacements](#drop-in-provider-replacements).

## Permit-First Mode

Use the Permit API directly when you want Keel to **decide** but not **execute** — your own code makes the AI call (with your own provider keys), and Keel never sees the prompt or the result. Keel evaluates policy and returns a permit; honoring it is up to your code.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    keel "github.com/keelapi/keel-go"
)

func main() {
    client := keel.NewClient(keel.ClientConfig{
        BaseURL: os.Getenv("KEEL_BASE_URL"),
        APIKey:  os.Getenv("KEEL_API_KEY"),
    })

    ctx := context.Background()

    // Create a permit
    permit, err := client.Permits.Create(ctx, keel.PermitRequest{
        ProjectID:      os.Getenv("KEEL_PROJECT_ID"),
        IdempotencyKey: "my-request-1",
        Subject:        keel.Subject{Type: "user", ID: "user-123"},
        Action:         keel.Action{Name: "generate.text"},
        Resource: keel.Resource{
            Type: "ai_model",
            ID:   "gpt-4",
            Attributes: keel.ResourceAttributes{
                Provider:  "openai",
                Model:     "gpt-4",
                Operation: keel.OpGenerateText,
                EstimatedInputTokens:  100,
                EstimatedOutputTokens: 200,
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Permit: %s → %s\n", permit.PermitID, permit.Decision)
}
```

## Request Lifecycle

Every request processed by Keel follows a consistent high-level flow:

- **Evaluate:** identity, policy, and budget constraints are checked
- **Decide:** a permit decision is issued — allow, deny, or constrain
- **Execute:** the provider call occurs only if permitted
- **Record:** usage, cost, and governance events are captured

Requests are only executed if explicitly permitted.

## Drop-in Provider Replacements

Swap your existing AI SDK imports with Keel-governed equivalents:

### OpenAI

```go
import "github.com/keelapi/keel-go/providers/openai"

client := openai.NewClient(openai.Config{})
resp, err := client.Chat.Completions.Create(ctx, openai.ChatCompletionParams{
    Model:    "gpt-4",
    Messages: []openai.ChatCompletionMessage{{Role: "user", Content: "Hello"}},
})
```

### Anthropic

```go
import "github.com/keelapi/keel-go/providers/anthropic"

client := anthropic.NewClient(anthropic.Config{})
msg, err := client.Messages.Create(ctx, anthropic.MessageCreateParams{
    Model:     "claude-sonnet-4-20250514",
    MaxTokens: 1024,
    Messages:  []anthropic.MessageParam{{Role: "user", Content: "Hello"}},
})
```

### Google Gemini

```go
import "github.com/keelapi/keel-go/providers/google"

client := google.NewClient(google.Config{})
model := client.GenerativeModel("gemini-pro")
resp, err := model.GenerateContent(ctx, google.GenerateContentRequest{
    Contents: []google.Content{{Parts: []google.Part{{Text: "Hello"}}}},
})
```

### xAI / Meta

```go
import "github.com/keelapi/keel-go/providers/xai"
import "github.com/keelapi/keel-go/providers/meta"

// Both use the OpenAI-compatible interface
xaiClient := xai.NewClient(xai.Config{})
metaClient := meta.NewClient(meta.Config{})
```

## Core API

### Permits

```go
client.Permits.Create(ctx, req)        // Create a permit
client.Permits.DryRun(ctx, req)        // Evaluate without creating
client.Permits.List(ctx, params)       // List permits
client.Permits.Get(ctx, permitID)      // Get a permit
client.Permits.Export(ctx, from, to, "json")  // Legacy raw export wrapper
client.Permits.ExportBundle(ctx, from, to)    // JSON audit export bundle
client.Permits.ExportCSV(ctx, from, to)       // CSV audit export text
client.Permits.ReportUsage(ctx, id, req)     // Report usage
client.Permits.VerifyUsage(ctx, id, req)     // Verify usage
client.Permits.Attest(ctx, id, req)          // Attest a permit
client.Permits.AddEvidence(ctx, id, req)     // Add evidence
client.Permits.ListEvidence(ctx, id)         // List evidence
client.Permits.Lineage(ctx, id)              // Get lineage
client.Permits.Bundle(ctx, id)               // Full audit bundle
```

### Executions

```go
client.Executions.Create(ctx, req)     // Sync execution
events, errc := client.Executions.Stream(ctx, req)  // Streaming
client.Executions.Get(ctx, requestID)  // Get timeline
```

### Execute (Combined)

```go
client.Execute.Run(ctx, req)  // Permit + execute in one call
```

### Proxy

```go
client.Proxy.OpenAI(ctx, payload)
client.Proxy.Anthropic(ctx, payload)
client.Proxy.Google(ctx, payload)
client.Proxy.XAI(ctx, payload)
client.Proxy.Meta(ctx, payload)

// Streaming variants
client.Proxy.OpenAIStream(ctx, payload)
client.Proxy.AnthropicStream(ctx, payload)
client.Proxy.GoogleStream(ctx, payload)
client.Proxy.XAIStream(ctx, payload)
client.Proxy.MetaStream(ctx, payload)
```

### Jobs, API Keys, Requests

```go
client.Jobs.Create(ctx, req)
client.Jobs.Get(ctx, jobID)

client.ApiKeys.Create(ctx, req)
client.ApiKeys.List(ctx, params)
client.ApiKeys.Revoke(ctx, keyID)
revoked, err := client.ApiKeys.RevokeRecord(ctx, keyID)

client.Requests.Timeline(ctx, requestID)
```

### Compliance And Integrity

```go
client.Compliance.CreateExport(ctx, req, true) // Signed export job with chain entries
client.Compliance.ListExports(ctx)
client.Compliance.GetExport(ctx, exportID)
client.Compliance.Keys(ctx)                    // Public verifier key manifest

client.Integrity.CheckpointPublicKey(ctx)
client.Integrity.PermitBindingPublicKeys(ctx)
```

## Configuration

```go
client := keel.NewClient(keel.ClientConfig{
    BaseURL:          "https://api.keelapi.com",
    APIKey:           "your-api-key",
    Timeout:          30 * time.Second,
    RequestFreshness: true, // Adds X-Keel-Timestamp + X-Keel-Nonce
    RetryConfig: &keel.RetryConfig{
        MaxRetries:    3,
        InitialDelay:  500 * time.Millisecond,
        MaxDelay:      30 * time.Second,
    },
})
```

## Environment Variables

| Variable | Description |
|---|---|
| `KEEL_BASE_URL` | Keel API base URL |
| `KEEL_API_KEY` | Bearer token |
| `KEEL_PROJECT_ID` | Project ID (used by provider wrappers) |

## Running the Examples

1. Copy the env template and fill in your credentials:

```bash
cp examples/.env.example .env
# Edit .env with your values from https://keelapi.com
```

2. Source your env and run any example:

```bash
source .env && export KEEL_BASE_URL KEEL_API_KEY KEEL_PROJECT_ID
go run ./examples/quickstart
go run ./examples/provider-swap
go run ./examples/end-to-end
```

## Error Handling

```go
resp, err := client.Permits.Create(ctx, req)
if err != nil {
    var te *keel.ThrottledError
    if errors.As(err, &te) {
        // Rate-limited (HTTP 429). Retry after the indicated delay.
        fmt.Printf("Throttled: retry after %ds (reason: %s)\n",
            te.RetryAfterSeconds, te.ReasonCode)
        return
    }
    var ke *keel.KeelError
    if errors.As(err, &ke) {
        fmt.Printf("Status: %d, Code: %s\n", ke.Status, ke.Code)
    }
}
```

## Rate Limiting and Retries

The SDK automatically retries failed requests with exponential backoff. By default, status codes 408, 429, 500, 502, 503, and 504 are retried up to 3 times.

When the API returns HTTP 429 (rate limit throttled), the SDK respects the `Retry-After` header and waits the indicated duration before retrying. If the header is absent, it falls back to `retry_after_seconds` from the response body.

After all retries are exhausted, a 429 response returns a `*ThrottledError` (not a `*KeelError`). This typed error exposes `RetryAfterSeconds`, `PermitID`, and `ReasonCode` so callers can implement custom backoff or surface details to end users.

Configure retry behavior via `RetryConfig`:

```go
client := keel.NewClient(keel.ClientConfig{
    BaseURL: "https://api.keelapi.com",
    APIKey:  "your-api-key",
    RetryConfig: &keel.RetryConfig{
        MaxRetries:        2,              // default: 3
        InitialDelay:      time.Second,    // default: 500ms
        MaxDelay:          15 * time.Second, // default: 30s
        BackoffMultiplier: 2.0,            // default: 2.0
    },
})
```

Non-retryable errors (400, 401, 403, 404) are returned immediately without retry.

## License

MIT
