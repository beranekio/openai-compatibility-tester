package suites

import "testing"

func TestIsOptionalResponsesStreamEvent(t *testing.T) {
	optional := []string{
		"response.created",
		"response.in_progress",
		"response.output_item.added",
		"response.output_item.done",
		"response.content_part.added",
		"response.content_part.done",
		"response.output_text.done",
		"response.refusal.done",
	}
	for _, eventType := range optional {
		if !isOptionalResponsesStreamEvent(eventType) {
			t.Fatalf("isOptionalResponsesStreamEvent(%q) = false, want true", eventType)
		}
	}

	required := []string{
		"response.output_text.delta",
		"response.completed",
		"response.function_call_arguments.delta",
	}
	for _, eventType := range required {
		if isOptionalResponsesStreamEvent(eventType) {
			t.Fatalf("isOptionalResponsesStreamEvent(%q) = true, want false", eventType)
		}
	}
}