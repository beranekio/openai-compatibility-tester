package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesCompact verifies POST /v1/responses/compact.
type ResponsesCompact struct{}

func (ResponsesCompact) Name() string { return "responses_compact" }
func (ResponsesCompact) Description() string {
	return "Responses API compact (POST /v1/responses/compact)"
}

func (ResponsesCompact) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.Compact(ctx, responses.ResponseCompactParams{
		Model: responses.ResponseCompactParamsModel(cfg.ResponsesModel),
		Input: responses.ResponseCompactParamsInputUnion{
			OfString: openai.String("Summarize this short conversation."),
		},
	})
	if err != nil {
		return fmt.Errorf("responses compact request failed: %w", err)
	}
	if resp == nil {
		return fail("responses_compact", "response is nil")
	}
	if resp.ID == "" {
		return fail("responses_compact", "response missing id")
	}
	if string(resp.Object) != "response.compaction" {
		return fail("responses_compact", fmt.Sprintf("object is %q, want response.compaction", resp.Object))
	}
	if !resp.JSON.CreatedAt.Valid() {
		return fail("responses_compact", "response missing created_at")
	}
	if len(resp.Output) == 0 {
		return fail("responses_compact", "response missing output")
	}
	hasCompaction := false
	for _, item := range resp.Output {
		if item.Type == "compaction" {
			compaction := item.AsCompaction()
			if compaction.ID == "" {
				return fail("responses_compact", "compaction item missing id")
			}
			if compaction.EncryptedContent == "" {
				return fail("responses_compact", "compaction item missing encrypted_content")
			}
			hasCompaction = true
		}
	}
	if !hasCompaction {
		return fail("responses_compact", "response output missing compaction item")
	}
	return nil
}