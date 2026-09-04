package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

// ResponsesReasoning verifies POST /v1/responses for reasoning models.
type ResponsesReasoning struct{}

func (ResponsesReasoning) Name() string { return "responses_reasoning" }
func (ResponsesReasoning) Description() string {
	return "Responses API with reasoning model output (POST /v1/responses)"
}

func (ResponsesReasoning) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ReasoningModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Reply with exactly the word: pong"),
		},
		Reasoning: shared.ReasoningParam{
			Effort: shared.ReasoningEffortLow,
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses reasoning request failed: %w", err)
	}
	return validateResponsesReasoningResult("responses_reasoning", resp)
}

func validateResponsesReasoningResult(suite string, resp *responses.Response) error {
	if err := validateResponseEnvelope(suite, resp); err != nil {
		return err
	}
	if string(resp.Status) == "completed" {
		if hasResponseReasoningSignal(resp) {
			return nil
		}
		return fail(suite, "completed response has no output text, refusal, or reasoning signal")
	}
	if isContentFilterIncompleteResponse(resp) {
		return nil
	}
	return fail(suite, fmt.Sprintf("response status is %q, want completed", resp.Status))
}

func hasResponseReasoningSignal(resp *responses.Response) bool {
	if hasResponseOutput(resp) {
		return true
	}
	if resp.JSON.Usage.Valid() && resp.Usage.OutputTokensDetails.ReasoningTokens > 0 {
		return true
	}
	for _, item := range resp.Output {
		if item.Type == "reasoning" {
			return true
		}
	}
	return false
}
