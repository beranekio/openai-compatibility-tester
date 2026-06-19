package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
	"github.com/openai/openai-go/v3/shared"
)

const (
	assistantCreateName         = "compatibility-test-assistant"
	assistantUpdateName         = "compatibility-test-assistant-updated"
	assistantCreateInstructions = "Reply with exactly the word: pong."
	assistantUpdateInstructions = "Reply with exactly the word: updated."
)

// Assistants verifies deprecated Assistants API CRUD via client.Beta.Assistants.*.
type Assistants struct{}

func (Assistants) Name() string { return "assistants" }
func (Assistants) Description() string {
	return "Deprecated Assistants API CRUD (POST/GET/DELETE /v1/assistants)"
}
func (Assistants) Deprecated() bool { return true }

func (Assistants) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	deleted := false
	var assistantID string
	defer func() {
		if assistantID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Beta.Assistants.Delete(cleanupCtx, assistantID)
		}
	}()

	created, err := client.Beta.Assistants.New(ctx, openai.BetaAssistantNewParams{
		Model:        shared.ChatModel(cfg.Model),
		Name:         openai.String(assistantCreateName),
		Instructions: openai.String(assistantCreateInstructions),
		Metadata: shared.Metadata{
			"suite": "assistants",
		},
	})
	if err != nil {
		return fmt.Errorf("assistant create failed: %w", err)
	}
	assistantID = created.ID
	if err := validateAssistantObject("assistants", created); err != nil {
		return err
	}
	if created.Name != assistantCreateName {
		return fail("assistants", fmt.Sprintf("create name is %q, want %q", created.Name, assistantCreateName))
	}

	got, err := client.Beta.Assistants.Get(ctx, assistantID)
	if err != nil {
		return fmt.Errorf("assistant get failed: %w", err)
	}
	if err := validateAssistantObject("assistants", got); err != nil {
		return err
	}
	if got.ID != assistantID {
		return fail("assistants", fmt.Sprintf("get id is %q, want %q", got.ID, assistantID))
	}

	updated, err := client.Beta.Assistants.Update(ctx, assistantID, openai.BetaAssistantUpdateParams{
		Name:         openai.String(assistantUpdateName),
		Instructions: openai.String(assistantUpdateInstructions),
		Metadata: shared.Metadata{
			"suite":  "assistants",
			"status": "updated",
		},
	})
	if err != nil {
		return fmt.Errorf("assistant update failed: %w", err)
	}
	if err := validateAssistantObject("assistants", updated); err != nil {
		return err
	}
	if updated.ID != assistantID {
		return fail("assistants", fmt.Sprintf("update id is %q, want %q", updated.ID, assistantID))
	}
	if updated.Name != assistantUpdateName {
		return fail("assistants", fmt.Sprintf("update name is %q, want %q", updated.Name, assistantUpdateName))
	}

	listPage, err := client.Beta.Assistants.List(ctx, openai.BetaAssistantListParams{
		Limit: openai.Int(100),
		Order: openai.BetaAssistantListParamsOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("assistant list failed: %w", err)
	}
	found, err := assistantListContains(listPage, assistantID)
	if err != nil {
		return err
	}
	if !found {
		return fail("assistants", "created assistant missing from list response")
	}

	deletedResp, err := client.Beta.Assistants.Delete(ctx, assistantID)
	if err != nil {
		return fmt.Errorf("assistant delete failed: %w", err)
	}
	if err := validateAssistantDeleted("assistants", deletedResp); err != nil {
		return err
	}
	if deletedResp.ID != assistantID {
		return fail("assistants", fmt.Sprintf("delete id is %q, want %q", deletedResp.ID, assistantID))
	}

	_, getErr := client.Beta.Assistants.Get(ctx, assistantID)
	if getErr == nil {
		return fail("assistants", "assistant get after delete succeeded; assistant still exists")
	}
	var apiError *openai.Error
	if !errors.As(getErr, &apiError) {
		return fmt.Errorf("assistant get after delete failed: %w", getErr)
	}
	if apiError.StatusCode != http.StatusNotFound {
		return fail("assistants", fmt.Sprintf("assistant get after delete returned status %d, want 404", apiError.StatusCode))
	}
	deleted = true
	return nil
}

func validateAssistantObject(suite string, assistant *openai.Assistant) error {
	if assistant == nil {
		return fail(suite, "assistant is nil")
	}
	if assistant.ID == "" {
		return fail(suite, "assistant missing id")
	}
	if !assistant.JSON.CreatedAt.Valid() {
		return fail(suite, "assistant missing created_at")
	}
	if !assistant.JSON.Metadata.Valid() {
		return fail(suite, "assistant missing metadata")
	}
	if !assistant.JSON.Model.Valid() {
		return fail(suite, "assistant missing model")
	}
	if assistant.Model == "" {
		return fail(suite, "assistant model is empty")
	}
	if !assistant.JSON.Name.Valid() {
		return fail(suite, "assistant missing name")
	}
	if !assistant.JSON.Object.Valid() {
		return fail(suite, "assistant missing object")
	}
	if string(assistant.Object) != "assistant" {
		return fail(suite, fmt.Sprintf("assistant object is %q, want assistant", assistant.Object))
	}
	if !assistant.JSON.Tools.Valid() {
		return fail(suite, "assistant missing tools")
	}
	return nil
}

func validateAssistantDeleted(suite string, deleted *openai.AssistantDeleted) error {
	if deleted == nil {
		return fail(suite, "delete response is nil")
	}
	if deleted.ID == "" {
		return fail(suite, "delete response missing id")
	}
	if !deleted.JSON.Deleted.Valid() {
		return fail(suite, "delete response missing deleted")
	}
	if !deleted.Deleted {
		return fail(suite, "delete response deleted is false")
	}
	if !deleted.JSON.Object.Valid() {
		return fail(suite, "delete response missing object")
	}
	if string(deleted.Object) != "assistant.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want assistant.deleted", deleted.Object))
	}
	return nil
}

func assistantListContains(page *pagination.CursorPage[openai.Assistant], assistantID string) (bool, error) {
	for page != nil {
		if err := validateAssistantListPage("assistants", page); err != nil {
			return false, err
		}
		for i := range page.Data {
			if page.Data[i].ID == assistantID {
				return true, nil
			}
		}
		next, err := page.GetNextPage()
		if err != nil {
			return false, fmt.Errorf("assistant list next page failed: %w", err)
		}
		page = next
	}
	return false, nil
}

func validateAssistantListPage(suite string, page *pagination.CursorPage[openai.Assistant]) error {
	if page == nil {
		return fail(suite, "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "list missing has_more")
	}
	var envelope struct {
		Object  string `json:"object"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	if len(page.Data) == 0 {
		return nil
	}
	if envelope.FirstID == "" {
		return fail(suite, "list missing first_id")
	}
	if envelope.LastID == "" {
		return fail(suite, "list missing last_id")
	}
	if envelope.FirstID != page.Data[0].ID {
		return fail(suite, fmt.Sprintf("list first_id is %q, want %q", envelope.FirstID, page.Data[0].ID))
	}
	if envelope.LastID != page.Data[len(page.Data)-1].ID {
		return fail(suite, fmt.Sprintf("list last_id is %q, want %q", envelope.LastID, page.Data[len(page.Data)-1].ID))
	}
	for i := range page.Data {
		if err := validateAssistantObject(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}