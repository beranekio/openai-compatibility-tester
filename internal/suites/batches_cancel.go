package suites

import (
	"context"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// BatchesCancel verifies POST /v1/batches/{id}/cancel via client.Batches.Cancel.
type BatchesCancel struct{}

func (BatchesCancel) Name() string { return "batches_cancel" }
func (BatchesCancel) Description() string {
	return "Batches API cancel (POST /v1/batches, then POST /v1/batches/{id}/cancel)"
}

func (BatchesCancel) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	deleted := false
	var fileID string
	defer func() {
		if fileID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Files.Delete(cleanupCtx, fileID)
		}
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
		return fail("batches_cancel", "batch create missing id")
	}

	cancelled, err := client.Batches.Cancel(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("batch cancel failed: %w", err)
	}
	if err := validateBatchObject("batches_cancel", cancelled); err != nil {
		return err
	}
	if cancelled.ID != created.ID {
		return fail("batches_cancel", fmt.Sprintf("cancel id is %q, want %q", cancelled.ID, created.ID))
	}
	if string(cancelled.Status) != "cancelled" {
		return fail("batches_cancel", fmt.Sprintf("cancel status is %q, want cancelled", cancelled.Status))
	}

	deletedResp, err := client.Files.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("batch input file delete failed: %w", err)
	}
	if deletedResp == nil || !deletedResp.Deleted {
		return fail("batches_cancel", "batch input file delete response invalid")
	}
	deleted = true
	return nil
}