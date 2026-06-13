package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/openai/openai-go/v3"
)

// ModelsGet verifies GET /v1/models/{id} via client.Models.Get.
type ModelsGet struct{}

func (ModelsGet) Name() string        { return "models_get" }
func (ModelsGet) Description() string { return "Retrieve model by ID (GET /v1/models/{id})" }

func (ModelsGet) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	model, err := client.Models.Get(ctx, cfg.Model)
	if err != nil {
		return fmt.Errorf("models get request failed: %w", err)
	}
	if model == nil {
		return fail("models_get", "response is nil")
	}
	if model.ID == "" {
		return fail("models_get", "model missing id")
	}
	if model.ID != cfg.Model {
		return fail("models_get", fmt.Sprintf("model id is %q, want %q", model.ID, cfg.Model))
	}
	if string(model.Object) != "model" {
		return fail("models_get", fmt.Sprintf("model object is %q, want model", model.Object))
	}
	if !model.JSON.Created.Valid() {
		return fail("models_get", "model missing created")
	}
	if model.OwnedBy == "" {
		return fail("models_get", "model missing owned_by")
	}
	return nil
}
