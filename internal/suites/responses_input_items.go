package suites

import (
	"context"
	"encoding/json"
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
	defer func() {
		_ = client.Responses.Delete(ctx, created.ID)
	}()

	page, err := client.Responses.InputItems.List(ctx, created.ID, responses.InputItemListParams{})
	if err != nil {
		return fmt.Errorf("responses input_items list failed: %w", err)
	}
	if page == nil {
		return fail("responses_input_items", "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail("responses_input_items", "list missing has_more")
	}
	var envelope struct {
		Object  string `json:"object"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail("responses_input_items", "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail("responses_input_items", fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	if len(page.Data) == 0 {
		return fail("responses_input_items", "list data is empty")
	}
	if envelope.FirstID == "" {
		return fail("responses_input_items", "list missing first_id")
	}
	if envelope.LastID == "" {
		return fail("responses_input_items", "list missing last_id")
	}
	for _, item := range page.Data {
		if item.ID == "" {
			return fail("responses_input_items", "input item missing id")
		}
		if item.Type == "" {
			return fail("responses_input_items", "input item missing type")
		}
	}
	if !listContainsStoredInput(page.Data, storedResponseInput) {
		return fail("responses_input_items", "list missing user input_text matching submitted prompt")
	}
	return nil
}

func listContainsStoredInput(items []responses.ResponseItemUnion, want string) bool {
	for _, item := range items {
		if item.Type != "message" {
			continue
		}
		msg := item.AsMessage()
		if string(msg.Role) != "user" {
			continue
		}
		for _, content := range msg.Content {
			if content.Type == "input_text" && content.AsInputText().Text == want {
				return true
			}
		}
	}
	return false
}