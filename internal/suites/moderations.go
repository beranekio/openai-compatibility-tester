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
	if len(resp.Results) != 1 {
		return fail("moderations", fmt.Sprintf("response has %d results, want 1", len(resp.Results)))
	}
	return validateModerationResult("moderations", resp.Results[0])
}

func validateModerationResult(suite string, result openai.Moderation) error {
	if !result.JSON.Flagged.Valid() {
		return fail(suite, "result missing flagged")
	}
	if !result.JSON.Categories.Valid() {
		return fail(suite, "result missing categories")
	}
	if !result.JSON.CategoryScores.Valid() {
		return fail(suite, "result missing category_scores")
	}

	for _, check := range coreModerationCategoryChecks(result) {
		if !check.categoryValid {
			return fail(suite, fmt.Sprintf("categories missing %s", check.name))
		}
		if err := validateModerationScore(suite, check.name, check.score, check.scoreValid); err != nil {
			return err
		}
	}
	return nil
}

type moderationCategoryCheck struct {
	name          string
	categoryValid bool
	score         float64
	scoreValid    bool
}

func coreModerationCategoryChecks(result openai.Moderation) []moderationCategoryCheck {
	return []moderationCategoryCheck{
		{"harassment", result.Categories.JSON.Harassment.Valid(), result.CategoryScores.Harassment, result.CategoryScores.JSON.Harassment.Valid()},
		{"hate", result.Categories.JSON.Hate.Valid(), result.CategoryScores.Hate, result.CategoryScores.JSON.Hate.Valid()},
		{"sexual", result.Categories.JSON.Sexual.Valid(), result.CategoryScores.Sexual, result.CategoryScores.JSON.Sexual.Valid()},
		{"violence", result.Categories.JSON.Violence.Valid(), result.CategoryScores.Violence, result.CategoryScores.JSON.Violence.Valid()},
		{"self-harm", result.Categories.JSON.SelfHarm.Valid(), result.CategoryScores.SelfHarm, result.CategoryScores.JSON.SelfHarm.Valid()},
	}
}

func validateModerationScore(suite, name string, score float64, valid bool) error {
	if !valid {
		return fail(suite, fmt.Sprintf("category_scores missing %s", name))
	}
	if score < 0 || score > 1 {
		return fail(suite, fmt.Sprintf("category_scores %s is %v, want value between 0 and 1", name, score))
	}
	return nil
}
