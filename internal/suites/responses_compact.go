package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

var compactUserTurns = []string{
	"What is 2+2?",
	"Summarize this conversation.",
}

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
				responses.ResponseInputItemParamOfMessage(compactUserTurns[0], responses.EasyInputMessageRoleUser),
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
				responses.ResponseInputItemParamOfMessage(compactUserTurns[1], responses.EasyInputMessageRoleUser),
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
	if len(resp.Output) != len(compactUserTurns)+1 {
		return fail("responses_compact", fmt.Sprintf("output has %d items, want %d user messages and 1 compaction item", len(resp.Output), len(compactUserTurns)))
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
		text, ok := compactOutputUserText(item)
		if !ok {
			return fail("responses_compact", fmt.Sprintf("output item %d missing input_text content", i))
		}
		if text != compactUserTurns[i] {
			return fail("responses_compact", fmt.Sprintf("output item %d text is %q, want %q", i, text, compactUserTurns[i]))
		}
	}
	return nil
}

func compactOutputUserText(item responses.ResponseOutputItemUnion) (string, bool) {
	for _, content := range item.Content {
		if content.Text == "" {
			continue
		}
		if content.Type == "input_text" || content.Type == "output_text" {
			return content.Text, true
		}
	}
	return "", false
}
