package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesInputItems verifies GET /v1/responses/{id}/input_items.
type ResponsesInputItems struct{}

func (ResponsesInputItems) Name() string { return "responses_input_items" }
func (ResponsesInputItems) Description() string {
	return "Responses API input items (POST /v1/responses with store, then GET /v1/responses/{id}/input_items)"
}

func (ResponsesInputItems) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredResponse(ctx, client, cfg)
	if err != nil {
		return err
	}

	page, err := client.Responses.InputItems.List(ctx, created.ID, responses.InputItemListParams{})
	if err != nil {
		return fmt.Errorf("responses input_items list failed: %w", err)
	}
	if page == nil {
		return fail("responses_input_items", "list page is nil")
	}
	if len(page.Data) == 0 {
		return fail("responses_input_items", "list data is empty")
	}
	for _, item := range page.Data {
		if item.ID == "" {
			return fail("responses_input_items", "input item missing id")
		}
		if item.Type == "" {
			return fail("responses_input_items", "input item missing type")
		}
	}
	return nil
}