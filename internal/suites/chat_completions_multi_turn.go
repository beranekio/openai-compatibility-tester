package suites

import (
	"context"
	"fmt"
	"strings"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

const (
	multiTurnToolCallID     = "call_mock_weather"
	multiTurnToolResultJSON = `{"temperature": 72, "unit": "fahrenheit", "condition": "sunny"}`
	multiTurnExpectedTempF  = "72"
)

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
			openai.DeveloperMessage("Use the weather tool result when answering follow-up questions."),
			openai.UserMessage("What is the weather in San Francisco?"),
			multiTurnAssistantToolCallMessage(),
			openai.ToolMessage(multiTurnToolResultJSON, multiTurnToolCallID),
			openai.UserMessage("What temperature did the weather tool report in Fahrenheit? Reply with the number only."),
		},
		Tools:      weatherTools(),
		ToolChoice: noToolChoice(),
		Store:      openai.Bool(false),
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
	if isContentFilterFinishReason(choice.FinishReason) {
		return nil
	}
	if choice.Message.Refusal != "" {
		return nil
	}
	if !strings.Contains(choice.Message.Content, multiTurnExpectedTempF) {
		return fail("chat_completions_multi_turn", fmt.Sprintf("choice content is %q, want response containing %q from tool context", choice.Message.Content, multiTurnExpectedTempF))
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
