package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// Completions verifies POST /v1/completions via client.Completions.New.
type Completions struct{}

func (Completions) Name() string        { return "completions" }
func (Completions) Description() string { return "Text completion (POST /v1/completions)" }

func (Completions) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Completions.New(ctx, openai.CompletionNewParams{
		Model: openai.CompletionNewParamsModel(cfg.CompletionModel),
		Prompt: openai.CompletionNewParamsPromptUnion{
			OfString: openai.String("Say hello"),
		},
		MaxTokens: openai.Int(16),
	})
	if err != nil {
		return fmt.Errorf("completion request failed: %w", err)
	}
	if resp == nil {
		return fail("completions", "response is nil")
	}
	if resp.ID == "" {
		return fail("completions", "response missing id")
	}
	if !resp.JSON.Created.Valid() {
		return fail("completions", "response missing created")
	}
	if resp.Model == "" {
		return fail("completions", "response missing model")
	}
	if string(resp.Object) != "text_completion" {
		return fail("completions", fmt.Sprintf("response object is %q, want text_completion", resp.Object))
	}
	if len(resp.Choices) == 0 {
		return fail("completions", "response missing choices")
	}
	choice := resp.Choices[0]
	if !choice.JSON.Index.Valid() {
		return fail("completions", "choice missing index")
	}
	if string(choice.FinishReason) == "" {
		return fail("completions", "choice missing finish_reason")
	}
	if choice.Text == "" && choice.FinishReason != openai.CompletionChoiceFinishReasonContentFilter {
		return fail("completions", "choice text is empty")
	}
	return nil
}
