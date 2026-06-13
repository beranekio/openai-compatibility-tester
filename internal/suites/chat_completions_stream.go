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
		Store: openai.Bool(false),
	})
	defer stream.Close()

	chunks := 0
	var hasOutput bool
	var finished bool
	var finishReason string
	for stream.Next() {
		chunk := stream.Current()
		chunks++
		if chunk.ID == "" {
			return fail("chat_completions_stream", "stream chunk missing id")
		}
		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]
			if choice.FinishReason != "" {
				finished = true
				finishReason = choice.FinishReason
			}
			if choice.Delta.Content != "" || choice.Delta.Refusal != "" {
				hasOutput = true
			}
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat completion stream failed: %w", err)
	}
	if chunks == 0 {
		return fail("chat_completions_stream", "stream returned no chunks")
	}
	if !finished {
		return fail("chat_completions_stream", "stream missing terminal finish_reason")
	}
	if !hasOutput && !isContentFilterFinishReason(finishReason) {
		return fail("chat_completions_stream", "stream produced no text content or refusal")
	}
	return nil
}