package suites

import (
	"context"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// VideosVariants verifies the video edit/extend/remix endpoints via
// client.Videos.Edit, client.Videos.Extend, and client.Videos.Remix. Each
// operates on a completed reference video and produces a new video job.
type VideosVariants struct{}

func (VideosVariants) Name() string { return "videos_variants" }
func (VideosVariants) Description() string {
	return "Videos API edit/extend/remix (POST /v1/videos/edits, /v1/videos/extensions, /v1/videos/{id}/remix)"
}

func (VideosVariants) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var cleanupIDs []string
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, id := range cleanupIDs {
			_, _ = client.Videos.Delete(cleanupCtx, id)
		}
	}()

	// Create and complete a reference video for the variant endpoints.
	reference, err := client.Videos.New(ctx, openai.VideoNewParams{
		Model:   openai.VideoModel(cfg.VideoModel),
		Prompt:  videoCreatePrompt,
		Seconds: openai.VideoSeconds4,
		Size:    openai.VideoSize720x1280,
	})
	if err != nil {
		return fmt.Errorf("reference video create failed: %w", err)
	}
	if reference != nil {
		cleanupIDs = append(cleanupIDs, reference.ID)
	}
	if err := validateVideoObject("videos_variants", reference); err != nil {
		return err
	}
	completed, err := client.Videos.PollStatus(ctx, reference.ID, 0)
	if err != nil {
		return fmt.Errorf("reference video poll failed: %w", err)
	}
	if err := validateVideoObject("videos_variants", completed); err != nil {
		return err
	}
	if completed.Status != openai.VideoStatusCompleted {
		return fail("videos_variants", fmt.Sprintf("reference video status is %q, want completed", completed.Status))
	}
	// Use the originally submitted id, not the poll response's id, so a
	// misbehaving provider can't redirect variant calls to a different video.
	refID := reference.ID

	const variantPrompt = "A compatibility-test variant of the reference clip."

	edited, err := client.Videos.Edit(ctx, openai.VideoEditParams{
		Prompt: variantPrompt,
		Video: openai.VideoEditParamsVideoUnion{
			OfVideoEditsVideoVideoReferenceInputParam: &openai.VideoEditParamsVideoVideoReferenceInputParam{ID: refID},
		},
	})
	if err != nil {
		return fmt.Errorf("video edit failed: %w", err)
	}
	if err := validateVideoObject("videos_variants", edited); err != nil {
		return err
	}
	if err := validateVideoVariant("videos_variants", edited, refID); err != nil {
		return err
	}
	cleanupIDs = append(cleanupIDs, edited.ID)

	extended, err := client.Videos.Extend(ctx, openai.VideoExtendParams{
		Prompt:  variantPrompt,
		Seconds: openai.VideoSeconds4,
		Video: openai.VideoExtendParamsVideoUnion{
			OfVideoExtendsVideoVideoReferenceInputParam: &openai.VideoExtendParamsVideoVideoReferenceInputParam{ID: refID},
		},
	})
	if err != nil {
		return fmt.Errorf("video extend failed: %w", err)
	}
	if err := validateVideoObject("videos_variants", extended); err != nil {
		return err
	}
	if err := validateVideoVariant("videos_variants", extended, refID); err != nil {
		return err
	}
	cleanupIDs = append(cleanupIDs, extended.ID)

	remixed, err := client.Videos.Remix(ctx, refID, openai.VideoRemixParams{
		Prompt: variantPrompt,
	})
	if err != nil {
		return fmt.Errorf("video remix failed: %w", err)
	}
	if err := validateVideoObject("videos_variants", remixed); err != nil {
		return err
	}
	if err := validateVideoVariant("videos_variants", remixed, refID); err != nil {
		return err
	}
	cleanupIDs = append(cleanupIDs, remixed.ID)
	return nil
}

// validateVideoVariant validates a video returned by an edit/extend/remix call:
// it must carry a non-empty id distinct from the reference (a new job, not an
// echo of the source) and must not be in the failed status (the variant
// generation must have been accepted, not immediately rejected).
func validateVideoVariant(suite string, video *openai.Video, refID string) error {
	if video.ID == "" {
		return fail(suite, "variant response missing id")
	}
	if video.ID == refID {
		return fail(suite, fmt.Sprintf("variant response id %q matches reference; expected a new job", video.ID))
	}
	if video.Status == openai.VideoStatusFailed {
		return fail(suite, fmt.Sprintf("variant response status is %q, want queued/in_progress/completed", video.Status))
	}
	return nil
}
