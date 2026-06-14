package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// BatchesGet verifies GET /v1/batches/{id} via client.Batches.Get.
type BatchesGet struct{}

func (BatchesGet) Name() string { return "batches_get" }
func (BatchesGet) Description() string {
	return "Batches API get (POST /v1/batches, then GET /v1/batches/{id})"
}

func (BatchesGet) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	uploaded, err := uploadBatchInputFile(ctx, client, cfg)
	if err != nil {
		return err
	}

	created, err := client.Batches.New(ctx, openai.BatchNewParams{
		CompletionWindow: openai.BatchNewParamsCompletionWindow24h,
		Endpoint:         openai.BatchNewParamsEndpointV1ChatCompletions,
		InputFileID:      uploaded.ID,
	})
	if err != nil {
		return fmt.Errorf("batch create failed: %w", err)
	}
	if created == nil || created.ID == "" {
		return fail("batches_get", "batch create missing id")
	}

	got, err := client.Batches.Get(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("batch get failed: %w", err)
	}
	if err := validateBatchObject("batches_get", got); err != nil {
		return err
	}
	if got.ID != created.ID {
		return fail("batches_get", fmt.Sprintf("batch id is %q, want %q", got.ID, created.ID))
	}
	if string(got.Status) != "completed" {
		return fail("batches_get", fmt.Sprintf("batch status is %q, want completed", got.Status))
	}
	if got.RequestCounts.Completed != 1 {
		return fail("batches_get", fmt.Sprintf("batch request_counts.completed is %d, want 1", got.RequestCounts.Completed))
	}
	if got.RequestCounts.Total != 1 {
		return fail("batches_get", fmt.Sprintf("batch request_counts.total is %d, want 1", got.RequestCounts.Total))
	}
	return nil
}