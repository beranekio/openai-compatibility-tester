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
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses request failed: %w", err)
	}
	if resp == nil {
		return fail("responses", "response is nil")
	}
	if resp.ID == "" {
		return fail("responses", "response missing id")
	}
	if !resp.JSON.CreatedAt.Valid() {
		return fail("responses", "response missing created_at")
	}
	if resp.Model == "" {
		return fail("responses", "response missing model")
	}
	if string(resp.Object) != "response" {
		return fail("responses", fmt.Sprintf("response object is %q, want response", resp.Object))
	}
	if string(resp.Status) != "completed" {
		return fail("responses", fmt.Sprintf("response status is %q, want completed", resp.Status))
	}
	if !hasResponseOutput(resp) {
		return fail("responses", "response produced no output text or refusal")
	}
	return nil
}