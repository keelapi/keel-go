# keel-go

The official Go SDK for [Keel](https://keelapi.com) — a permit-first AI governance and execution control plane.

Keel sits between your application and AI providers (OpenAI, Anthropic, Google, xAI, Meta).

> **⚠️ Keel is currently in private beta.** You'll need a Keel account and API key to use this SDK.
> [Sign up for early access →](https://dashboard.keelapi.com/signup)

## Installation

```bash
go get github.com/keelapi/keel-go
```

**Zero external dependencies.** Only Go standard library.

## Quick Start

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
client.Permits.Export(ctx, from, to, format) // Export permits
client.Permits.ReportUsage(ctx, id, req)     // Report usage
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
// ... etc.
```

### Jobs, API Keys, Requests

```go
client.Jobs.Create(ctx, req)
client.Jobs.Get(ctx, jobID)

client.ApiKeys.Create(ctx, req)
client.ApiKeys.List(ctx, params)
client.ApiKeys.Revoke(ctx, keyID)

client.Requests.Timeline(ctx, requestID)
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
    var ke *keel.KeelError
    if errors.As(err, &ke) {
        fmt.Printf("Status: %d, Code: %s\n", ke.Status, ke.Code)
        if ke.IsRetryable() {
            // retry logic
        }
    }
}
```

## License

MIT
