package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesInputTokens verifies POST /v1/responses/input_tokens.
type ResponsesInputTokens struct{}

func (ResponsesInputTokens) Name() string { return "responses_input_tokens" }
func (ResponsesInputTokens) Description() string {
	return "Responses API input token count (POST /v1/responses/input_tokens)"
}

func (ResponsesInputTokens) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.InputTokens.Count(ctx, responses.InputTokenCountParams{
		Model: openai.String(cfg.ResponsesModel),
		Input: responses.InputTokenCountParamsInputUnion{
			OfString: openai.String("Count tokens for this input."),
		},
	})
	if err != nil {
		return fmt.Errorf("responses input_tokens request failed: %w", err)
	}
	if resp == nil {
		return fail("responses_input_tokens", "response is nil")
	}
	if string(resp.Object) != "response.input_tokens" {
		return fail("responses_input_tokens", fmt.Sprintf("object is %q, want response.input_tokens", resp.Object))
	}
	if !resp.JSON.InputTokens.Valid() {
		return fail("responses_input_tokens", "response missing input_tokens")
	}
	if resp.InputTokens <= 0 {
		return fail("responses_input_tokens", "input_tokens must be greater than zero")
	}
	return nil
}
