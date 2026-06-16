package suites

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
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
}
