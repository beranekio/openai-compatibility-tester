package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsStream verifies streaming chat completions.
type ChatCompletionsStream struct{}

func (ChatCompletionsStream) Name() string { return "chat_completions_stream" }
func (ChatCompletionsStream) Description() string {
	return "Streaming chat completion (POST /v1/chat/completions, stream=true)"
}

func (ChatCompletionsStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Count from one to three."),
		},
		MaxTokens: openai.Int(32),
	})

	chunks := 0
	var content string
	for stream.Next() {
		chunk := stream.Current()
		chunks++
		if len(chunk.Choices) > 0 {
			content += chunk.Choices[0].Delta.Content
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat completion stream failed: %w", err)
	}
	if chunks == 0 {
		return fail("chat_completions_stream", "stream returned no chunks")
	}
	if content == "" {
		return fail("chat_completions_stream", "stream produced no text content")
	}
	return nil
}