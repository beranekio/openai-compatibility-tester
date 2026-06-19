package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsTools verifies tool calling on POST /v1/chat/completions.
type ChatCompletionsTools struct{}

func (ChatCompletionsTools) Name() string { return "chat_completions_tools" }
func (ChatCompletionsTools) Description() string {
	return "Chat completion with tools (POST /v1/chat/completions, tool_choice required)"
}

func (ChatCompletionsTools) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("What is the weather in San Francisco?"),
		},
		Tools:      weatherTools(),
		ToolChoice: requiredToolChoice(),
		Store:      openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("chat completion tools request failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_tools", "response is nil")
	}
	if resp.ID == "" {
		return fail("chat_completions_tools", "response missing id")
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_tools", "response missing choices")
	}

	choice := resp.Choices[0]
	if choice.FinishReason == "" {
		return fail("chat_completions_tools", "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail("chat_completions_tools", fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if isContentFilterFinishReason(choice.FinishReason) {
		return nil
	}
	if choice.Message.Refusal != "" {
		return nil
	}
	if !hasToolCalls(choice.Message.ToolCalls) {
		return fail("chat_completions_tools", "choice has no tool_calls")
	}
	if !isToolCallsFinishReason(choice.FinishReason) {
		return fail("chat_completions_tools", fmt.Sprintf("finish_reason is %q, want tool_calls", choice.FinishReason))
	}
	for _, call := range choice.Message.ToolCalls {
		if err := validateFunctionToolCall("chat_completions_tools", call); err != nil {
			return err
		}
	}
	return nil
}
