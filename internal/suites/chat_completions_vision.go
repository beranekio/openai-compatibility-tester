package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// smallPNGDataURL is an 8x8 PNG encoded as a data URL for vision requests.
// Avoid 1x1 images: some vLLM versions crash on tiny inputs (CVE-2026-22773).
const smallPNGDataURL = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAYAAADED76LAAAAEklEQVR4nGP4n2L0Hx9mGBkKACBDpQFoN/xgAAAAAElFTkSuQmCC"

// ChatCompletionsVision verifies multimodal POST /v1/chat/completions with image input.
type ChatCompletionsVision struct{}

func (ChatCompletionsVision) Name() string { return "chat_completions_vision" }
func (ChatCompletionsVision) Description() string {
	return "Chat completion with vision input (POST /v1/chat/completions)"
}

func (ChatCompletionsVision) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.VisionModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.TextContentPart("Describe this image in one short sentence."),
				openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: smallPNGDataURL,
				}),
			}),
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("chat completion vision request failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_vision", "response is nil")
	}
	if resp.ID == "" {
		return fail("chat_completions_vision", "response missing id")
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_vision", "response missing choices")
	}

	choice := resp.Choices[0]
	if choice.FinishReason == "" {
		return fail("chat_completions_vision", "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail("chat_completions_vision", fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if !hasChatMessageOutput(choice.Message) && !isContentFilterFinishReason(choice.FinishReason) {
		return fail("chat_completions_vision", "choice message has no content or refusal")
	}
	return nil
}
