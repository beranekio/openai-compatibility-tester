package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// VectorStoreFileBatches verifies vector store file batch ingestion via client.VectorStores.FileBatches.*.
type VectorStoreFileBatches struct{}

func (VectorStoreFileBatches) Name() string { return "vector_store_file_batches" }
func (VectorStoreFileBatches) Description() string {
	return "Vector Store File Batches API lifecycle (POST/GET /v1/vector_stores/{id}/file_batches)"
}

func (VectorStoreFileBatches) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	store, err := createVectorStoreForSuite(ctx, client, "vector_store_file_batches", "compatibility-test-vector-store-file-batches")
	if err != nil {
		return err
	}
	var cleanupFileIDs []string
	defer func() {
		cleanupVectorStoreArtifacts(client, store.ID, cleanupFileIDs...)
	}()

	var batchFileIDs []string
	for range 2 {
		uploaded, err := uploadVectorStoreSourceFile(ctx, client, "vector_store_file_batches")
		if err != nil {
			return err
		}
		batchFileIDs = append(batchFileIDs, uploaded.ID)
		cleanupFileIDs = append(cleanupFileIDs, uploaded.ID)
	}

	created, err := client.VectorStores.FileBatches.New(ctx, store.ID, openai.VectorStoreFileBatchNewParams{
		FileIDs: batchFileIDs,
	})
	if err != nil {
		return fmt.Errorf("vector store file batch create failed: %w", err)
	}
	if err := validateVectorStoreFileBatchObject("vector_store_file_batches", created, store.ID); err != nil {
		return err
	}
	if err := validateVectorStoreFileBatchUploadCounts("vector_store_file_batches", created, int64(len(batchFileIDs))); err != nil {
		return err
	}
	if err := validateVectorStoreFileBatchPreCancelStatus("vector_store_file_batches", created); err != nil {
		return err
	}

	got, err := client.VectorStores.FileBatches.Get(ctx, store.ID, created.ID)
	if err != nil {
		return fmt.Errorf("vector store file batch get failed: %w", err)
	}
	if err := validateVectorStoreFileBatchObject("vector_store_file_batches", got, store.ID); err != nil {
		return err
	}
	if got.ID != created.ID {
		return fail("vector_store_file_batches", fmt.Sprintf("get batch id is %q, want %q", got.ID, created.ID))
	}
	if err := validateVectorStoreFileBatchUploadCounts("vector_store_file_batches", got, int64(len(batchFileIDs))); err != nil {
		return err
	}
	if err := validateVectorStoreFileBatchPreCancelStatus("vector_store_file_batches", got); err != nil {
		return err
	}

	otherBatchFile, err := uploadVectorStoreSourceFile(ctx, client, "vector_store_file_batches")
	if err != nil {
		return err
	}
	cleanupFileIDs = append(cleanupFileIDs, otherBatchFile.ID)
	otherBatch, err := client.VectorStores.FileBatches.New(ctx, store.ID, openai.VectorStoreFileBatchNewParams{
		FileIDs: []string{otherBatchFile.ID},
	})
	if err != nil {
		return fmt.Errorf("vector store second file batch create failed: %w", err)
	}
	if err := validateVectorStoreFileBatchObject("vector_store_file_batches", otherBatch, store.ID); err != nil {
		return err
	}
	if err := validateVectorStoreFileBatchUploadCounts("vector_store_file_batches", otherBatch, 1); err != nil {
		return err
	}
	if err := validateVectorStoreFileBatchPreCancelStatus("vector_store_file_batches", otherBatch); err != nil {
		return err
	}

	listPage, err := client.VectorStores.FileBatches.ListFiles(ctx, store.ID, created.ID, openai.VectorStoreFileBatchListFilesParams{
		Limit: openai.Int(10),
	})
	if err != nil {
		return fmt.Errorf("vector store file batch list files failed: %w", err)
	}
	if err := validateVectorStoreFileListPage("vector_store_file_batches", listPage, store.ID); err != nil {
		return err
	}
	for _, fileID := range batchFileIDs {
		if !vectorStoreFileListContains(listPage.Data, fileID) {
			return fail("vector_store_file_batches", fmt.Sprintf("batch file %q missing from list files response", fileID))
		}
	}
	if vectorStoreFileListContains(listPage.Data, otherBatchFile.ID) {
		return fail("vector_store_file_batches", "batch list files response included non-batch file")
	}

	cancelled, err := client.VectorStores.FileBatches.Cancel(ctx, store.ID, created.ID)
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && isVectorStoreFileBatchCancelAlreadyTerminalError(apiErr) {
			return nil
		}
		return fmt.Errorf("vector store file batch cancel failed: %w", err)
	}
	if err := validateVectorStoreFileBatchObject("vector_store_file_batches", cancelled, store.ID); err != nil {
		return err
	}
	if cancelled.ID != created.ID {
		return fail("vector_store_file_batches", fmt.Sprintf("cancel batch id is %q, want %q", cancelled.ID, created.ID))
	}
	if err := validateVectorStoreFileBatchUploadCounts("vector_store_file_batches", cancelled, int64(len(batchFileIDs))); err != nil {
		return err
	}
	if !isVectorStoreFileBatchCancelStatusOK(string(cancelled.Status)) {
		return fail("vector_store_file_batches", fmt.Sprintf("cancel status is %q, want in_progress, cancelled, or completed", cancelled.Status))
	}
	return nil
}

