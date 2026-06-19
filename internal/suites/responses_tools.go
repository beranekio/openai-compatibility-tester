package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesTools verifies tool calling on POST /v1/responses.
type ResponsesTools struct{}

func (ResponsesTools) Name() string { return "responses_tools" }
func (ResponsesTools) Description() string {
	return "Responses API with tools (POST /v1/responses, tool_choice required)"
}

func (ResponsesTools) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("What is the weather in San Francisco?"),
		},
		Tools:      weatherResponseTools(),
		ToolChoice: requiredResponseToolChoice(),
		Store:      openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses tools request failed: %w", err)
	}
	if err := validateResponseEnvelope("responses_tools", resp); err != nil {
		return err
	}
	if string(resp.Status) == "completed" {
		if hasResponseFunctionCalls(resp) {
			for _, call := range responseFunctionCalls(resp) {
				if err := validateResponseFunctionToolCall("responses_tools", call); err != nil {
					return err
				}
			}
			return nil
		}
		if responseOutputRefusal(resp) != "" {
			return nil
		}
		return fail("responses_tools", "response produced no function_call output or refusal")
	}
	if isContentFilterIncompleteResponse(resp) {
		return nil
	}
	return fail("responses_tools", fmt.Sprintf("response status is %q, want completed", resp.Status))
}
