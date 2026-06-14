package suites

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// Uploads verifies the Uploads API lifecycle via client.Uploads.*.
type Uploads struct{}

func (Uploads) Name() string { return "uploads" }
func (Uploads) Description() string {
	return "Uploads API lifecycle (POST /v1/uploads, parts, complete)"
}

func (Uploads) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	content := smallTextFileBytes()
	deleted := false
	var fileID string
	defer func() {
		if fileID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Files.Delete(cleanupCtx, fileID)
		}
	}()

	created, err := client.Uploads.New(ctx, openai.UploadNewParams{
		Bytes:    int64(len(content)),
		Filename: "test.txt",
		MimeType: "text/plain",
		Purpose:  openai.FilePurposeUserData,
	})
	if err != nil {
		return fmt.Errorf("upload create failed: %w", err)
	}
	if err := validateUploadObject("uploads", created); err != nil {
		return err
	}
	if created.Status != openai.UploadStatusPending {
		return fail("uploads", fmt.Sprintf("create status is %q, want pending", created.Status))
	}
	if created.Bytes != int64(len(content)) {
		return fail("uploads", fmt.Sprintf("create bytes is %d, want %d", created.Bytes, len(content)))
	}
	if created.Filename != "test.txt" {
		return fail("uploads", fmt.Sprintf("create filename is %q, want test.txt", created.Filename))
	}
	if created.Purpose != string(openai.FilePurposeUserData) {
		return fail("uploads", fmt.Sprintf("create purpose is %q, want user_data", created.Purpose))
	}

	part, err := client.Uploads.Parts.New(ctx, created.ID, openai.UploadPartNewParams{
		Data: bytes.NewReader(content),
	})
	if err != nil {
		return fmt.Errorf("upload part failed: %w", err)
	}
	if err := validateUploadPartObject("uploads", part); err != nil {
		return err
	}
	if part.UploadID != created.ID {
		return fail("uploads", fmt.Sprintf("part upload_id is %q, want %q", part.UploadID, created.ID))
	}

	completed, err := client.Uploads.Complete(ctx, created.ID, openai.UploadCompleteParams{
		PartIDs: []string{part.ID},
	})
	if err != nil {
		return fmt.Errorf("upload complete failed: %w", err)
	}
	if err := validateUploadObject("uploads", completed); err != nil {
		return err
	}
	if completed.Status != openai.UploadStatusCompleted {
		return fail("uploads", fmt.Sprintf("complete status is %q, want completed", completed.Status))
	}
	if !completed.JSON.File.Valid() {
		return fail("uploads", "complete response missing file")
	}
	if err := validateFileObject("uploads", &completed.File); err != nil {
		return err
	}
	if completed.File.Filename != "test.txt" {
		return fail("uploads", fmt.Sprintf("file filename is %q, want test.txt", completed.File.Filename))
	}
	if string(completed.File.Purpose) != string(openai.FilePurposeUserData) {
		return fail("uploads", fmt.Sprintf("file purpose is %q, want user_data", completed.File.Purpose))
	}
	fileID = completed.File.ID

	deletedResp, err := client.Files.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file delete failed: %w", err)
	}
	if deletedResp == nil || !deletedResp.Deleted {
		return fail("uploads", "file delete response invalid")
	}
	deleted = true
	return nil
}

func validateUploadObject(suite string, upload *openai.Upload) error {
	if upload == nil {
		return fail(suite, "upload is nil")
	}
	if upload.ID == "" {
		return fail(suite, "upload missing id")
	}
	if !upload.JSON.Bytes.Valid() {
		return fail(suite, "upload missing bytes")
	}
	if !upload.JSON.CreatedAt.Valid() {
		return fail(suite, "upload missing created_at")
	}
	if !upload.JSON.ExpiresAt.Valid() {
		return fail(suite, "upload missing expires_at")
	}
	if upload.Filename == "" {
		return fail(suite, "upload missing filename")
	}
	if string(upload.Object) != "upload" {
		return fail(suite, fmt.Sprintf("upload object is %q, want upload", upload.Object))
	}
	if upload.Purpose == "" {
		return fail(suite, "upload missing purpose")
	}
	if !upload.JSON.Status.Valid() {
		return fail(suite, "upload missing status")
	}
	return nil
}

func validateUploadPartObject(suite string, part *openai.UploadPart) error {
	if part == nil {
		return fail(suite, "upload part is nil")
	}
	if part.ID == "" {
		return fail(suite, "upload part missing id")
	}
	if !part.JSON.CreatedAt.Valid() {
		return fail(suite, "upload part missing created_at")
	}
	if string(part.Object) != "upload.part" {
		return fail(suite, fmt.Sprintf("upload part object is %q, want upload.part", part.Object))
	}
	if part.UploadID == "" {
		return fail(suite, "upload part missing upload_id")
	}
	return nil
}