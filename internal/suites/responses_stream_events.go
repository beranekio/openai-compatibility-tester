package suites

import (
	"fmt"

	"github.com/openai/openai-go/v3/responses"
)

// Responses stream event taxonomy for compatibility testing.
//
// Required for a passing text stream:
//   - At least one output delta: response.output_text.delta or response.refusal.delta
//   - A terminal event: response.completed, or response.incomplete (content_filter only)
//
// Optional lifecycle and metadata events (proxies may omit these):
//   - response.created, response.in_progress
//   - response.output_item.added, response.output_item.done
//   - response.content_part.added, response.content_part.done
//   - response.output_text.done, response.refusal.done
//
// Failure terminal events (always fail unless content_filter incomplete):
//   - response.failed, error
func isOptionalResponsesStreamEvent(eventType string) bool {
	switch eventType {
	case "response.created",
		"response.in_progress",
		"response.output_item.added",
		"response.output_item.done",
		"response.content_part.added",
		"response.content_part.done",
		"response.output_text.done",
		"response.refusal.done":
		return true
	default:
		return false
	}
}

func validateOptionalResponsesStreamEvent(suite string, event responses.ResponseStreamEventUnion) error {
	switch event.Type {
	case "response.created":
		created := event.AsResponseCreated()
		if !created.JSON.Response.Valid() {
			return fail(suite, "response.created missing response object")
		}
		if created.Response.ID == "" {
			return fail(suite, "response.created response missing id")
		}
	case "response.in_progress":
		inProgress := event.AsResponseInProgress()
		if !inProgress.JSON.Response.Valid() {
			return fail(suite, "response.in_progress missing response object")
		}
		if inProgress.Response.ID == "" {
			return fail(suite, "response.in_progress response missing id")
		}
	case "response.output_item.added":
		added := event.AsResponseOutputItemAdded()
		if !added.JSON.OutputIndex.Valid() {
			return fail(suite, "response.output_item.added missing output_index")
		}
		if !added.JSON.SequenceNumber.Valid() {
			return fail(suite, "response.output_item.added missing sequence_number")
		}
	case "response.output_item.done":
		done := event.AsResponseOutputItemDone()
		if !done.JSON.OutputIndex.Valid() {
			return fail(suite, "response.output_item.done missing output_index")
		}
		if !done.JSON.SequenceNumber.Valid() {
			return fail(suite, "response.output_item.done missing sequence_number")
		}
	case "response.content_part.added":
		added := event.AsResponseContentPartAdded()
		if added.ItemID == "" {
			return fail(suite, "response.content_part.added missing item_id")
		}
		if !added.JSON.OutputIndex.Valid() {
			return fail(suite, "response.content_part.added missing output_index")
		}
		if !added.JSON.ContentIndex.Valid() {
			return fail(suite, "response.content_part.added missing content_index")
		}
	case "response.content_part.done":
		done := event.AsResponseContentPartDone()
		if done.ItemID == "" {
			return fail(suite, "response.content_part.done missing item_id")
		}
		if !done.JSON.OutputIndex.Valid() {
			return fail(suite, "response.content_part.done missing output_index")
		}
		if !done.JSON.ContentIndex.Valid() {
			return fail(suite, "response.content_part.done missing content_index")
		}
	case "response.output_text.done":
		done := event.AsResponseOutputTextDone()
		if done.ItemID == "" {
			return fail(suite, "response.output_text.done missing item_id")
		}
	case "response.refusal.done":
		done := event.AsResponseRefusalDone()
		if done.ItemID == "" {
			return fail(suite, "response.refusal.done missing item_id")
		}
	default:
		return fail(suite, fmt.Sprintf("unexpected optional event %q", event.Type))
	}
	return nil
}