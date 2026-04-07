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
		BaseURL:          os.Getenv("KEEL_BASE_URL"),
		APIKey:           os.Getenv("KEEL_API_KEY"),
		RequestFreshness: true,
	})

	ctx := context.Background()

	// 1. Dry-run to check if request would be allowed
	dryRun, err := client.Permits.DryRun(ctx, keel.PermitRequest{
		ProjectID:      os.Getenv("KEEL_PROJECT_ID"),
		IdempotencyKey: "dry-run-1",
		Subject:        keel.Subject{Type: "user", ID: "user-123"},
		Action:         keel.Action{Name: string(keel.OpGenerateText)},
		Resource: keel.Resource{
			Type: "ai_model",
			ID:   "gpt-4",
			Attributes: keel.ResourceAttributes{
				Provider:              "openai",
				Model:                 "gpt-4",
				Operation:             keel.OpGenerateText,
				EstimatedInputTokens:  100,
				EstimatedOutputTokens: 200,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Dry run: %s\n", dryRun.Decision)

	// 2. Use the combined execute endpoint (permit + execution in one call)
	result, err := client.Execute.Run(ctx, keel.ExecuteRequest{
		Subject: keel.Subject{Type: "user", ID: "user-123"},
		Action:  keel.Action{Name: string(keel.OpGenerateText)},
		Resource: keel.Resource{
			Type: "ai_model",
			ID:   "gpt-4",
			Attributes: keel.ResourceAttributes{
				Provider:              "openai",
				Model:                 "gpt-4",
				Operation:             keel.OpGenerateText,
				EstimatedInputTokens:  100,
				EstimatedOutputTokens: 500,
			},
		},
		Messages: []keel.MessageInput{{Role: "user", Content: "Explain AI governance"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Execution: %s (provider: %s)\n", result.RequestID, result.Provider)

	// 3. Get the timeline for the request
	timeline, err := client.Requests.Timeline(ctx, result.RequestID)
	if err != nil {
		log.Fatal(err)
	}
	for _, event := range timeline.Events {
		fmt.Printf("  [%s] %s\n", event.Timestamp, event.Phase)
	}

	// 4. Get the full audit bundle
	bundle, err := client.Permits.Bundle(ctx, result.PermitID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bundle: permit=%s, events=%d\n", bundle.Permit.PermitID, len(bundle.Timeline))
}
