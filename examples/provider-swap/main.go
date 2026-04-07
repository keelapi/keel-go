package main

import (
	"context"
	"fmt"
	"log"

	"github.com/keelapi/keel-go/providers/anthropic"
	"github.com/keelapi/keel-go/providers/openai"
)

func main() {
	ctx := context.Background()

	// Use OpenAI through Keel — same interface as the official SDK
	oai := openai.NewClient(openai.Config{})
	resp, err := oai.Chat.Completions.Create(ctx, openai.ChatCompletionParams{
		Model:    "gpt-4",
		Messages: []openai.ChatCompletionMessage{{Role: "user", Content: "Hello from OpenAI via Keel"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("OpenAI: %s\n", resp.Choices[0].Message.Content)

	// Swap to Anthropic — just change the import and types
	ant := anthropic.NewClient(anthropic.Config{})
	msg, err := ant.Messages.Create(ctx, anthropic.MessageCreateParams{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		Messages:  []anthropic.MessageParam{{Role: "user", Content: "Hello from Anthropic via Keel"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Anthropic: %s\n", msg.Content[0].Text)
}
