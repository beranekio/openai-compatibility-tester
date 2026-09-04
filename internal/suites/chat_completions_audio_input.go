package suites

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/testutil"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsAudioInput verifies POST /v1/chat/completions with audio input.
type ChatCompletionsAudioInput struct{}

func (ChatCompletionsAudioInput) Name() string { return "chat_completions_audio_input" }
func (ChatCompletionsAudioInput) Description() string {
	return "Chat completion with audio input (POST /v1/chat/completions)"
}

func (ChatCompletionsAudioInput) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.TextContentPart("Transcribe this audio in one short sentence."),
				openai.InputAudioContentPart(openai.ChatCompletionContentPartInputAudioInputAudioParam{
					Data:   base64.StdEncoding.EncodeToString(testutil.SmallWAVBytes()),
					Format: "wav",
				}),
			}),
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("chat completion audio input request failed: %w", err)
	}
	if err := validateChatCompletionEnvelope("chat_completions_audio_input", resp); err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_audio_input", "response missing choices")
	}
	return validateChatCompletionChoice("chat_completions_audio_input", resp.Choices[0])
}
