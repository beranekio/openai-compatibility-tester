package suites

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

const (
	videoCreatePrompt      = "A short compatibility test clip of a red circle on a white background."
	videoContentProbeBytes = 512
)

// Videos verifies the Videos API lifecycle via client.Videos.*.
type Videos struct{}

func (Videos) Name() string { return "videos" }
func (Videos) Description() string {
	return "Videos API lifecycle (POST/GET/DELETE /v1/videos, GET /v1/videos/{id}/content)"
}

func (Videos) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	deleted := false
	var videoID string
	defer func() {
		if videoID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Videos.Delete(cleanupCtx, videoID)
		}
	}()

	submitted, err := client.Videos.New(ctx, openai.VideoNewParams{
		Model:   openai.VideoModel(cfg.VideoModel),
		Prompt:  videoCreatePrompt,
		Seconds: openai.VideoSeconds4,
		Size:    openai.VideoSize720x1280,
	})
	if err != nil {
		return fmt.Errorf("video create failed: %w", err)
	}
	if err := validateVideoObject("videos", submitted); err != nil {
		return err
	}
	videoID = submitted.ID

	created, err := client.Videos.PollStatus(ctx, videoID, 0)
	if err != nil {
		return fmt.Errorf("video poll failed: %w", err)
	}
	if err := validateVideoObject("videos", created); err != nil {
		return err
	}
	if created.Status != openai.VideoStatusCompleted {
		return fail("videos", fmt.Sprintf("create status is %q, want completed", created.Status))
	}

	got, err := client.Videos.Get(ctx, videoID)
	if err != nil {
		return fmt.Errorf("video get failed: %w", err)
	}
	if err := validateVideoObject("videos", got); err != nil {
		return err
	}
	if got.ID != videoID {
		return fail("videos", fmt.Sprintf("get id is %q, want %q", got.ID, videoID))
	}
	if got.Status != openai.VideoStatusCompleted {
		return fail("videos", fmt.Sprintf("get status is %q, want completed", got.Status))
	}

	listPage, err := client.Videos.List(ctx, openai.VideoListParams{
		Limit: openai.Int(10),
		Order: openai.VideoListParamsOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("video list failed: %w", err)
	}
	if err := validateVideoListPage("videos", listPage); err != nil {
		return err
	}
	if !videoListContainsID(listPage, videoID) {
		return fail("videos", "created video missing from list response")
	}

	contentResp, err := client.Videos.DownloadContent(ctx, videoID, openai.VideoDownloadContentParams{})
	if err != nil {
		return fmt.Errorf("video content download failed: %w", err)
	}
	if err := validateVideoContentResponse("videos", contentResp, 1); err != nil {
		return err
	}

	deletedResp, err := client.Videos.Delete(ctx, videoID)
	if err != nil {
		return fmt.Errorf("video delete failed: %w", err)
	}
	if err := validateVideoDeleteResponse("videos", deletedResp, videoID); err != nil {
		return err
	}

	_, getErr := client.Videos.Get(ctx, videoID)
	if getErr == nil {
		return fail("videos", "get after delete succeeded; video still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("videos", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}
	deleted = true
	return nil
}

func validateVideoObject(suite string, video *openai.Video) error {
	if video == nil {
		return fail(suite, "video is nil")
	}
	if video.ID == "" {
		return fail(suite, "video missing id")
	}
	if !video.JSON.CreatedAt.Valid() {
		return fail(suite, "video missing created_at")
	}
	if video.Status == openai.VideoStatusCompleted && !video.JSON.CompletedAt.Valid() {
		return fail(suite, "completed video missing completed_at")
	}
	if video.Status == openai.VideoStatusCompleted && !video.JSON.ExpiresAt.Valid() {
		return fail(suite, "completed video missing expires_at")
	}
	if video.Status == openai.VideoStatusFailed && !video.JSON.Error.Valid() {
		return fail(suite, "failed video missing error")
	}
	if !video.JSON.Model.Valid() {
		return fail(suite, "video missing model")
	}
	if !video.JSON.Object.Valid() {
		return fail(suite, "video missing object")
	}
	if string(video.Object) != "video" {
		return fail(suite, fmt.Sprintf("video object is %q, want video", video.Object))
	}
	if !video.JSON.Progress.Valid() {
		return fail(suite, "video missing progress")
	}
	if !video.JSON.Seconds.Valid() {
		return fail(suite, "video missing seconds")
	}
	if !video.JSON.Size.Valid() {
		return fail(suite, "video missing size")
	}
	if !video.JSON.Status.Valid() {
		return fail(suite, "video missing status")
	}
	if !isVideoStatusOK(video.Status) {
		return fail(suite, fmt.Sprintf("video status is %q, want queued, in_progress, completed, or failed", video.Status))
	}
	if video.Status == openai.VideoStatusCompleted && video.Progress < 100 {
		return fail(suite, fmt.Sprintf("completed video progress is %d, want >= 100", video.Progress))
	}
	return nil
}

func isVideoStatusOK(status openai.VideoStatus) bool {
	switch status {
	case openai.VideoStatusQueued, openai.VideoStatusInProgress, openai.VideoStatusCompleted, openai.VideoStatusFailed:
		return true
	default:
		return false
	}
}

func validateVideoListPage(suite string, page *pagination.ConversationCursorPage[openai.Video]) error {
	if page == nil {
		return fail(suite, "list page is nil")
	}
	for i := range page.Data {
		if err := validateVideoListItem(suite, &page.Data[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateVideoListItem(suite string, video *openai.Video) error {
	if video == nil {
		return fail(suite, "list video is nil")
	}
	if video.ID == "" {
		return fail(suite, "list video missing id")
	}
	if video.JSON.Object.Valid() && string(video.Object) != "video" {
		return fail(suite, fmt.Sprintf("list video object is %q, want video", video.Object))
	}
	if video.JSON.Model.Valid() && video.Model == "" {
		return fail(suite, "list video model is empty")
	}
	if video.JSON.Status.Valid() && !isVideoStatusOK(video.Status) {
		return fail(suite, fmt.Sprintf("list video status is %q, want queued, in_progress, completed, or failed", video.Status))
	}
	return nil
}

func videoListContainsID(page *pagination.ConversationCursorPage[openai.Video], videoID string) bool {
	for _, video := range page.Data {
		if video.ID == videoID {
			return true
		}
	}
	return false
}

func validateVideoDeleteResponse(suite string, deleted *openai.VideoDeleteResponse, wantID string) error {
	if deleted == nil {
		return fail(suite, "delete response is nil")
	}
	if deleted.ID != wantID {
		return fail(suite, fmt.Sprintf("delete id is %q, want %q", deleted.ID, wantID))
	}
	if !deleted.JSON.Deleted.Valid() {
		return fail(suite, "delete response missing deleted")
	}
	if !deleted.Deleted {
		return fail(suite, "delete response deleted is false")
	}
	if !deleted.JSON.Object.Valid() {
		return fail(suite, "delete response missing object")
	}
	if string(deleted.Object) != "video.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want video.deleted", deleted.Object))
	}
	return nil
}

func validateVideoContentResponse(suite string, resp *http.Response, minBytes int) error {
	if resp == nil {
		return fail(suite, "content response is nil")
	}
	if resp.Body == nil {
		return fail(suite, "content response body is nil")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fail(suite, fmt.Sprintf("content status is %d, want 200", resp.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, videoContentProbeBytes))
	if err != nil {
		return fmt.Errorf("%s: read content body: %w", suite, err)
	}
	if len(body) < minBytes {
		return fail(suite, fmt.Sprintf("content has %d bytes, want at least %d", len(body), minBytes))
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		return fail(suite, "content response missing Content-Type")
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fail(suite, fmt.Sprintf("Content-Type %q is invalid: %v", contentType, err))
	}
	if !strings.HasPrefix(mediaType, "video/") && mediaType != "application/octet-stream" {
		return fail(suite, fmt.Sprintf("Content-Type is %q, want video/* or application/octet-stream", mediaType))
	}
	return nil
}