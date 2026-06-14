package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// Moderations verifies POST /v1/moderations via client.Moderations.New.
type Moderations struct{}

func (Moderations) Name() string        { return "moderations" }
func (Moderations) Description() string { return "Moderations (POST /v1/moderations)" }

func (Moderations) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	resp, err := client.Moderations.New(ctx, openai.ModerationNewParams{
		Input: openai.ModerationNewParamsInputUnion{
			OfString: openai.String("compatibility test input"),
		},
	})
	if err != nil {
		return fmt.Errorf("moderation request failed: %w", err)
	}
	if resp == nil {
		return fail("moderations", "response is nil")
	}
	if resp.ID == "" {
		return fail("moderations", "response missing id")
	}
	if resp.Model == "" {
		return fail("moderations", "response missing model")
	}
	if len(resp.Results) == 0 {
		return fail("moderations", "response missing results")
	}

	result := resp.Results[0]
	if !result.JSON.Flagged.Valid() {
		return fail("moderations", "result missing flagged")
	}
	if !result.JSON.Categories.Valid() {
		return fail("moderations", "result missing categories")
	}
	if !result.JSON.CategoryScores.Valid() {
		return fail("moderations", "result missing category_scores")
	}
	if !result.JSON.CategoryAppliedInputTypes.Valid() {
		return fail("moderations", "result missing category_applied_input_types")
	}
	return nil
}