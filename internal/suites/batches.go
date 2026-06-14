package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

func uploadBatchInputFile(ctx context.Context, client openai.Client, cfg *config.Config) (*openai.FileObject, error) {
	uploaded, err := client.Files.New(ctx, openai.FileNewParams{
		File:    smallBatchJSONLReader(cfg.Model),
		Purpose: openai.FilePurposeBatch,
	})
	if err != nil {
		return nil, fmt.Errorf("batch input file upload failed: %w", err)
	}
	if err := validateFileObject("batches", uploaded); err != nil {
		return nil, err
	}
	if string(uploaded.Purpose) != string(openai.FilePurposeBatch) {
		return nil, fail("batches", fmt.Sprintf("upload purpose is %q, want batch", uploaded.Purpose))
	}
	return uploaded, nil
}

func validateBatchObject(suite string, batch *openai.Batch) error {
	if batch == nil {
		return fail(suite, "batch is nil")
	}
	if batch.ID == "" {
		return fail(suite, "batch missing id")
	}
	if !batch.JSON.CreatedAt.Valid() {
		return fail(suite, "batch missing created_at")
	}
	if !batch.JSON.CompletionWindow.Valid() {
		return fail(suite, "batch missing completion_window")
	}
	if batch.CompletionWindow != "24h" {
		return fail(suite, fmt.Sprintf("batch completion_window is %q, want 24h", batch.CompletionWindow))
	}
	if !batch.JSON.Endpoint.Valid() {
		return fail(suite, "batch missing endpoint")
	}
	if batch.Endpoint != "/v1/chat/completions" {
		return fail(suite, fmt.Sprintf("batch endpoint is %q, want /v1/chat/completions", batch.Endpoint))
	}
	if !batch.JSON.InputFileID.Valid() {
		return fail(suite, "batch missing input_file_id")
	}
	if batch.InputFileID == "" {
		return fail(suite, "batch input_file_id is empty")
	}
	if !batch.JSON.Object.Valid() {
		return fail(suite, "batch missing object")
	}
	if string(batch.Object) != "batch" {
		return fail(suite, fmt.Sprintf("batch object is %q, want batch", batch.Object))
	}
	if !batch.JSON.Status.Valid() {
		return fail(suite, "batch missing status")
	}
	if !batch.JSON.RequestCounts.Valid() {
		return fail(suite, "batch missing request_counts")
	}
	if !batch.RequestCounts.JSON.Total.Valid() {
		return fail(suite, "batch request_counts missing total")
	}
	if !batch.RequestCounts.JSON.Completed.Valid() {
		return fail(suite, "batch request_counts missing completed")
	}
	if !batch.RequestCounts.JSON.Failed.Valid() {
		return fail(suite, "batch request_counts missing failed")
	}
	return nil
}