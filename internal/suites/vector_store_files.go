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

// VectorStoreFiles verifies vector store file attachment via client.VectorStores.Files.*.
type VectorStoreFiles struct{}

func (VectorStoreFiles) Name() string { return "vector_store_files" }
func (VectorStoreFiles) Description() string {
	return "Vector Store Files API lifecycle (POST/GET/DELETE /v1/vector_stores/{id}/files)"
}

func (VectorStoreFiles) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	store, err := createVectorStoreForSuite(ctx, client, "vector_store_files", "compatibility-test-vector-store-files")
	if err != nil {
		return err
	}
	var fileID string
	defer func() {
		cleanupVectorStoreArtifacts(client, store.ID, fileID)
	}()

	uploaded, err := uploadVectorStoreSourceFile(ctx, client, "vector_store_files")
	if err != nil {
		return err
	}
	fileID = uploaded.ID

	attached, err := client.VectorStores.Files.New(ctx, store.ID, openai.VectorStoreFileNewParams{
		FileID: uploaded.ID,
	})
	if err != nil {
		return fmt.Errorf("vector store file attach failed: %w", err)
	}
	if err := validateVectorStoreFileObject("vector_store_files", attached, store.ID); err != nil {
		return err
	}
	if attached.ID != uploaded.ID {
		return fail("vector_store_files", fmt.Sprintf("attached file id is %q, want %q", attached.ID, uploaded.ID))
	}

	listPage, err := client.VectorStores.Files.List(ctx, store.ID, openai.VectorStoreFileListParams{
		Limit: openai.Int(10),
	})
	if err != nil {
		return fmt.Errorf("vector store file list failed: %w", err)
	}
	if err := validateVectorStoreFileListPage("vector_store_files", listPage, store.ID); err != nil {
		return err
	}
	if !vectorStoreFileListContains(listPage.Data, uploaded.ID) {
		return fail("vector_store_files", "attached file missing from list response")
	}

	got, err := client.VectorStores.Files.Get(ctx, store.ID, uploaded.ID)
	if err != nil {
		return fmt.Errorf("vector store file get failed: %w", err)
	}
	if err := validateVectorStoreFileObject("vector_store_files", got, store.ID); err != nil {
		return err
	}
	if got.ID != uploaded.ID {
		return fail("vector_store_files", fmt.Sprintf("get file id is %q, want %q", got.ID, uploaded.ID))
	}

	deleted, err := client.VectorStores.Files.Delete(ctx, store.ID, uploaded.ID)
	if err != nil {
		return fmt.Errorf("vector store file delete failed: %w", err)
	}
	if deleted == nil {
		return fail("vector_store_files", "delete response is nil")
	}
	if deleted.ID != uploaded.ID {
		return fail("vector_store_files", fmt.Sprintf("delete file id is %q, want %q", deleted.ID, uploaded.ID))
	}
	if !deleted.Deleted {
		return fail("vector_store_files", "delete response deleted is false")
	}

	_, getErr := client.VectorStores.Files.Get(ctx, store.ID, uploaded.ID)
	if getErr == nil {
		return fail("vector_store_files", "get after delete succeeded; vector store file still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("vector_store_files", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}

	sourceFile, err := client.Files.Get(ctx, uploaded.ID)
	if err != nil {
		return fmt.Errorf("source file get after vector store file delete failed: %w", err)
	}
	if err := validateFileObject("vector_store_files", sourceFile); err != nil {
		return err
	}
	if sourceFile.ID != uploaded.ID {
		return fail("vector_store_files", fmt.Sprintf("source file id after vector store file delete is %q, want %q", sourceFile.ID, uploaded.ID))
	}
	return nil
}

func createVectorStoreForSuite(ctx context.Context, client openai.Client, suite, name string) (*openai.VectorStore, error) {
	store, err := client.VectorStores.New(ctx, openai.VectorStoreNewParams{
		Name: openai.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: vector store create failed: %w", suite, err)
	}
	if err := validateVectorStoreObject(suite, store); err != nil {
		return nil, err
	}
	return store, nil
}

func uploadVectorStoreSourceFile(ctx context.Context, client openai.Client, suite string) (*openai.FileObject, error) {
	uploaded, err := client.Files.New(ctx, openai.FileNewParams{
		File:    smallTextFileReader(),
		Purpose: openai.FilePurposeAssistants,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: source file upload failed: %w", suite, err)
	}
	if err := validateFileObject(suite, uploaded); err != nil {
		cleanupVectorStoreUploadedFile(client, uploaded)
		return nil, err
	}
	if string(uploaded.Purpose) != string(openai.FilePurposeAssistants) {
		cleanupVectorStoreUploadedFile(client, uploaded)
		return nil, fail(suite, fmt.Sprintf("upload purpose is %q, want assistants", uploaded.Purpose))
	}
	return uploaded, nil
}

func cleanupVectorStoreUploadedFile(client openai.Client, file *openai.FileObject) {
	if file != nil {
		cleanupVectorStoreArtifacts(client, "", file.ID)
	}
}

func cleanupVectorStoreArtifacts(client openai.Client, vectorStoreID string, fileIDs ...string) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if vectorStoreID != "" {
		_, _ = client.VectorStores.Delete(cleanupCtx, vectorStoreID)
	}
	for _, fileID := range fileIDs {
		if fileID != "" {
			_, _ = client.Files.Delete(cleanupCtx, fileID)
		}
	}
}

func validateVectorStoreFileObject(suite string, file *openai.VectorStoreFile, expectedVectorStoreID string) error {
	if file == nil {
		return fail(suite, "vector store file is nil")
	}
	if file.ID == "" {
		return fail(suite, "vector store file missing id")
	}
	if !file.JSON.CreatedAt.Valid() {
		return fail(suite, "vector store file missing created_at")
	}
	if !file.JSON.Object.Valid() {
		return fail(suite, "vector store file missing object")
	}
	if string(file.Object) != "vector_store.file" {
		return fail(suite, fmt.Sprintf("vector store file object is %q, want vector_store.file", file.Object))
	}
	if !file.JSON.Status.Valid() {
		return fail(suite, "vector store file missing status")
	}
	if !isVectorStoreFileStatusOK(file.Status) {
		return fail(suite, fmt.Sprintf("vector store file status is %q, want in_progress, completed, or cancelled", file.Status))
	}
	if file.JSON.LastError.Raw() == "" {
		return fail(suite, "vector store file missing last_error")
	}
	if !file.JSON.UsageBytes.Valid() {
		return fail(suite, "vector store file missing usage_bytes")
	}
	if file.UsageBytes < 0 {
		return fail(suite, fmt.Sprintf("vector store file usage_bytes is %d, want >= 0", file.UsageBytes))
	}
	if !file.JSON.VectorStoreID.Valid() {
		return fail(suite, "vector store file missing vector_store_id")
	}
	if file.VectorStoreID == "" {
		return fail(suite, "vector store file vector_store_id is empty")
	}
	if expectedVectorStoreID != "" && file.VectorStoreID != expectedVectorStoreID {
		return fail(suite, fmt.Sprintf("vector store file vector_store_id is %q, want %q", file.VectorStoreID, expectedVectorStoreID))
	}
	return nil
}

func isVectorStoreFileStatusOK(status openai.VectorStoreFileStatus) bool {
	switch status {
	case openai.VectorStoreFileStatusInProgress, openai.VectorStoreFileStatusCompleted, openai.VectorStoreFileStatusCancelled:
		return true
	default:
		return false
	}
}

func validateVectorStoreFileListPage(suite string, page *pagination.CursorPage[openai.VectorStoreFile], expectedVectorStoreID string) error {
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
		if err := validateVectorStoreFileObject(suite, &page.Data[i], expectedVectorStoreID); err != nil {
			return err
		}
	}
	return nil
}

func vectorStoreFileListContains(files []openai.VectorStoreFile, fileID string) bool {
	for _, file := range files {
		if file.ID == fileID {
			return true
		}
	}
	return false
}