func validateVectorStoreFileBatchUploadCounts(suite string, batch *openai.VectorStoreFileBatch, wantTotal int64) error {
	if batch.FileCounts.Total != wantTotal {
		return fail(suite, fmt.Sprintf("batch file_counts.total is %d, want %d", batch.FileCounts.Total, wantTotal))
	}
	if batch.FileCounts.Failed != 0 {
		return fail(suite, fmt.Sprintf("batch file_counts.failed is %d, want 0", batch.FileCounts.Failed))
	}
	return nil
}

func validateVectorStoreFileBatchPreCancelStatus(suite string, batch *openai.VectorStoreFileBatch) error {
	if !isVectorStoreFileBatchPreCancelStatusOK(string(batch.Status)) {
		return fail(suite, fmt.Sprintf("batch status is %q, want in_progress or completed", batch.Status))
	}
	return nil
}

func validateVectorStoreFileBatchObject(suite string, batch *openai.VectorStoreFileBatch, expectedVectorStoreID string) error {
	if batch == nil {
		return fail(suite, "vector store file batch is nil")
	}
	if batch.ID == "" {
		return fail(suite, "vector store file batch missing id")
	}
	if !batch.JSON.CreatedAt.Valid() {
		return fail(suite, "vector store file batch missing created_at")
	}
	if !batch.JSON.FileCounts.Valid() {
		return fail(suite, "vector store file batch missing file_counts")
	}
	if err := validateVectorStoreFileBatchFileCounts(suite, batch.FileCounts); err != nil {
		return err
	}
	if !batch.JSON.Object.Valid() {
		return fail(suite, "vector store file batch missing object")
	}
	if !isVectorStoreFileBatchObjectOK(string(batch.Object)) {
		return fail(suite, fmt.Sprintf("vector store file batch object is %q, want vector_store.files_batch or vector_store.file_batch", batch.Object))
	}
	if !batch.JSON.Status.Valid() {
		return fail(suite, "vector store file batch missing status")
	}
	if !isVectorStoreFileBatchStatusOK(batch.Status) {
		return fail(suite, fmt.Sprintf("vector store file batch status is %q, want in_progress, completed, or cancelled", batch.Status))
	}
	if !batch.JSON.VectorStoreID.Valid() {
		return fail(suite, "vector store file batch missing vector_store_id")
	}
	if batch.VectorStoreID == "" {
		return fail(suite, "vector store file batch vector_store_id is empty")
	}
	if expectedVectorStoreID != "" && batch.VectorStoreID != expectedVectorStoreID {
		return fail(suite, fmt.Sprintf("vector store file batch vector_store_id is %q, want %q", batch.VectorStoreID, expectedVectorStoreID))
	}
	return nil
}

