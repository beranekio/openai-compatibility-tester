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
	if err := validateBatchObject("batches_create", created); err != nil {
		return err
	}
	if created.InputFileID != uploaded.ID {
		return fail("batches_create", fmt.Sprintf("batch input_file_id is %q, want %q", created.InputFileID, uploaded.ID))
	}
	if string(created.Status) != "in_progress" {
		return fail("batches_create", fmt.Sprintf("batch status is %q, want in_progress", created.Status))
	}
	return nil
}