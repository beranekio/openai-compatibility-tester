package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/openai/openai-go/v3"
)

// Models verifies GET /v1/models via client.Models.List.
type Models struct{}

func (Models) Name() string        { return "models" }
func (Models) Description() string { return "List models (GET /v1/models)" }

func (Models) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	page, err := client.Models.List(ctx)
	if err != nil {
		return fmt.Errorf("models list request failed: %w", err)
	}
	if page == nil {
		return fail("models", "response is nil")
	}
	if page.Object != "list" {
		return fail("models", fmt.Sprintf("response object is %q, want list", page.Object))
	}
	if len(page.Data) == 0 {
		return fail("models", "expected at least one model in list response")
	}
	for _, model := range page.Data {
		if model.ID == "" {
			return fail("models", "model entry missing id")
		}
		if !model.JSON.Created.Valid() {
			return fail("models", "model entry missing created")
		}
		if string(model.Object) != "model" {
			return fail("models", fmt.Sprintf("model object is %q, want model", model.Object))
		}
		if model.OwnedBy == "" {
			return fail("models", "model entry missing owned_by")
		}
	}
	return nil
}