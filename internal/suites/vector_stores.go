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

const (
	vectorStoreCreateName = "compatibility-test-vector-store"
	vectorStoreUpdateName = "compatibility-test-vector-store-updated"
)

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
		Name: openai.String(vectorStoreCreateName),
	})
	if err != nil {
		return fmt.Errorf("vector store create failed: %w", err)
	}
	if created != nil && created.ID != "" {
		vectorStoreID = created.ID
	}
	if err := validateVectorStoreObject("vector_stores", created); err != nil {
		return err
	}
	if created.Name != vectorStoreCreateName {
		return fail("vector_stores", fmt.Sprintf("create name is %q, want %q", created.Name, vectorStoreCreateName))
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
	if got.Name != vectorStoreCreateName {
		return fail("vector_stores", fmt.Sprintf("get name is %q, want %q", got.Name, vectorStoreCreateName))
	}

	updated, err := client.VectorStores.Update(ctx, vectorStoreID, openai.VectorStoreUpdateParams{
		Name: openai.String(vectorStoreUpdateName),
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
	if updated.Name != vectorStoreUpdateName {
		return fail("vector_stores", fmt.Sprintf("update name is %q, want %q", updated.Name, vectorStoreUpdateName))
	}

	listPage, err := client.VectorStores.List(ctx, openai.VectorStoreListParams{
		Limit: openai.Int(100),
		Order: openai.VectorStoreListParamsOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("vector store list failed: %w", err)
	}
	found, err := vectorStoreListContains(listPage, vectorStoreID)
	if err != nil {
		return err
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
	if err := validateVectorStoreDeleteResponse("vector_stores", deletedResp, vectorStoreID); err != nil {
		return err
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

func vectorStoreListContains(page *pagination.CursorPage[openai.VectorStore], vectorStoreID string) (bool, error) {
	for page != nil {
		if err := validateVectorStoreListPage("vector_stores", page); err != nil {
			return false, err
		}
		for _, item := range page.Data {
			if item.ID == vectorStoreID {
				return true, nil
			}
		}
		next, err := page.GetNextPage()
		if err != nil {
			return false, fmt.Errorf("vector store list next page failed: %w", err)
		}
		page = next
	}
	return false, nil
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
	if !isVectorStoreStatusOK(store.Status) {
		return fail(suite, fmt.Sprintf("vector store status is %q, want completed, in_progress, or expired", store.Status))
	}
	if !store.JSON.UsageBytes.Valid() {
		return fail(suite, "vector store missing usage_bytes")
	}
	if store.UsageBytes < 0 {
		return fail(suite, fmt.Sprintf("vector store usage_bytes is %d, want >= 0", store.UsageBytes))
	}
	return nil
}

func isVectorStoreStatusOK(status openai.VectorStoreStatus) bool {
	switch status {
	case openai.VectorStoreStatusCompleted, openai.VectorStoreStatusInProgress, openai.VectorStoreStatusExpired:
		return true
	default:
		return false
	}
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
			return fail(suite, fmt.Sprintf("vector store file_counts %s is %d, want >= 0", count.name, count.value))
		}
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
		Object  string `json:"object"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	if len(page.Data) == 0 {
		return nil
	}
	if envelope.FirstID == "" {
		return fail(suite, "list missing first_id")
	}
	if envelope.LastID == "" {
		return fail(suite, "list missing last_id")
	}
	if envelope.FirstID != page.Data[0].ID {
		return fail(suite, fmt.Sprintf("list first_id is %q, want %q", envelope.FirstID, page.Data[0].ID))
	}
	if envelope.LastID != page.Data[len(page.Data)-1].ID {
		return fail(suite, fmt.Sprintf("list last_id is %q, want %q", envelope.LastID, page.Data[len(page.Data)-1].ID))
	}
	for i := range page.Data {
		if err := validateVectorStoreObject(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateVectorStoreDeleteResponse(suite string, deleted *openai.VectorStoreDeleted, wantID string) error {
	if deleted.ID != wantID {
		return fail(suite, fmt.Sprintf("delete id is %q, want %q", deleted.ID, wantID))
	}
	if !deleted.Deleted {
		return fail(suite, "delete response deleted is false")
	}
	if !deleted.JSON.Object.Valid() {
		return fail(suite, "delete response missing object")
	}
	if string(deleted.Object) != "vector_store.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want vector_store.deleted", deleted.Object))
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
	if !page.JSON.Data.Valid() {
		return fail(suite, "search missing data")
	}
	if page.Object != "vector_store.search_results.page" {
		return fail(suite, fmt.Sprintf("search object is %q, want vector_store.search_results.page", page.Object))
	}
	for _, result := range page.Data {
		if !result.JSON.Attributes.Valid() {
			return fail(suite, "search result missing attributes")
		}
		if !result.JSON.Content.Valid() {
			return fail(suite, "search result missing content")
		}
		if result.FileID == "" {
			return fail(suite, "search result missing file_id")
		}
		if result.Filename == "" {
			return fail(suite, "search result missing filename")
		}
		if !result.JSON.Score.Valid() {
			return fail(suite, "search result missing score")
		}
		if result.Score < 0 || result.Score > 1 {
			return fail(suite, fmt.Sprintf("search result score is %g, want between 0 and 1", result.Score))
		}
		for _, content := range result.Content {
			if content.Type != "text" {
				return fail(suite, fmt.Sprintf("search content type is %q, want text", content.Type))
			}
			if !content.JSON.Text.Valid() {
				return fail(suite, "search content missing text")
			}
		}
	}
	return nil
}
