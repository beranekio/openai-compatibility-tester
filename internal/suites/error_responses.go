package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

const errorTriggerModel = "oct-invalid-model"

// ErrorResponses verifies that the endpoint returns parseable OpenAI error payloads.
type ErrorResponses struct{}

func (ErrorResponses) Name() string { return "error_responses" }
func (ErrorResponses) Description() string {
	return "OpenAI-compatible error responses (invalid request)"
}

func (ErrorResponses) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	_, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: errorTriggerModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("This request should fail."),
		},
		Store: openai.Bool(false),
	})
	if err == nil {
		return fail("error_responses", "expected request to fail with an API error")
	}

	var apiErr *openai.Error
	if !errors.As(err, &apiErr) {
		return fail("error_responses", fmt.Sprintf("error is %T, want *openai.Error", err))
	}
	if apiErr.Message == "" {
		return fail("error_responses", "error missing message")
	}
	if apiErr.Type == "" {
		return fail("error_responses", "error missing type")
	}
	if apiErr.StatusCode != http.StatusBadRequest && apiErr.StatusCode != http.StatusNotFound {
		return fail("error_responses", fmt.Sprintf("status code is %d, want 400 or 404 for invalid model", apiErr.StatusCode))
	}
	if apiErr.JSON.Param.Valid() && apiErr.Param != "model" {
		return fail("error_responses", fmt.Sprintf("error param is %q, want model", apiErr.Param))
	}
	return nil
}