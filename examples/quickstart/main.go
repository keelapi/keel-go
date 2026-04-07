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
		IdempotencyKey: "quickstart-1",
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
	fmt.Printf("Permit: %s → %s\n", permit.PermitID, permit.Decision)

	if permit.Decision != keel.DecisionAllow {
		log.Fatal("permit denied")
	}

	// Execute with the permit
	result, err := client.Executions.Create(ctx, keel.ExecutionCreateRequest{
		Operation: string(keel.OpGenerateText),
		Mode:      keel.ModeSync,
		Messages:  []keel.MessageInput{{Role: "user", Content: "What is Keel?"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result: %s\n", result.Status)
}
