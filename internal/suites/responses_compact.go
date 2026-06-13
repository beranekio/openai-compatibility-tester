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
			OfResponseInputItemArray: []responses.ResponseInputItemUnionParam{
				responses.ResponseInputItemParamOfMessage("What is 2+2?", responses.EasyInputMessageRoleUser),
				responses.ResponseInputItemParamOfOutputMessage(
					[]responses.ResponseOutputMessageContentUnionParam{{
						OfOutputText: &responses.ResponseOutputTextParam{
							Text:        "4",
							Annotations: []responses.ResponseOutputTextAnnotationUnionParam{},
						},
					}},
					"msg-assistant-1",
					responses.ResponseOutputMessageStatusCompleted,
				),
				responses.ResponseInputItemParamOfMessage("Summarize this conversation.", responses.EasyInputMessageRoleUser),
			},
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
	if !resp.JSON.Usage.Valid() {
		return fail("responses_compact", "response missing usage")
	}
	if len(resp.Output) < 2 {
		return fail("responses_compact", "output must include user messages and a compaction item")
	}
	last := resp.Output[len(resp.Output)-1]
	if last.Type != "compaction" {
		return fail("responses_compact", fmt.Sprintf("last output item type is %q, want compaction", last.Type))
	}
	compaction := last.AsCompaction()
	if compaction.ID == "" {
		return fail("responses_compact", "compaction item missing id")
	}
	if compaction.EncryptedContent == "" {
		return fail("responses_compact", "compaction item missing encrypted_content")
	}
	for i, item := range resp.Output[:len(resp.Output)-1] {
		if item.Type != "message" {
			return fail("responses_compact", fmt.Sprintf("output item %d type is %q, want message", i, item.Type))
		}
		if item.Role != "user" {
			return fail("responses_compact", fmt.Sprintf("output item %d role is %q, want user", i, item.Role))
		}
	}
	return nil
}