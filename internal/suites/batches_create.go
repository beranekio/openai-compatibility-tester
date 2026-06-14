package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// BatchesCreate verifies POST /v1/batches via client.Batches.New.
type BatchesCreate struct{}

func (BatchesCreate) Name() string { return "batches_create" }
func (BatchesCreate) Description() string {
	return "Batches API create (POST /v1/batches)"
}

func (BatchesCreate) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
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
	if err := validateBatchObject("batches_create", created); err != nil {
		return err
	}
	batchID = created.ID
	if created.InputFileID != uploaded.ID {
		return fail("batches_create", fmt.Sprintf("batch input_file_id is %q, want %q", created.InputFileID, uploaded.ID))
	}
	if !isBatchCreateStatusOK(string(created.Status)) {
		return fail("batches_create", fmt.Sprintf("batch status is %q, want validating or in_progress", created.Status))
	}
	return nil
}