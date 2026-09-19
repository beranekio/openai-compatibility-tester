package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesPrewarm verifies POST /v1/responses with prewarm=true.
type ResponsesPrewarm struct{}

func (ResponsesPrewarm) Name() string { return "responses_prewarm" }
func (ResponsesPrewarm) Description() string {
	return "Responses API prompt-cache prewarm (POST /v1/responses, prewarm=true)"
}

func (ResponsesPrewarm) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Reply with exactly the word: pong"),
		},
		Store: openai.Bool(false),
		PromptCacheOptions: responses.ResponseNewParamsPromptCacheOptions{
			Prewarm: openai.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("responses prewarm request failed: %w", err)
	}
	if err := validateResponseEnvelope("responses_prewarm", resp); err != nil {
		return err
	}
	if string(resp.Status) == "completed" {
		return nil
	}
	if isContentFilterIncompleteResponse(resp) {
		return nil
	}
	return fail("responses_prewarm", fmt.Sprintf("response status is %q, want completed", resp.Status))
}
