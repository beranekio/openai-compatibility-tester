package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsReasoning verifies POST /v1/chat/completions for reasoning models.
type ChatCompletionsReasoning struct{}

func (ChatCompletionsReasoning) Name() string { return "chat_completions_reasoning" }
func (ChatCompletionsReasoning) Description() string {
	return "Chat completion with reasoning model output (POST /v1/chat/completions)"
}

func (ChatCompletionsReasoning) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.ReasoningModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Reply with exactly the word: pong"),
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("chat completion reasoning request failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_reasoning", "response is nil")
	}
	if resp.ID == "" {
		return fail("chat_completions_reasoning", "response missing id")
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_reasoning", "response missing choices")
	}

	choice := resp.Choices[0]
	if choice.FinishReason == "" {
		return fail("chat_completions_reasoning", "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail("chat_completions_reasoning", fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if hasChatReasoningOutput(choice.Message, choice.FinishReason, resp.Usage) {
		return nil
	}
	return fail("chat_completions_reasoning", "choice has no content, refusal, reasoning signal, or content_filter finish")
}

func hasChatReasoningOutput(msg openai.ChatCompletionMessage, finishReason string, usage openai.CompletionUsage) bool {
	if hasChatMessageOutput(msg) {
		return true
	}
	if usage.CompletionTokensDetails.ReasoningTokens > 0 {
		return true
	}
	if hasChatMessageReasoningContent(msg) {
		return true
	}
	return isContentFilterFinishReason(finishReason)
}

func hasChatMessageReasoningContent(msg openai.ChatCompletionMessage) bool {
	field, ok := msg.JSON.ExtraFields["reasoning_content"]
	if !ok || !field.Valid() {
		return false
	}
	var content string
	if err := json.Unmarshal([]byte(field.Raw()), &content); err != nil {
		return false
	}
	return strings.TrimSpace(content) != ""
}
