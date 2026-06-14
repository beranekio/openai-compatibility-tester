package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsLogprobs verifies chat completions with log probability output.
type ChatCompletionsLogprobs struct{}

func (ChatCompletionsLogprobs) Name() string { return "chat_completions_logprobs" }
func (ChatCompletionsLogprobs) Description() string {
	return "Chat completion with logprobs (POST /v1/chat/completions)"
}

func (ChatCompletionsLogprobs) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Reply with exactly the word: pong"),
		},
		Store:       openai.Bool(false),
		Logprobs:    openai.Bool(true),
		TopLogprobs: openai.Int(0),
	})
	if err != nil {
		return fmt.Errorf("chat completion request failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_logprobs", "response is nil")
	}
	if resp.ID == "" {
		return fail("chat_completions_logprobs", "response missing id")
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_logprobs", "response missing choices")
	}

	choice := resp.Choices[0]
	if choice.FinishReason == "" {
		return fail("chat_completions_logprobs", "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail("chat_completions_logprobs", fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if !hasChatMessageOutput(choice.Message) && !isContentFilterFinishReason(choice.FinishReason) {
		return fail("chat_completions_logprobs", "choice message has no content or refusal")
	}
	if !choice.JSON.Logprobs.Valid() {
		return fail("chat_completions_logprobs", "choice missing logprobs")
	}
	if !choice.Logprobs.JSON.Content.Valid() {
		return fail("chat_completions_logprobs", "choice logprobs missing content field")
	}
	return nil
}