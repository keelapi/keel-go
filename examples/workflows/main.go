package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	keel "github.com/keelapi/keel-go"
)

func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func main() {
	client := keel.NewClient(keel.ClientConfig{
		BaseURL: os.Getenv("KEEL_BASE_URL"),
		APIKey:  os.Getenv("KEEL_API_KEY"),
	})

	ctx := context.Background()
	workflowID := fmt.Sprintf("invoice-batch-%d", time.Now().Unix())

	declaration, err := client.Workflows.Declare(ctx, keel.WorkflowDeclareRequest{
		WorkflowID: workflowID,
		Intent: keel.WorkflowIntent{
			ExpectedCalls:               intPtr(2),
			MaxCalls:                    intPtr(4),
			ExpectedModel:               stringPtr("gpt-4o-mini"),
			ExpectedInputTokensPerCall:  intPtr(500),
			ExpectedOutputTokensPerCall: intPtr(200),
			MaxDurationSeconds:          intPtr(3600),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Workflow %s declared: %s\n", declaration.WorkflowID, declaration.Decision)

	err = keel.RunInWorkflow(ctx, workflowID, func(ctx context.Context) error {
		for i := 0; i < 2; i++ {
			permit, err := client.Permits.Create(ctx, keel.PermitRequest{
				ProjectID:      os.Getenv("KEEL_PROJECT_ID"),
				IdempotencyKey: fmt.Sprintf("%s-permit-%d", workflowID, i),
				Subject:        keel.Subject{Type: "service", ID: "invoice-worker"},
				Action:         keel.Action{Name: string(keel.OpGenerateText)},
				Resource: keel.Resource{
					Type: "ai_model",
					ID:   "gpt-4o-mini",
					Attributes: keel.ResourceAttributes{
						Provider:              string(keel.ProviderOpenAI),
						Model:                 "gpt-4o-mini",
						Operation:             keel.OpGenerateText,
						EstimatedInputTokens:  500,
						EstimatedOutputTokens: 200,
					},
				},
			})
			if err != nil {
				return err
			}
			fmt.Printf("Permit %s: %s\n", permit.PermitID, permit.Decision)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	version := 1
	if declaration.Version != nil {
		version = *declaration.Version
	}
	amended, err := client.Workflows.Amend(ctx, workflowID, keel.WorkflowAmendRequest{
		IfMatchVersion: version,
		NewMaxCalls:    intPtr(6),
		ReasonProvided: stringPtr("invoice volume increased"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Workflow %s amended to version %d\n", amended.WorkflowID, valueOrZero(amended.Version))

	completed, err := client.Workflows.Complete(ctx, workflowID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Workflow %s completed with %d calls\n", completed.WorkflowID, completed.AuthoritativeActualCalls)
}

func valueOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
