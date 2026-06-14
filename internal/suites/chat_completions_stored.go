package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

const storedChatCompletionInput = "Reply with exactly the word: pong"

func createStoredChatCompletion(ctx context.Context, client openai.Client, cfg *config.Config) (*openai.ChatCompletion, error) {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(storedChatCompletionInput),
		},
		Store: openai.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("create stored chat completion failed: %w", err)
	}
	if resp == nil {
		return nil, fail("chat_completions", "create stored chat completion returned nil")
	}
	if resp.ID == "" {
		return nil, fail("chat_completions", "create stored chat completion missing id")
	}
	return resp, nil
}