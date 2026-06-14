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
	if err := validateChatCompletionEnvelope("chat_completions_logprobs", resp); err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_logprobs", "response missing choices")
	}

	choice := resp.Choices[0]
	if isContentFilterFinishReason(choice.FinishReason) {
		return nil
	}
	if err := validateChatCompletionChoice("chat_completions_logprobs", choice); err != nil {
		return err
	}
	if !choice.JSON.Logprobs.Valid() {
		return fail("chat_completions_logprobs", "choice missing logprobs")
	}
	if err := validateChatCompletionLogprobsFields("chat_completions_logprobs", choice.Logprobs); err != nil {
		return err
	}
	if choice.Message.Content != "" {
		if err := validateChatCompletionLogprobsContent("chat_completions_logprobs", choice.Logprobs); err != nil {
			return err
		}
	}
	if choice.Message.Refusal != "" {
		if err := validateChatCompletionLogprobsRefusal("chat_completions_logprobs", choice.Logprobs); err != nil {
			return err
		}
	}
	return nil
}

func validateChatCompletionLogprobsFields(suite string, logprobs openai.ChatCompletionChoiceLogprobs) error {
	if !logprobs.JSON.Content.Valid() {
		return fail(suite, "choice logprobs missing content field")
	}
	if !logprobs.JSON.Refusal.Valid() {
		return fail(suite, "choice logprobs missing refusal field")
	}
	return nil
}

func validateChatCompletionLogprobsContent(suite string, logprobs openai.ChatCompletionChoiceLogprobs) error {
	if len(logprobs.Content) == 0 {
		return fail(suite, "choice logprobs content field is empty")
	}
	for _, entry := range logprobs.Content {
		if err := validateChatCompletionTokenLogprob(suite, entry); err != nil {
			return err
		}
	}
	return nil
}

func validateChatCompletionLogprobsRefusal(suite string, logprobs openai.ChatCompletionChoiceLogprobs) error {
	if len(logprobs.Refusal) == 0 {
		return fail(suite, "choice logprobs refusal field is empty")
	}
	for _, entry := range logprobs.Refusal {
		if err := validateChatCompletionTokenLogprob(suite, entry); err != nil {
			return err
		}
	}
	return nil
}

func validateChatCompletionTokenLogprob(suite string, entry openai.ChatCompletionTokenLogprob) error {
	if !entry.JSON.Token.Valid() {
		return fail(suite, "logprob entry missing token")
	}
	if !entry.JSON.Bytes.Valid() {
		return fail(suite, "logprob entry missing bytes")
	}
	if !entry.JSON.Logprob.Valid() {
		return fail(suite, "logprob entry missing logprob")
	}
	if !entry.JSON.TopLogprobs.Valid() {
		return fail(suite, "logprob entry missing top_logprobs")
	}
	return nil
}