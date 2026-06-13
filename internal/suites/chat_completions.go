package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletions verifies POST /v1/chat/completions via client.Chat.Completions.New.
type ChatCompletions struct{}

func (ChatCompletions) Name() string { return "chat_completions" }
func (ChatCompletions) Description() string {
	return "Chat completion (POST /v1/chat/completions)"
}

func (ChatCompletions) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Reply with exactly the word: pong"),
		},
	})
	if err != nil {
		return fmt.Errorf("chat completion request failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions", "response missing choices")
	}
	if resp.Choices[0].Message.Content == "" {
		return fail("chat_completions", "choice message content is empty")
	}
	if resp.ID == "" {
		return fail("chat_completions", "response missing id")
	}
	return nil
}