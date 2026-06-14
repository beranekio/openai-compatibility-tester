package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ChatCompletionsStreamUsage verifies streaming chat completions with usage stats.
type ChatCompletionsStreamUsage struct{}

func (ChatCompletionsStreamUsage) Name() string { return "chat_completions_stream_usage" }
func (ChatCompletionsStreamUsage) Description() string {
	return "Streaming chat completion with stream_options.include_usage (POST /v1/chat/completions, stream=true)"
}

func (ChatCompletionsStreamUsage) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Count from one to three."),
		},
		Store: openai.Bool(false),
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat completion stream failed: %w", err)
	}
	if err := validateEventStreamContentType("chat_completions_stream_usage", httpResp); err != nil {
		return err
	}

	chunks := 0
	var hasOutput bool
	var finished bool
	var finishReason string
	var expectUsageOnlyChunk bool
	var usageOnTerminalChunk bool
	var usageOnFinalChunk bool
	for stream.Next() {
		chunk := stream.Current()
		chunks++
		if chunk.ID == "" {
			return fail("chat_completions_stream_usage", "stream chunk missing id")
		}
		if err := validateChatCompletionChunkEnvelope("chat_completions_stream_usage", chunk); err != nil {
			return err
		}

		if len(chunk.Choices) == 0 {
			if !finished {
				return fail("chat_completions_stream_usage", "stream emitted usage-only chunk before finish_reason")
			}
			if !expectUsageOnlyChunk {
				return fail("chat_completions_stream_usage", "stream emitted unexpected usage-only chunk after finish_reason")
			}
			if err := validateChatCompletionStreamUsage("chat_completions_stream_usage", chunk); err != nil {
				return err
			}
			usageOnFinalChunk = true
			expectUsageOnlyChunk = false
			continue
		}

		if finished {
			return fail("chat_completions_stream_usage", "stream emitted choice chunk after finish_reason")
		}

		if err := validateChatCompletionChunkChoice("chat_completions_stream_usage", chunk); err != nil {
			return err
		}
		choice := chunk.Choices[0]
		if choice.FinishReason != "" {
			finished = true
			finishReason = choice.FinishReason
			if chunk.JSON.Usage.Valid() {
				if err := validateChatCompletionStreamUsage("chat_completions_stream_usage", chunk); err != nil {
					return err
				}
				usageOnTerminalChunk = true
			} else {
				expectUsageOnlyChunk = true
			}
		}
		if choice.Delta.Content != "" || choice.Delta.Refusal != "" {
			hasOutput = true
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat completion stream failed: %w", err)
	}
	if chunks == 0 {
		return fail("chat_completions_stream_usage", "stream returned no chunks")
	}
	if !finished {
		return fail("chat_completions_stream_usage", "stream missing terminal finish_reason")
	}
	if expectUsageOnlyChunk {
		return fail("chat_completions_stream_usage", "stream missing usage-only chunk after finish_reason")
	}
	if !hasOutput && !isContentFilterFinishReason(finishReason) {
		return fail("chat_completions_stream_usage", "stream produced no text content or refusal")
	}
	if !usageOnFinalChunk && !usageOnTerminalChunk {
		return fail("chat_completions_stream_usage", "stream missing usage on terminal or final chunk")
	}
	return nil
}