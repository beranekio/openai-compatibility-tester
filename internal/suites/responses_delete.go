package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesDelete verifies DELETE /v1/responses/{id}.
type ResponsesDelete struct{}

func (ResponsesDelete) Name() string { return "responses_delete" }
func (ResponsesDelete) Description() string {
	return "Responses API delete (POST /v1/responses with store, then DELETE /v1/responses/{id})"
}

func (ResponsesDelete) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredResponse(ctx, client, cfg)
	if err != nil {
		return err
	}

	if err := client.Responses.Delete(ctx, created.ID); err != nil {
		return fmt.Errorf("responses delete failed: %w", err)
	}

	_, getErr := client.Responses.Get(ctx, created.ID, responses.ResponseGetParams{})
	if getErr == nil {
		return fail("responses_delete", "get after delete succeeded; response still exists")
	}
	var apiErr *openai.Error
	if errors.As(getErr, &apiErr) && apiErr.StatusCode != http.StatusNotFound {
		return fail("responses_delete", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}
	return nil
}