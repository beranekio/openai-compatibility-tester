package suites

import (
	"context"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/testutil"

	"github.com/openai/openai-go/v3"
)

// UploadsCancel verifies POST /v1/uploads/{id}/cancel via client.Uploads.Cancel.
// Cancel only applies to a pending (in-progress) upload, so the suite creates a
// fresh upload and cancels it before any parts are added or it is completed.
type UploadsCancel struct{}

func (UploadsCancel) Name() string { return "uploads_cancel" }
func (UploadsCancel) Description() string {
	return "Uploads API cancel (POST /v1/uploads, then POST /v1/uploads/{id}/cancel)"
}

func (UploadsCancel) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	content := testutil.SmallTextFileBytes()
	cancelled := false
	var uploadID string
	defer func() {
		if cancelled {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if uploadID != "" {
			_, _ = client.Uploads.Cancel(cleanupCtx, uploadID)
		}
	}()

	created, err := client.Uploads.New(ctx, openai.UploadNewParams{
		Bytes:    int64(len(content)),
		Filename: "cancel.txt",
		MimeType: "text/plain",
		Purpose:  openai.FilePurposeUserData,
	})
	if err != nil {
		return fmt.Errorf("upload create failed: %w", err)
	}
	if err := validateUploadObject("uploads_cancel", created); err != nil {
		return err
	}
	uploadID = created.ID
	if created.Status != openai.UploadStatusPending {
		return fail("uploads_cancel", fmt.Sprintf("create status is %q, want pending", created.Status))
	}

	result, err := client.Uploads.Cancel(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("upload cancel failed: %w", err)
	}
	if err := validateUploadObject("uploads_cancel", result); err != nil {
		return err
	}
	if result.ID != created.ID {
		return fail("uploads_cancel", fmt.Sprintf("cancel id is %q, want %q", result.ID, created.ID))
	}
	if result.Status != openai.UploadStatusCancelled {
		return fail("uploads_cancel", fmt.Sprintf("cancel status is %q, want cancelled", result.Status))
	}
	cancelled = true
	return nil
}
