package suites

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

// Files verifies the Files API lifecycle via client.Files.*.
type Files struct{}

func (Files) Name() string { return "files" }
func (Files) Description() string {
	return "Files API lifecycle (POST/GET/DELETE /v1/files, GET /v1/files/{id}/content)"
}

func (Files) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	deleted := false
	var fileID string
	defer func() {
		if fileID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Files.Delete(cleanupCtx, fileID)
		}
	}()

	uploaded, err := client.Files.New(ctx, openai.FileNewParams{
		File:    smallTextFileReader(),
		Purpose: openai.FilePurposeUserData,
	})
	if err != nil {
		return fmt.Errorf("file upload failed: %w", err)
	}
	if err := validateFileObject("files", uploaded); err != nil {
		return err
	}
	fileID = uploaded.ID

	listPage, err := client.Files.List(ctx, openai.FileListParams{})
	if err != nil {
		return fmt.Errorf("file list failed: %w", err)
	}
	if err := validateFileListPage("files", listPage); err != nil {
		return err
	}
	found := false
	for _, item := range listPage.Data {
		if item.ID == fileID {
			found = true
			break
		}
	}
	if !found {
		return fail("files", "uploaded file missing from list response")
	}

	got, err := client.Files.Get(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file get failed: %w", err)
	}
	if err := validateFileObject("files", got); err != nil {
		return err
	}
	if got.ID != fileID {
		return fail("files", fmt.Sprintf("get id is %q, want %q", got.ID, fileID))
	}
	if got.Filename != "test.txt" {
		return fail("files", fmt.Sprintf("get filename is %q, want test.txt", got.Filename))
	}
	if string(got.Purpose) != string(openai.FilePurposeUserData) {
		return fail("files", fmt.Sprintf("get purpose is %q, want user_data", got.Purpose))
	}

	contentResp, err := client.Files.Content(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file content failed: %w", err)
	}
	if err := validateFileContentResponse("files", contentResp, smallTextFileBytes()); err != nil {
		return err
	}

	deletedResp, err := client.Files.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file delete failed: %w", err)
	}
	if deletedResp == nil {
		return fail("files", "delete response is nil")
	}
	if deletedResp.ID != fileID {
		return fail("files", fmt.Sprintf("delete id is %q, want %q", deletedResp.ID, fileID))
	}
	if !deletedResp.Deleted {
		return fail("files", "delete response deleted is false")
	}

	_, getErr := client.Files.Get(ctx, fileID)
	if getErr == nil {
		return fail("files", "get after delete succeeded; file still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("files", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}
	deleted = true
	return nil
}

func validateFileObject(suite string, file *openai.FileObject) error {
	if err := validateFileEnvelope(suite, file); err != nil {
		return err
	}
	if !file.JSON.Status.Valid() {
		return fail(suite, "file missing status")
	}
	return nil
}

func validateFileEnvelope(suite string, file *openai.FileObject) error {
	if file == nil {
		return fail(suite, "file is nil")
	}
	if file.ID == "" {
		return fail(suite, "file missing id")
	}
	if !file.JSON.Bytes.Valid() {
		return fail(suite, "file missing bytes")
	}
	if !file.JSON.CreatedAt.Valid() {
		return fail(suite, "file missing created_at")
	}
	if file.Filename == "" {
		return fail(suite, "file missing filename")
	}
	if string(file.Object) != "file" {
		return fail(suite, fmt.Sprintf("file object is %q, want file", file.Object))
	}
	if file.Purpose == "" {
		return fail(suite, "file missing purpose")
	}
	return nil
}

func validateFileListPage(suite string, page *pagination.CursorPage[openai.FileObject]) error {
	return validateCursorListPage(suite, page, nil)
}

func validateFileContentResponse(suite string, resp *http.Response, want []byte) error {
	if resp == nil {
		return fail(suite, "content response is nil")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fail(suite, fmt.Sprintf("content status is %d, want 200", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: read content body: %w", suite, err)
	}
	if len(body) != len(want) {
		return fail(suite, fmt.Sprintf("content has %d bytes, want %d", len(body), len(want)))
	}
	for i := range want {
		if body[i] != want[i] {
			return fail(suite, "content body does not match uploaded file")
		}
	}
	return nil
}