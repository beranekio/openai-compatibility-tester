package suites

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

func TestValidateVectorStoreFileObjectRejectsUnknownStatus(t *testing.T) {
	var file openai.VectorStoreFile
	raw := `{
		"id": "file_mock",
		"object": "vector_store.file",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "bogus",
		"last_error": null,
		"usage_bytes": 24
	}`
	if err := json.Unmarshal([]byte(raw), &file); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileObject("vector_store_files", &file, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), `status is "bogus"`) {
		t.Fatalf("expected bogus status validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileObjectRequiresLastError(t *testing.T) {
	var file openai.VectorStoreFile
	raw := `{
		"id": "file_mock",
		"object": "vector_store.file",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "completed",
		"usage_bytes": 24
	}`
	if err := json.Unmarshal([]byte(raw), &file); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileObject("vector_store_files", &file, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), "missing last_error") {
		t.Fatalf("expected missing last_error validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileObjectRejectsWrongVectorStoreID(t *testing.T) {
	var file openai.VectorStoreFile
	raw := `{
		"id": "file_mock",
		"object": "vector_store.file",
		"created_at": 1700000000,
		"vector_store_id": "vs_other",
		"status": "completed",
		"last_error": null,
		"usage_bytes": 24
	}`
	if err := json.Unmarshal([]byte(raw), &file); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileObject("vector_store_files", &file, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), `vector_store_id is "vs_other"`) {
		t.Fatalf("expected wrong vector store id validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileObjectRejectsNegativeUsageBytes(t *testing.T) {
	var file openai.VectorStoreFile
	raw := `{
		"id": "file_mock",
		"object": "vector_store.file",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "completed",
		"last_error": null,
		"usage_bytes": -1
	}`
	if err := json.Unmarshal([]byte(raw), &file); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileObject("vector_store_files", &file, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), "usage_bytes is -1") {
		t.Fatalf("expected negative usage_bytes validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileObjectRejectsCancelledStatus(t *testing.T) {
	var file openai.VectorStoreFile
	raw := `{
		"id": "file_mock",
		"object": "vector_store.file",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "cancelled",
		"last_error": null,
		"usage_bytes": 24
	}`
	if err := json.Unmarshal([]byte(raw), &file); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileObject("vector_store_files", &file, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), `status is "cancelled"`) {
		t.Fatalf("expected cancelled status validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileDeleteResponseRequiresObject(t *testing.T) {
	var deleted openai.VectorStoreFileDeleted
	if err := json.Unmarshal([]byte(`{"id":"file_mock","deleted":true}`), &deleted); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileDeleteResponse("vector_store_files", &deleted, "file_mock")
	if err == nil || !strings.Contains(err.Error(), "delete response missing object") {
		t.Fatalf("expected missing object validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileDeleteResponseRejectsWrongObject(t *testing.T) {
	var deleted openai.VectorStoreFileDeleted
	if err := json.Unmarshal([]byte(`{"id":"file_mock","object":"file.deleted","deleted":true}`), &deleted); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileDeleteResponse("vector_store_files", &deleted, "file_mock")
	if err == nil || !strings.Contains(err.Error(), `delete object is "file.deleted"`) {
		t.Fatalf("expected wrong object validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileListPageRequiresCursorIDs(t *testing.T) {
	page := parseVectorStoreFileListPage(t, `{
		"object": "list",
		"data": [{
			"id": "file_mock",
			"object": "vector_store.file",
			"created_at": 1700000000,
			"vector_store_id": "vs_mock",
			"status": "completed",
			"last_error": null,
			"usage_bytes": 24
		}],
		"has_more": false
	}`)

	err := validateVectorStoreFileListPage("vector_store_files", page, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), "list missing first_id") {
		t.Fatalf("expected missing first_id validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileListPageRejectsWrongLastID(t *testing.T) {
	page := parseVectorStoreFileListPage(t, `{
		"object": "list",
		"data": [{
			"id": "file_mock",
			"object": "vector_store.file",
			"created_at": 1700000000,
			"vector_store_id": "vs_mock",
			"status": "completed",
			"last_error": null,
			"usage_bytes": 24
		}],
		"first_id": "file_mock",
		"last_id": "file_other",
		"has_more": false
	}`)

	err := validateVectorStoreFileListPage("vector_store_files", page, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), `list last_id is "file_other"`) {
		t.Fatalf("expected wrong last_id validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileBatchObjectAcceptsSingularObject(t *testing.T) {
	var batch openai.VectorStoreFileBatch
	raw := `{
		"id": "vsfb_mock",
		"object": "vector_store.file_batch",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "completed",
		"file_counts": {
			"in_progress": 0,
			"completed": 1,
			"failed": 0,
			"cancelled": 0,
			"total": 1
		}
	}`
	if err := json.Unmarshal([]byte(raw), &batch); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if err := validateVectorStoreFileBatchObject("vector_store_file_batches", &batch, "vs_mock"); err != nil {
		t.Fatalf("validateVectorStoreFileBatchObject() error = %v", err)
	}
}

func TestValidateVectorStoreFileBatchObjectRejectsUnknownStatus(t *testing.T) {
	var batch openai.VectorStoreFileBatch
	raw := `{
		"id": "vsfb_mock",
		"object": "vector_store.files_batch",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "bogus",
		"file_counts": {
			"in_progress": 0,
			"completed": 1,
			"failed": 0,
			"cancelled": 0,
			"total": 1
		}
	}`
	if err := json.Unmarshal([]byte(raw), &batch); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileBatchObject("vector_store_file_batches", &batch, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), `status is "bogus"`) {
		t.Fatalf("expected bogus status validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileBatchObjectRejectsWrongVectorStoreID(t *testing.T) {
	var batch openai.VectorStoreFileBatch
	raw := `{
		"id": "vsfb_mock",
		"object": "vector_store.files_batch",
		"created_at": 1700000000,
		"vector_store_id": "vs_other",
		"status": "completed",
		"file_counts": {
			"in_progress": 0,
			"completed": 1,
			"failed": 0,
			"cancelled": 0,
			"total": 1
		}
	}`
	if err := json.Unmarshal([]byte(raw), &batch); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileBatchObject("vector_store_file_batches", &batch, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), `vector_store_id is "vs_other"`) {
		t.Fatalf("expected wrong vector store id validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileBatchObjectRejectsNegativeFileCounts(t *testing.T) {
	var batch openai.VectorStoreFileBatch
	raw := `{
		"id": "vsfb_mock",
		"object": "vector_store.files_batch",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "completed",
		"file_counts": {
			"in_progress": 0,
			"completed": 2,
			"failed": -1,
			"cancelled": 0,
			"total": 1
		}
	}`
	if err := json.Unmarshal([]byte(raw), &batch); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileBatchObject("vector_store_file_batches", &batch, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), "file_counts failed is -1") {
		t.Fatalf("expected negative file_counts validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileBatchUploadCountsRejectsWrongTotal(t *testing.T) {
	var batch openai.VectorStoreFileBatch
	raw := `{
		"id": "vsfb_mock",
		"object": "vector_store.files_batch",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "completed",
		"file_counts": {
			"in_progress": 0,
			"completed": 1,
			"failed": 0,
			"cancelled": 0,
			"total": 1
		}
	}`
	if err := json.Unmarshal([]byte(raw), &batch); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileBatchUploadCounts("vector_store_file_batches", &batch, 2)
	if err == nil || !strings.Contains(err.Error(), "file_counts.total is 1, want 2") {
		t.Fatalf("expected wrong total validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileBatchUploadCountsRejectsFailedFiles(t *testing.T) {
	var batch openai.VectorStoreFileBatch
	raw := `{
		"id": "vsfb_mock",
		"object": "vector_store.files_batch",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "completed",
		"file_counts": {
			"in_progress": 0,
			"completed": 1,
			"failed": 1,
			"cancelled": 0,
			"total": 2
		}
	}`
	if err := json.Unmarshal([]byte(raw), &batch); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileBatchUploadCounts("vector_store_file_batches", &batch, 2)
	if err == nil || !strings.Contains(err.Error(), "file_counts.failed is 1") {
		t.Fatalf("expected failed file count validation error, got %v", err)
	}
}

func TestValidateVectorStoreFileBatchPreCancelStatusRejectsCancelled(t *testing.T) {
	var batch openai.VectorStoreFileBatch
	raw := `{
		"id": "vsfb_mock",
		"object": "vector_store.files_batch",
		"created_at": 1700000000,
		"vector_store_id": "vs_mock",
		"status": "cancelled",
		"file_counts": {
			"in_progress": 0,
			"completed": 0,
			"failed": 0,
			"cancelled": 1,
			"total": 1
		}
	}`
	if err := json.Unmarshal([]byte(raw), &batch); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreFileBatchPreCancelStatus("vector_store_file_batches", &batch)
	if err == nil || !strings.Contains(err.Error(), `status is "cancelled"`) {
		t.Fatalf("expected cancelled pre-cancel status validation error, got %v", err)
	}
}

func TestVectorStoreFileBatchCancelStatusAllowsInProgress(t *testing.T) {
	for _, status := range []string{"in_progress", "cancelled", "cancelling", "completed"} {
		if !isVectorStoreFileBatchCancelStatusOK(status) {
			t.Fatalf("cancel status %q was rejected", status)
		}
	}
}

func TestVectorStoreFileBatchCancelAlreadyTerminalErrorRequiresTerminalSignal(t *testing.T) {
	genericErr := &openai.Error{
		StatusCode: http.StatusBadRequest,
		Type:       "invalid_request_error",
		Message:    "unsupported endpoint",
	}
	if isVectorStoreFileBatchCancelAlreadyTerminalError(genericErr) {
		t.Fatal("generic invalid_request_error was treated as already-terminal")
	}

	terminalErr := &openai.Error{
		StatusCode: http.StatusConflict,
		Type:       "invalid_request_error",
		Code:       "batch_already_completed",
		Message:    "Batch is already completed",
	}
	if !isVectorStoreFileBatchCancelAlreadyTerminalError(terminalErr) {
		t.Fatal("already-completed conflict was not treated as already-terminal")
	}

	failedErr := &openai.Error{
		StatusCode: http.StatusConflict,
		Type:       "invalid_request_error",
		Code:       "batch_already_failed",
		Message:    "Batch is already failed",
	}
	if isVectorStoreFileBatchCancelAlreadyTerminalError(failedErr) {
		t.Fatal("already-failed conflict was treated as already-terminal")
	}

	incompleteErr := &openai.Error{
		StatusCode: http.StatusBadRequest,
		Type:       "invalid_request_error",
		Message:    "Cannot cancel an incomplete batch",
	}
	if isVectorStoreFileBatchCancelAlreadyTerminalError(incompleteErr) {
		t.Fatal("incomplete cancel error was treated as already-terminal")
	}

	notCompleteErr := &openai.Error{
		StatusCode: http.StatusBadRequest,
		Type:       "invalid_request_error",
		Message:    "Cannot cancel a batch that is not complete",
	}
	if isVectorStoreFileBatchCancelAlreadyTerminalError(notCompleteErr) {
		t.Fatal("not-complete cancel error was treated as already-terminal")
	}
}

func parseVectorStoreFileListPage(t *testing.T, raw string) *pagination.CursorPage[openai.VectorStoreFile] {
	t.Helper()

	var page pagination.CursorPage[openai.VectorStoreFile]
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return &page
}
