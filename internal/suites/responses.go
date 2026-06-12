package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// Responses verifies POST /v1/responses via client.Responses.New.
type Responses struct{}

func (Responses) Name() string        { return "responses" }
func (Responses) Description() string { return "Responses API (POST /v1/responses)" }

func (Responses) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.Model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Reply with exactly the word: pong"),
		},
		MaxOutputTokens: openai.Int(16),
	})
	if err != nil {
		return fmt.Errorf("responses request failed: %w", err)
	}
	if resp.ID == "" {
		return fail("responses", "response missing id")
	}
	if resp.OutputText() == "" {
		return fail("responses", "response produced no output text")
	}
	return nil
}