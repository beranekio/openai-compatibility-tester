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
