package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ChatCompletionsToolsStream verifies streaming tool calling on chat completions.
type ChatCompletionsToolsStream struct{}

func (ChatCompletionsToolsStream) Name() string { return "chat_completions_tools_stream" }
func (ChatCompletionsToolsStream) Description() string {
	return "Streaming chat completion with tools (POST /v1/chat/completions, stream=true, tool_choice required)"
}

func (ChatCompletionsToolsStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("What is the weather in San Francisco?"),
		},
		Tools:      weatherTools(),
		ToolChoice: requiredToolChoice(),
		Store:      openai.Bool(false),
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat completion tools stream failed: %w", err)
	}
	if err := validateEventStreamContentType("chat_completions_tools_stream", httpResp); err != nil {
		return err
	}

	chunks := 0
	var hasRefusal bool
	var finished bool
	var finishReason string
	var terminalReached bool
	toolCalls := make(map[int]*accumulatedToolCall)
	for stream.Next() {
		if terminalReached {
			return fail("chat_completions_tools_stream", "stream chunk after terminal finish_reason")
		}

		chunk := stream.Current()
		chunks++
		if chunk.ID == "" {
			return fail("chat_completions_tools_stream", "stream chunk missing id")
		}
		if err := validateChatCompletionChunk("chat_completions_tools_stream", chunk); err != nil {
			return err
		}
		choice := chunk.Choices[0]
		if choice.Delta.Refusal != "" {
			hasRefusal = true
		}
		accumulateDeltaToolCalls(toolCalls, choice.Delta.ToolCalls)
		if choice.FinishReason != "" {
			finished = true
			finishReason = choice.FinishReason
			terminalReached = true
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat completion tools stream failed: %w", err)
	}
	if chunks == 0 {
		return fail("chat_completions_tools_stream", "stream returned no chunks")
	}
	if !finished {
		return fail("chat_completions_tools_stream", "stream missing terminal finish_reason")
	}
	if isContentFilterFinishReason(finishReason) {
		return nil
	}
	if hasRefusal {
		return nil
	}
	if !isToolCallsFinishReason(finishReason) {
		return fail("chat_completions_tools_stream", fmt.Sprintf("terminal finish_reason is %q, want tool_calls", finishReason))
	}
	if len(toolCalls) == 0 {
		return fail("chat_completions_tools_stream", "stream produced no tool_calls")
	}
	for _, call := range toolCalls {
		if err := validateAccumulatedToolCall("chat_completions_tools_stream", call); err != nil {
			return err
		}
	}
	return nil
}
