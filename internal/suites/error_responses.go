package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	return validateErrorResponseAPIError("error_responses", apiErr)
}

func validateErrorResponseAPIError(suite string, apiErr *openai.Error) error {
	if apiErr.Message == "" {
		return fail(suite, "error missing message")
	}
	if apiErr.Type == "" {
		return fail(suite, "error missing type")
	}
	if apiErr.StatusCode < 400 || apiErr.StatusCode >= 500 {
		return fail(suite, fmt.Sprintf("status code is %d, want 4xx", apiErr.StatusCode))
	}
	if isExcludedErrorStatus(apiErr.StatusCode) {
		return fail(suite, fmt.Sprintf("status code is %d, want client error other than 401/403/429", apiErr.StatusCode))
	}
	if !hasModelErrorEvidence(apiErr) {
		return fail(suite, "error lacks model-specific evidence (code, param, or message)")
	}
	return nil
}

func isExcludedErrorStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}

func hasModelErrorEvidence(apiErr *openai.Error) bool {
	if apiErr.JSON.Code.Valid() && isModelRelatedErrorCode(apiErr.Code) {
		return true
	}
	if apiErr.JSON.Param.Valid() && apiErr.Param == "model" {
		return true
	}
	if strings.Contains(apiErr.Message, errorTriggerModel) {
		return true
	}
	return strings.Contains(strings.ToLower(apiErr.Message), "model")
}

func isModelRelatedErrorCode(code string) bool {
	switch code {
	case "model_not_found", "invalid_model", "model_not_available", "invalid_model_error":
		return true
	default:
		return strings.Contains(strings.ToLower(code), "model")
	}
}