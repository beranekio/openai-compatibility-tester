package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ModerationsImage verifies POST /v1/moderations with image input.
type ModerationsImage struct{}

func (ModerationsImage) Name() string { return "moderations_image" }
func (ModerationsImage) Description() string {
	return "Moderations with image input (POST /v1/moderations)"
}

func (ModerationsImage) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	resp, err := client.Moderations.New(ctx, openai.ModerationNewParams{
		Input: openai.ModerationNewParamsInputUnion{
			OfModerationMultiModalArray: []openai.ModerationMultiModalInputUnionParam{{
				OfImageURL: &openai.ModerationImageURLInputParam{
					ImageURL: openai.ModerationImageURLInputImageURLParam{
						URL: smallPNGDataURL,
					},
				},
			}},
		},
	})
	if err != nil {
		return fmt.Errorf("moderation image request failed: %w", err)
	}
	if resp == nil {
		return fail("moderations_image", "response is nil")
	}
	if resp.ID == "" {
		return fail("moderations_image", "response missing id")
	}
	if resp.Model == "" {
		return fail("moderations_image", "response missing model")
	}
	if len(resp.Results) != 1 {
		return fail("moderations_image", fmt.Sprintf("response has %d results, want 1", len(resp.Results)))
	}
	return validateModerationResult("moderations_image", resp.Results[0])
}
