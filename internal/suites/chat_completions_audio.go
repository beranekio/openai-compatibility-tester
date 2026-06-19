package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsAudio verifies POST /v1/chat/completions with audio output modalities.
type ChatCompletionsAudio struct{}

func (ChatCompletionsAudio) Name() string { return "chat_completions_audio" }
func (ChatCompletionsAudio) Description() string {
	return "Chat completion with audio output (POST /v1/chat/completions)"
}

func (ChatCompletionsAudio) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say pong"),
		},
		Modalities: []string{"text", "audio"},
		Audio: openai.ChatCompletionAudioParam{
			Format: openai.ChatCompletionAudioParamFormatWAV,
			Voice: openai.ChatCompletionAudioParamVoiceUnion{
				OfChatCompletionAudioVoiceString2: openai.String("alloy"),
			},
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("chat completion audio request failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_audio", "response is nil")
	}
	if resp.ID == "" {
		return fail("chat_completions_audio", "response missing id")
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_audio", "response missing choices")
	}

	choice := resp.Choices[0]
	if choice.FinishReason == "" {
		return fail("chat_completions_audio", "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail("chat_completions_audio", fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if isContentFilterFinishReason(choice.FinishReason) {
		return nil
	}
	if choice.Message.Refusal != "" && choice.Message.Content == "" {
		return nil
	}
	if !hasChatMessageOutput(choice.Message) && !hasChatCompletionAudioOutput(choice.Message) {
		return fail("chat_completions_audio", "choice message has no content, refusal, or audio")
	}
	return validateChatCompletionAudio("chat_completions_audio", choice.Message.Audio)
}

func hasChatCompletionAudioOutput(msg openai.ChatCompletionMessage) bool {
	return msg.JSON.Audio.Valid()
}

func validateChatCompletionAudio(suite string, audio openai.ChatCompletionAudio) error {
	if audio.ID == "" {
		return fail(suite, "audio missing id field")
	}
	if audio.Data == "" {
		return fail(suite, "audio missing data field")
	}
	if !audio.JSON.ExpiresAt.Valid() {
		return fail(suite, "audio missing expires_at field")
	}
	if !audio.JSON.Transcript.Valid() {
		return fail(suite, "audio missing transcript field")
	}
	return validateBase64WAVData(suite, audio.Data, 12)
}
