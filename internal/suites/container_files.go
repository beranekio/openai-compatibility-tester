package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

const (
	containerFileCreateName = "compatibility-test-container-files"
	containerFilePath       = "/test.txt"
)

// ContainerFiles verifies container file lifecycle via client.Containers.Files.*.
type ContainerFiles struct{}

func (ContainerFiles) Name() string { return "container_files" }
func (ContainerFiles) Description() string {
	return "Container Files API lifecycle (POST/GET/DELETE /v1/containers/{id}/files, GET /v1/containers/{id}/files/{file_id}/content)"
}

func (ContainerFiles) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	deleted := false
	var containerID, fileID string
	defer func() {
		if containerID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = client.Containers.Delete(cleanupCtx, containerID)
		}
	}()

	created, err := createContainerForSuite(ctx, client, "container_files", containerFileCreateName)
	if err != nil {
		return err
	}
	containerID = created.ID

	uploaded, err := client.Containers.Files.New(ctx, containerID, openai.ContainerFileNewParams{
		File: smallTextFileReader(),
	})
	if err != nil {
		return fmt.Errorf("container file upload failed: %w", err)
	}
	if err := validateContainerFileObject("container_files", uploaded, containerID); err != nil {
		return err
	}
	fileID = uploaded.ID
	if uploaded.Path != containerFilePath {
		return fail("container_files", fmt.Sprintf("upload path is %q, want %q", uploaded.Path, containerFilePath))
	}
	if uploaded.Source != "user" {
		return fail("container_files", fmt.Sprintf("upload source is %q, want user", uploaded.Source))
	}
	if uploaded.Bytes != int64(len(smallTextFileBytes())) {
		return fail("container_files", fmt.Sprintf("upload bytes is %d, want %d", uploaded.Bytes, len(smallTextFileBytes())))
	}

	listPage, err := client.Containers.Files.List(ctx, containerID, openai.ContainerFileListParams{
		Limit: openai.Int(10),
		Order: openai.ContainerFileListParamsOrderAsc,
	})
	if err != nil {
		return fmt.Errorf("container file list failed: %w", err)
	}
	if err := validateContainerFileListPage("container_files", listPage, containerID); err != nil {
		return err
	}
	if !containerFileListContainsID(listPage.Data, fileID) {
		return fail("container_files", "uploaded file missing from list response")
	}

	got, err := client.Containers.Files.Get(ctx, containerID, fileID)
	if err != nil {
		return fmt.Errorf("container file get failed: %w", err)
	}
	if err := validateContainerFileGetObject("container_files", got, containerID); err != nil {
		return err
	}
	if got.ID != fileID {
		return fail("container_files", fmt.Sprintf("get id is %q, want %q", got.ID, fileID))
	}
	if got.Path != containerFilePath {
		return fail("container_files", fmt.Sprintf("get path is %q, want %q", got.Path, containerFilePath))
	}

	contentResp, err := client.Containers.Files.Content.Get(ctx, containerID, fileID)
	if err != nil {
		return fmt.Errorf("container file content failed: %w", err)
	}
	if err := validateContainerFileContentResponse("container_files", contentResp, smallTextFileBytes()); err != nil {
		return err
	}

	if err := client.Containers.Files.Delete(ctx, containerID, fileID); err != nil {
		return fmt.Errorf("container file delete failed: %w", err)
	}

	_, getErr := client.Containers.Files.Get(ctx, containerID, fileID)
	if getErr == nil {
		return fail("container_files", "get after delete succeeded; container file still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("container_files", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}

	if err := client.Containers.Delete(ctx, containerID); err != nil {
		return fmt.Errorf("container delete failed: %w", err)
	}
	_, containerGetErr := client.Containers.Get(ctx, containerID)
	if containerGetErr == nil {
		return fail("container_files", "container get after delete succeeded; container still exists")
	}
	var containerAPIErr *openai.Error
	if !errors.As(containerGetErr, &containerAPIErr) {
		return fmt.Errorf("container get after delete failed: %w", containerGetErr)
	}
	if containerAPIErr.StatusCode != http.StatusNotFound {
		return fail("container_files", fmt.Sprintf("container get after delete returned status %d, want 404", containerAPIErr.StatusCode))
	}
	deleted = true
	return nil
}

func validateContainerFileObject(suite string, file *openai.ContainerFileNewResponse, expectedContainerID string) error {
	if file == nil {
		return fail(suite, "container file is nil")
	}
	if file.ID == "" {
		return fail(suite, "container file missing id")
	}
	if !file.JSON.Bytes.Valid() {
		return fail(suite, "container file missing bytes")
	}
	if !file.JSON.ContainerID.Valid() {
		return fail(suite, "container file missing container_id")
	}
	if file.ContainerID != expectedContainerID {
		return fail(suite, fmt.Sprintf("container file container_id is %q, want %q", file.ContainerID, expectedContainerID))
	}
	if !file.JSON.CreatedAt.Valid() {
		return fail(suite, "container file missing created_at")
	}
	if !file.JSON.Object.Valid() {
		return fail(suite, "container file missing object")
	}
	if string(file.Object) != "container.file" {
		return fail(suite, fmt.Sprintf("container file object is %q, want container.file", file.Object))
	}
	if !file.JSON.Path.Valid() {
		return fail(suite, "container file missing path")
	}
	if !file.JSON.Source.Valid() {
		return fail(suite, "container file missing source")
	}
	return nil
}

func validateContainerFileGetObject(suite string, file *openai.ContainerFileGetResponse, expectedContainerID string) error {
	if file == nil {
		return fail(suite, "container file is nil")
	}
	if file.ID == "" {
		return fail(suite, "container file missing id")
	}
	if !file.JSON.Bytes.Valid() {
		return fail(suite, "container file missing bytes")
	}
	if !file.JSON.ContainerID.Valid() {
		return fail(suite, "container file missing container_id")
	}
	if file.ContainerID != expectedContainerID {
		return fail(suite, fmt.Sprintf("container file container_id is %q, want %q", file.ContainerID, expectedContainerID))
	}
	if !file.JSON.CreatedAt.Valid() {
		return fail(suite, "container file missing created_at")
	}
	if !file.JSON.Object.Valid() {
		return fail(suite, "container file missing object")
	}
	if string(file.Object) != "container.file" {
		return fail(suite, fmt.Sprintf("container file object is %q, want container.file", file.Object))
	}
	if !file.JSON.Path.Valid() {
		return fail(suite, "container file missing path")
	}
	if !file.JSON.Source.Valid() {
		return fail(suite, "container file missing source")
	}
	return nil
}

func validateContainerFileListPage(suite string, page *pagination.CursorPage[openai.ContainerFileListResponse], expectedContainerID string) error {
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
		item := &page.Data[i]
		if item.ID == "" {
			return fail(suite, "list item missing id")
		}
		if !item.JSON.Bytes.Valid() {
			return fail(suite, "list item missing bytes")
		}
		if !item.JSON.ContainerID.Valid() {
			return fail(suite, "list item missing container_id")
		}
		if item.ContainerID != expectedContainerID {
			return fail(suite, fmt.Sprintf("list item container_id is %q, want %q", item.ContainerID, expectedContainerID))
		}
		if !item.JSON.CreatedAt.Valid() {
			return fail(suite, "list item missing created_at")
		}
		if !item.JSON.Object.Valid() {
			return fail(suite, "list item missing object")
		}
		if string(item.Object) != "container.file" {
			return fail(suite, fmt.Sprintf("list item object is %q, want container.file", item.Object))
		}
		if !item.JSON.Path.Valid() {
			return fail(suite, "list item missing path")
		}
		if !item.JSON.Source.Valid() {
			return fail(suite, "list item missing source")
		}
	}
	return nil
}

func containerFileListContainsID(files []openai.ContainerFileListResponse, fileID string) bool {
	for _, file := range files {
		if file.ID == fileID {
			return true
		}
	}
	return false
}

func validateContainerFileContentResponse(suite string, resp *http.Response, want []byte) error {
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