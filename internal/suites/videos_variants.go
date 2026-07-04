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
	cleanupIDs = append(cleanupIDs, reference.ID)
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
	refID := completed.ID

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
	if edited.ID == "" {
		return fail("videos_variants", "edit response missing id")
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
	if extended.ID == "" {
		return fail("videos_variants", "extend response missing id")
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
	if remixed.ID == "" {
		return fail("videos_variants", "remix response missing id")
	}
	cleanupIDs = append(cleanupIDs, remixed.ID)
	return nil
}
