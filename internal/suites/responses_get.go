package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesGet verifies GET /v1/responses/{id} after creating a stored response.
type ResponsesGet struct{}

func (ResponsesGet) Name() string { return "responses_get" }
func (ResponsesGet) Description() string {
	return "Responses API get (POST /v1/responses with store, then GET /v1/responses/{id})"
}

func (ResponsesGet) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredResponse(ctx, client, cfg)
	if err != nil {
		return err
	}
	defer deleteStoredResponseBestEffort(client, created.ID)

	got, err := client.Responses.Get(ctx, created.ID, responses.ResponseGetParams{})
	if err != nil {
		return fmt.Errorf("responses get failed: %w", err)
	}
	if err := validateResponseEnvelope("responses_get", got); err != nil {
		return err
	}
	if got.ID != created.ID {
		return fail("responses_get", fmt.Sprintf("get id is %q, want %q", got.ID, created.ID))
	}
	if string(got.Status) == "completed" {
		if err := validateCompletedResponseHasOutput("responses_get", got); err != nil {
			return err
		}
		return nil
	}
	if isContentFilterIncompleteResponse(got) {
		return nil
	}
	return fail("responses_get", fmt.Sprintf("get status is %q, want completed", got.Status))
}