func isVectorStoreFileBatchObjectOK(object string) bool {
	return object == "vector_store.files_batch" || object == "vector_store.file_batch"
}

func isVectorStoreFileBatchStatusOK(status openai.VectorStoreFileBatchStatus) bool {
	switch status {
	case openai.VectorStoreFileBatchStatusInProgress, openai.VectorStoreFileBatchStatusCompleted, openai.VectorStoreFileBatchStatusCancelled:
		return true
	default:
		return false
	}
}

func validateVectorStoreFileBatchFileCounts(suite string, counts openai.VectorStoreFileBatchFileCounts) error {
	if !counts.JSON.Cancelled.Valid() {
		return fail(suite, "vector store file batch file_counts missing cancelled")
	}
	if !counts.JSON.Completed.Valid() {
		return fail(suite, "vector store file batch file_counts missing completed")
	}
	if !counts.JSON.Failed.Valid() {
		return fail(suite, "vector store file batch file_counts missing failed")
	}
	if !counts.JSON.InProgress.Valid() {
		return fail(suite, "vector store file batch file_counts missing in_progress")
	}
	if !counts.JSON.Total.Valid() {
		return fail(suite, "vector store file batch file_counts missing total")
	}
	for _, count := range []struct {
		name  string
		value int64
	}{
		{name: "cancelled", value: counts.Cancelled},
		{name: "completed", value: counts.Completed},
		{name: "failed", value: counts.Failed},
		{name: "in_progress", value: counts.InProgress},
		{name: "total", value: counts.Total},
	} {
		if count.value < 0 {
			return fail(suite, fmt.Sprintf("vector store file batch file_counts %s is %d, want >= 0", count.name, count.value))
		}
	}
	if counts.Total != counts.Cancelled+counts.Completed+counts.Failed+counts.InProgress {
		return fail(suite, "vector store file batch file_counts total does not match status counts")
	}
	return nil
}

func isVectorStoreFileBatchPreCancelStatusOK(status string) bool {
	return status == "in_progress" || status == "completed"
}

func isVectorStoreFileBatchCancelStatusOK(status string) bool {
	return status == "cancelled" || status == "cancelling" || status == "completed" || status == "in_progress"
}

func isVectorStoreFileBatchCancelAlreadyTerminalError(apiErr *openai.Error) bool {
	if apiErr == nil {
		return false
	}
	switch apiErr.StatusCode {
	case http.StatusConflict, http.StatusBadRequest:
		detail := strings.ToLower(strings.Join([]string{apiErr.Code, apiErr.Message, apiErr.Type}, " "))
		statusIsCompleted := detailContainsWord(detail, "complete", "completed") && !strings.Contains(detail, "not complete")
		statusIsCancelled := detailContainsWord(detail, "cancelled", "canceled")
		statusIsFailed := detailContainsWord(detail, "fail", "failed", "failure")
		terminalSignal := detailContainsWord(detail, "already", "terminal") ||
			(strings.Contains(detail, "cannot") || strings.Contains(detail, "can't") || strings.Contains(detail, "can not"))
		return terminalSignal && !statusIsFailed && (statusIsCompleted || statusIsCancelled)
	default:
		return false
	}
}

func detailContainsWord(detail string, words ...string) bool {
	for _, token := range strings.FieldsFunc(detail, func(r rune) bool {
		return r < 'a' || r > 'z'
	}) {
		for _, word := range words {
			if token == word {
				return true
			}
		}
	}
	return false
}
