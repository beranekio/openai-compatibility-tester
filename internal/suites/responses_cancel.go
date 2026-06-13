package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesCancel verifies POST /v1/responses/{id}/cancel for background responses.
type ResponsesCancel struct{}

func (ResponsesCancel) Name() string { return "responses_cancel" }
func (ResponsesCancel) Description() string {
	return "Responses API cancel (POST /v1/responses with background, then POST /v1/responses/{id}/cancel)"
}

func (ResponsesCancel) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Background task placeholder"),
		},
		Background: openai.Bool(true),
		Store:      openai.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("create background response failed: %w", err)
	}
	if created == nil || created.ID == "" {
		return fail("responses_cancel", "background create missing response id")
	}
	status := string(created.Status)
	if status != "queued" && status != "in_progress" {
		return fail("responses_cancel", fmt.Sprintf("background create status is %q, want queued or in_progress", status))
	}

	cancelled, err := client.Responses.Cancel(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("responses cancel failed: %w", err)
	}
	if cancelled == nil {
		return fail("responses_cancel", "cancel response is nil")
	}
	if err := validateResponseEnvelope("responses_cancel", cancelled); err != nil {
		return err
	}
	if cancelled.ID != created.ID {
		return fail("responses_cancel", fmt.Sprintf("cancel id is %q, want %q", cancelled.ID, created.ID))
	}
	if string(cancelled.Status) != "cancelled" {
		return fail("responses_cancel", fmt.Sprintf("cancel status is %q, want cancelled", cancelled.Status))
	}
	return nil
}