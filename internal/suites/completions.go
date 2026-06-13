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
	if len(resp.Choices) == 0 {
		return fail("completions", "response missing choices")
	}
	if string(resp.Choices[0].FinishReason) == "" {
		return fail("completions", "choice missing finish_reason")
	}
	if resp.Choices[0].Text == "" {
		return fail("completions", "choice text is empty")
	}
	return nil
}