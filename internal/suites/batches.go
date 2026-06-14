package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

const batchPollInterval = 2 * time.Second

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
	if err := waitForBatchInputFile(ctx, client, uploaded.ID); err != nil {
		deleteBatchInputFile(client, uploaded.ID)
		return nil, err
	}
	return uploaded, nil
}

func deleteBatchInputFile(client openai.Client, fileID string) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _ = client.Files.Delete(cleanupCtx, fileID)
}

func waitForBatchInputFile(ctx context.Context, client openai.Client, fileID string) error {
	for {
		file, err := client.Files.Get(ctx, fileID)
		if err != nil {
			return fmt.Errorf("batch input file get failed: %w", err)
		}
		if !file.JSON.Status.Valid() {
			return nil
		}
		switch file.Status {
		case openai.FileObjectStatusProcessed:
			return nil
		case openai.FileObjectStatusError:
			return fail("batches", "batch input file processing failed")
		case openai.FileObjectStatusUploaded:
			// keep polling
		default:
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for batch input file processing: %w", ctx.Err())
		case <-time.After(batchPollInterval):
		}
	}
}

func isBatchCreateStatusOK(status string) bool {
	return status == "validating" || status == "in_progress"
}

func isBatchTerminalFailure(status string) bool {
	return status == "failed" || status == "expired" || status == "cancelled"
}

func isBatchCancelStatusOK(status string) bool {
	return status == "cancelling" || status == "cancelled"
}

func waitForBatchStatus(ctx context.Context, client openai.Client, suite, batchID string, accept func(string) bool) (*openai.Batch, error) {
	for {
		got, err := client.Batches.Get(ctx, batchID)
		if err != nil {
			return nil, fmt.Errorf("batch get failed: %w", err)
		}
		if err := validateBatchEnvelope(suite, got); err != nil {
			return nil, err
		}
		if got.ID != batchID {
			return nil, fail(suite, fmt.Sprintf("batch id is %q, want %q", got.ID, batchID))
		}
		status := string(got.Status)
		if accept(status) {
			return got, nil
		}
		if isBatchTerminalFailure(status) {
			return nil, fail(suite, fmt.Sprintf("batch failed with terminal status %q", status))
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for batch status: %w", ctx.Err())
		case <-time.After(batchPollInterval):
		}
	}
}

func cleanupBatchArtifacts(client openai.Client, batchID, fileID string) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if batchID != "" {
		skipCancel, err := waitForBatchCancelable(cleanupCtx, client, "batches", batchID)
		if err != nil || !skipCancel {
			_, _ = client.Batches.Cancel(cleanupCtx, batchID)
		}
	}
	if fileID != "" {
		_, _ = client.Files.Delete(cleanupCtx, fileID)
	}
}

func isBatchCancelAlreadyTerminalError(apiErr *openai.Error) bool {
	switch apiErr.StatusCode {
	case http.StatusConflict, http.StatusBadRequest:
		return true
	default:
		return false
	}
}

// exerciseBatchCancelEndpoint calls Cancel when the batch already completed before
// becoming cancelable. Only expected terminal-state errors prove the route exists.
func exerciseBatchCancelEndpoint(ctx context.Context, client openai.Client, suite, batchID string) error {
	cancelled, err := client.Batches.Cancel(ctx, batchID)
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && isBatchCancelAlreadyTerminalError(apiErr) {
			return nil
		}
		return fmt.Errorf("batch cancel failed: %w", err)
	}
	if err := validateBatchEnvelope(suite, cancelled); err != nil {
		return err
	}
	if cancelled.ID != batchID {
		return fail(suite, fmt.Sprintf("cancel id is %q, want %q", cancelled.ID, batchID))
	}
	return nil
}

func validateBatchObject(suite string, batch *openai.Batch) error {
	return validateBatchEnvelope(suite, batch)
}

func validateBatchEnvelope(suite string, batch *openai.Batch) error {
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
	return nil
}

func validateBatchRequestCounts(suite string, batch *openai.Batch) error {
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

// waitForBatchCancelable polls until the batch is in_progress (cancelable) or
// completed (too fast to cancel). Returns skipCancel=true when cancel is unnecessary.
func waitForBatchCancelable(ctx context.Context, client openai.Client, suite, batchID string) (skipCancel bool, err error) {
	for {
		got, err := client.Batches.Get(ctx, batchID)
		if err != nil {
			return false, fmt.Errorf("batch get failed: %w", err)
		}
		if err := validateBatchEnvelope(suite, got); err != nil {
			return false, err
		}
		if got.ID != batchID {
			return false, fail(suite, fmt.Sprintf("batch id is %q, want %q", got.ID, batchID))
		}
		status := string(got.Status)
		switch status {
		case "in_progress":
			return false, nil
		case "completed":
			return true, nil
		}
		if isBatchTerminalFailure(status) {
			return false, fail(suite, fmt.Sprintf("batch failed with terminal status %q", status))
		}
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("timed out waiting for cancelable batch status: %w", ctx.Err())
		case <-time.After(batchPollInterval):
		}
	}
}