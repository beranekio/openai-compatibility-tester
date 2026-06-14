package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

const multiTurnToolCallID = "call_mock_weather"

// ChatCompletionsMultiTurn verifies multi-turn POST /v1/chat/completions with
// system, user, assistant history, developer, and tool messages.
type ChatCompletionsMultiTurn struct{}

func (ChatCompletionsMultiTurn) Name() string { return "chat_completions_multi_turn" }
func (ChatCompletionsMultiTurn) Description() string {
	return "Multi-turn chat completion with history, developer, and tool messages (POST /v1/chat/completions)"
}

func (ChatCompletionsMultiTurn) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant."),
			openai.UserMessage("What is the weather in San Francisco?"),
			multiTurnAssistantToolCallMessage(),
			openai.DeveloperMessage("Use the weather tool result when answering follow-up questions."),
			openai.ToolMessage(`{"temperature": 72, "unit": "fahrenheit", "condition": "sunny"}`, multiTurnToolCallID),
			openai.UserMessage("Reply with exactly the word: pong"),
		},
		Tools: weatherTools(),
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("multi-turn chat completion request failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_multi_turn", "response is nil")
	}
	if resp.ID == "" {
		return fail("chat_completions_multi_turn", "response missing id")
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_multi_turn", "response missing choices")
	}

	choice := resp.Choices[0]
	if choice.FinishReason == "" {
		return fail("chat_completions_multi_turn", "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail("chat_completions_multi_turn", fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if !hasChatMessageOutput(choice.Message) && !isContentFilterFinishReason(choice.FinishReason) {
		return fail("chat_completions_multi_turn", "choice message has no content or refusal")
	}
	return nil
}

func multiTurnAssistantToolCallMessage() openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
				{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: multiTurnToolCallID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      weatherToolName,
							Arguments: `{"location":"San Francisco, CA"}`,
						},
					},
				},
			},
		},
	}
}