package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

// VectorStores verifies the Vector Stores API lifecycle via client.VectorStores.*.
type VectorStores struct{}

func (VectorStores) Name() string { return "vector_stores" }
func (VectorStores) Description() string {
	return "Vector Stores API lifecycle and search (POST/GET/DELETE /v1/vector_stores, POST /v1/vector_stores/{id}/search)"
}

func (VectorStores) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	deleted := false
	var vectorStoreID string
	defer func() {
		if vectorStoreID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.VectorStores.Delete(cleanupCtx, vectorStoreID)
		}
	}()

	created, err := client.VectorStores.New(ctx, openai.VectorStoreNewParams{
		Name: openai.String("compatibility-test-vector-store"),
	})
	if err != nil {
		return fmt.Errorf("vector store create failed: %w", err)
	}
	if err := validateVectorStoreObject("vector_stores", created); err != nil {
		return err
	}
	vectorStoreID = created.ID
	if created.Name != "compatibility-test-vector-store" {
		return fail("vector_stores", fmt.Sprintf("create name is %q, want compatibility-test-vector-store", created.Name))
	}

	got, err := client.VectorStores.Get(ctx, vectorStoreID)
	if err != nil {
		return fmt.Errorf("vector store get failed: %w", err)
	}
	if err := validateVectorStoreObject("vector_stores", got); err != nil {
		return err
	}
	if got.ID != vectorStoreID {
		return fail("vector_stores", fmt.Sprintf("get id is %q, want %q", got.ID, vectorStoreID))
	}

	updated, err := client.VectorStores.Update(ctx, vectorStoreID, openai.VectorStoreUpdateParams{
		Name: openai.String("compatibility-test-vector-store-updated"),
	})
	if err != nil {
		return fmt.Errorf("vector store update failed: %w", err)
	}
	if err := validateVectorStoreObject("vector_stores", updated); err != nil {
		return err
	}
	if updated.ID != vectorStoreID {
		return fail("vector_stores", fmt.Sprintf("update id is %q, want %q", updated.ID, vectorStoreID))
	}
	if updated.Name != "compatibility-test-vector-store-updated" {
		return fail("vector_stores", fmt.Sprintf("update name is %q, want compatibility-test-vector-store-updated", updated.Name))
	}

	listPage, err := client.VectorStores.List(ctx, openai.VectorStoreListParams{
		Limit: openai.Int(10),
	})
	if err != nil {
		return fmt.Errorf("vector store list failed: %w", err)
	}
	if err := validateVectorStoreListPage("vector_stores", listPage); err != nil {
		return err
	}
	found := false
	for _, item := range listPage.Data {
		if item.ID == vectorStoreID {
			found = true
			break
		}
	}
	if !found {
		return fail("vector_stores", "created vector store missing from list response")
	}

	searchPage, err := client.VectorStores.Search(ctx, vectorStoreID, openai.VectorStoreSearchParams{
		Query: openai.VectorStoreSearchParamsQueryUnion{
			OfString: openai.String("compatibility"),
		},
		MaxNumResults: openai.Int(1),
	})
	if err != nil {
		return fmt.Errorf("vector store search failed: %w", err)
	}
	if err := validateVectorStoreSearchPage("vector_stores", searchPage); err != nil {
		return err
	}

	deletedResp, err := client.VectorStores.Delete(ctx, vectorStoreID)
	if err != nil {
		return fmt.Errorf("vector store delete failed: %w", err)
	}
	if deletedResp == nil {
		return fail("vector_stores", "delete response is nil")
	}
	if deletedResp.ID != vectorStoreID {
		return fail("vector_stores", fmt.Sprintf("delete id is %q, want %q", deletedResp.ID, vectorStoreID))
	}
	if !deletedResp.Deleted {
		return fail("vector_stores", "delete response deleted is false")
	}

	_, getErr := client.VectorStores.Get(ctx, vectorStoreID)
	if getErr == nil {
		return fail("vector_stores", "get after delete succeeded; vector store still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("vector_stores", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}
	deleted = true
	return nil
}

func validateVectorStoreObject(suite string, store *openai.VectorStore) error {
	if store == nil {
		return fail(suite, "vector store is nil")
	}
	if store.ID == "" {
		return fail(suite, "vector store missing id")
	}
	if !store.JSON.CreatedAt.Valid() {
		return fail(suite, "vector store missing created_at")
	}
	if !store.JSON.FileCounts.Valid() {
		return fail(suite, "vector store missing file_counts")
	}
	if err := validateVectorStoreFileCounts(suite, store.FileCounts); err != nil {
		return err
	}
	if !store.JSON.LastActiveAt.Valid() {
		return fail(suite, "vector store missing last_active_at")
	}
	if !store.JSON.Metadata.Valid() {
		return fail(suite, "vector store missing metadata")
	}
	if !store.JSON.Name.Valid() {
		return fail(suite, "vector store missing name")
	}
	if !store.JSON.Object.Valid() {
		return fail(suite, "vector store missing object")
	}
	if string(store.Object) != "vector_store" {
		return fail(suite, fmt.Sprintf("vector store object is %q, want vector_store", store.Object))
	}
	if !store.JSON.Status.Valid() {
		return fail(suite, "vector store missing status")
	}
	if store.Status != openai.VectorStoreStatusCompleted && store.Status != openai.VectorStoreStatusInProgress {
		return fail(suite, fmt.Sprintf("vector store status is %q, want completed or in_progress", store.Status))
	}
	if !store.JSON.UsageBytes.Valid() {
		return fail(suite, "vector store missing usage_bytes")
	}
	return nil
}

func validateVectorStoreFileCounts(suite string, counts openai.VectorStoreFileCounts) error {
	if !counts.JSON.Cancelled.Valid() {
		return fail(suite, "vector store file_counts missing cancelled")
	}
	if !counts.JSON.Completed.Valid() {
		return fail(suite, "vector store file_counts missing completed")
	}
	if !counts.JSON.Failed.Valid() {
		return fail(suite, "vector store file_counts missing failed")
	}
	if !counts.JSON.InProgress.Valid() {
		return fail(suite, "vector store file_counts missing in_progress")
	}
	if !counts.JSON.Total.Valid() {
		return fail(suite, "vector store file_counts missing total")
	}
	if counts.Total != counts.Cancelled+counts.Completed+counts.Failed+counts.InProgress {
		return fail(suite, "vector store file_counts total does not match status counts")
	}
	return nil
}

func validateVectorStoreListPage(suite string, page *pagination.CursorPage[openai.VectorStore]) error {
	if page == nil {
		return fail(suite, "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "list missing has_more")
	}
	var envelope struct {
		Object string `json:"object"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	for i := range page.Data {
		if err := validateVectorStoreObject(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateVectorStoreSearchPage(suite string, page *pagination.Page[openai.VectorStoreSearchResponse]) error {
	if page == nil {
		return fail(suite, "search page is nil")
	}
	if !page.JSON.Object.Valid() {
		return fail(suite, "search missing object")
	}
	if page.Object != "vector_store.search_results.page" {
		return fail(suite, fmt.Sprintf("search object is %q, want vector_store.search_results.page", page.Object))
	}
	for _, result := range page.Data {
		if result.FileID == "" {
			return fail(suite, "search result missing file_id")
		}
		if result.Filename == "" {
			return fail(suite, "search result missing filename")
		}
		if !result.JSON.Score.Valid() {
			return fail(suite, "search result missing score")
		}
		for _, content := range result.Content {
			if content.Type != "text" {
				return fail(suite, fmt.Sprintf("search content type is %q, want text", content.Type))
			}
		}
	}
	return nil
}
