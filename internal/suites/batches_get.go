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
	var batchID string
	var fileID string
	defer func() {
		cleanupBatchArtifacts(client, batchID, fileID)
	}()

	uploaded, err := uploadBatchInputFile(ctx, client, cfg)
	if err != nil {
		return err
	}
	fileID = uploaded.ID

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
	batchID = created.ID

	got, err := waitForBatchStatus(ctx, client, "batches_get", created.ID, func(status string) bool {
		return status == "completed"
	})
	if err != nil {
		return err
	}
	if got.RequestCounts.Completed != 1 {
		return fail("batches_get", fmt.Sprintf("batch request_counts.completed is %d, want 1", got.RequestCounts.Completed))
	}
	if got.RequestCounts.Total != 1 {
		return fail("batches_get", fmt.Sprintf("batch request_counts.total is %d, want 1", got.RequestCounts.Total))
	}
	return nil
}