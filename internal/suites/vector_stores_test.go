package suites

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

func TestValidateVectorStoreObjectAllowsExpiredWithoutOptionalFields(t *testing.T) {
	var store openai.VectorStore
	raw := `{
		"id": "vs_expired",
		"object": "vector_store",
		"created_at": 1700000000,
		"name": "expired vector store",
		"status": "expired",
		"usage_bytes": 0,
		"file_counts": {
			"in_progress": 0,
			"completed": 0,
			"failed": 0,
			"cancelled": 0,
			"total": 0
		}
	}`
	if err := json.Unmarshal([]byte(raw), &store); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if err := validateVectorStoreObject("vector_stores", &store); err != nil {
		t.Fatalf("validateVectorStoreObject() error = %v", err)
	}
}

func TestValidateVectorStoreSearchPageRequiresData(t *testing.T) {
	var page pagination.Page[openai.VectorStoreSearchResponse]
	if err := json.Unmarshal([]byte(`{"object":"vector_store.search_results.page"}`), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreSearchPage("vector_stores", &page)
	if err == nil || !strings.Contains(err.Error(), "search missing data") {
		t.Fatalf("expected missing data validation error, got %v", err)
	}
}

func TestValidateVectorStoreSearchPageAllowsEmptyData(t *testing.T) {
	var page pagination.Page[openai.VectorStoreSearchResponse]
	if err := json.Unmarshal([]byte(`{"object":"vector_store.search_results.page","data":[]}`), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if err := validateVectorStoreSearchPage("vector_stores", &page); err != nil {
		t.Fatalf("validateVectorStoreSearchPage() error = %v", err)
	}
}

func TestValidateVectorStoreListPageRequiresFirstIDWithData(t *testing.T) {
	page := parseVectorStoreListPage(t, `{
		"object": "list",
		"data": [{
			"id": "vs_mock",
			"object": "vector_store",
			"created_at": 1700000000,
			"name": "mock vector store",
			"status": "completed",
			"usage_bytes": 0,
			"file_counts": {
				"in_progress": 0,
				"completed": 0,
				"failed": 0,
				"cancelled": 0,
				"total": 0
			}
		}],
		"last_id": "vs_mock",
		"has_more": false
	}`)

	err := validateVectorStoreListPage("vector_stores", page)
	if err == nil || !strings.Contains(err.Error(), "list missing first_id") {
		t.Fatalf("expected missing first_id validation error, got %v", err)
	}
}

func TestValidateVectorStoreListPageRequiresLastIDWithData(t *testing.T) {
	page := parseVectorStoreListPage(t, `{
		"object": "list",
		"data": [{
			"id": "vs_mock",
			"object": "vector_store",
			"created_at": 1700000000,
			"name": "mock vector store",
			"status": "completed",
			"usage_bytes": 0,
			"file_counts": {
				"in_progress": 0,
				"completed": 0,
				"failed": 0,
				"cancelled": 0,
				"total": 0
			}
		}],
		"first_id": "vs_mock",
		"has_more": false
	}`)

	err := validateVectorStoreListPage("vector_stores", page)
	if err == nil || !strings.Contains(err.Error(), "list missing last_id") {
		t.Fatalf("expected missing last_id validation error, got %v", err)
	}
}

func TestValidateVectorStoreObjectRejectsNegativeFileCounts(t *testing.T) {
	var store openai.VectorStore
	raw := `{
		"id": "vs_negative_counts",
		"object": "vector_store",
		"created_at": 1700000000,
		"name": "negative counts",
		"status": "completed",
		"usage_bytes": 0,
		"file_counts": {
			"in_progress": 0,
			"completed": 1,
			"failed": 0,
			"cancelled": -1,
			"total": 0
		}
	}`
	if err := json.Unmarshal([]byte(raw), &store); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreObject("vector_stores", &store)
	if err == nil || !strings.Contains(err.Error(), "file_counts cancelled is -1") {
		t.Fatalf("expected negative count validation error, got %v", err)
	}
}

func TestValidateVectorStoreObjectRejectsNegativeUsageBytes(t *testing.T) {
	var store openai.VectorStore
	raw := `{
		"id": "vs_negative_usage",
		"object": "vector_store",
		"created_at": 1700000000,
		"name": "negative usage",
		"status": "completed",
		"usage_bytes": -1,
		"file_counts": {
			"in_progress": 0,
			"completed": 0,
			"failed": 0,
			"cancelled": 0,
			"total": 0
		}
	}`
	if err := json.Unmarshal([]byte(raw), &store); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreObject("vector_stores", &store)
	if err == nil || !strings.Contains(err.Error(), "usage_bytes is -1") {
		t.Fatalf("expected negative usage_bytes validation error, got %v", err)
	}
}

func TestValidateVectorStoreDeleteResponseRequiresObject(t *testing.T) {
	var deleted openai.VectorStoreDeleted
	if err := json.Unmarshal([]byte(`{"id":"vs_mock","deleted":true}`), &deleted); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreDeleteResponse("vector_stores", &deleted, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), "delete response missing object") {
		t.Fatalf("expected missing object validation error, got %v", err)
	}
}

func TestValidateVectorStoreDeleteResponseRejectsWrongObject(t *testing.T) {
	var deleted openai.VectorStoreDeleted
	if err := json.Unmarshal([]byte(`{"id":"vs_mock","object":"wrong","deleted":true}`), &deleted); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreDeleteResponse("vector_stores", &deleted, "vs_mock")
	if err == nil || !strings.Contains(err.Error(), `delete object is "wrong"`) {
		t.Fatalf("expected wrong object validation error, got %v", err)
	}
}

func TestValidateVectorStoreSearchPageRequiresResultAttributes(t *testing.T) {
	var page pagination.Page[openai.VectorStoreSearchResponse]
	raw := `{
		"object": "vector_store.search_results.page",
		"data": [{
			"file_id": "file_mock",
			"filename": "test.txt",
			"score": 0.7,
			"content": [{"type": "text", "text": "hello"}]
		}]
	}`
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreSearchPage("vector_stores", &page)
	if err == nil || !strings.Contains(err.Error(), "search result missing attributes") {
		t.Fatalf("expected missing attributes validation error, got %v", err)
	}
}

func TestValidateVectorStoreSearchPageRequiresResultContent(t *testing.T) {
	var page pagination.Page[openai.VectorStoreSearchResponse]
	raw := `{
		"object": "vector_store.search_results.page",
		"data": [{
			"file_id": "file_mock",
			"filename": "test.txt",
			"score": 0.7,
			"attributes": {}
		}]
	}`
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreSearchPage("vector_stores", &page)
	if err == nil || !strings.Contains(err.Error(), "search result missing content") {
		t.Fatalf("expected missing content validation error, got %v", err)
	}
}

func TestValidateVectorStoreSearchPageRequiresContentText(t *testing.T) {
	var page pagination.Page[openai.VectorStoreSearchResponse]
	raw := `{
		"object": "vector_store.search_results.page",
		"data": [{
			"file_id": "file_mock",
			"filename": "test.txt",
			"score": 0.7,
			"attributes": {},
			"content": [{"type": "text"}]
		}]
	}`
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreSearchPage("vector_stores", &page)
	if err == nil || !strings.Contains(err.Error(), "search content missing text") {
		t.Fatalf("expected missing content text validation error, got %v", err)
	}
}

func TestValidateVectorStoreSearchPageRejectsOutOfRangeScore(t *testing.T) {
	var page pagination.Page[openai.VectorStoreSearchResponse]
	raw := `{
		"object": "vector_store.search_results.page",
		"data": [{
			"file_id": "file_mock",
			"filename": "test.txt",
			"score": 7,
			"attributes": {},
			"content": [{"type": "text", "text": "hello"}]
		}]
	}`
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	err := validateVectorStoreSearchPage("vector_stores", &page)
	if err == nil || !strings.Contains(err.Error(), "score is 7") {
		t.Fatalf("expected score bounds validation error, got %v", err)
	}
}

func parseVectorStoreListPage(t *testing.T, raw string) *pagination.CursorPage[openai.VectorStore] {
	t.Helper()

	var page pagination.CursorPage[openai.VectorStore]
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return &page
}
