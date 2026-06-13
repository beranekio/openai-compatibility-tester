package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesStream verifies streaming responses API.
type ResponsesStream struct{}

func (ResponsesStream) Name() string { return "responses_stream" }
func (ResponsesStream) Description() string {
	return "Streaming responses API (POST /v1/responses, stream=true)"
}

func (ResponsesStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Count from one to three."),
		},
		Store: openai.Bool(false),
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("responses stream failed: %w", err)
	}
	if err := validateEventStreamContentType("responses_stream", httpResp); err != nil {
		return err
	}

	var hasOutput bool
	var completed bool
	var contentFilterIncomplete bool
	var terminalFailure bool
	var terminalReached bool
	for stream.Next() {
		if terminalReached {
			return fail("responses_stream", fmt.Sprintf("stream event %q after terminal event", stream.Current().Type))
		}

		event := stream.Current()
		switch event.Type {
		case "response.created", "response.in_progress",
			"response.output_item.added", "response.output_item.done",
			"response.content_part.added", "response.content_part.done",
			"response.output_text.done", "response.refusal.done":
			if err := validateOptionalResponsesStreamEvent("responses_stream", event); err != nil {
				return err
			}
		case "response.output_text.delta":
			delta := event.AsResponseOutputTextDelta()
			if err := validateResponseTextDelta("responses_stream", delta); err != nil {
				return err
			}
			if delta.Delta != "" {
				hasOutput = true
			}
		case "response.refusal.delta":
			delta := event.AsResponseRefusalDelta()
			if err := validateResponseRefusalDelta("responses_stream", delta); err != nil {
				return err
			}
			if delta.Delta != "" {
				hasOutput = true
			}
		case "response.completed":
			completedEvent := event.AsResponseCompleted()
			if !completedEvent.JSON.Response.Valid() {
				return fail("responses_stream", "response.completed missing response object")
			}
			if completedEvent.Response.ID == "" {
				return fail("responses_stream", "response.completed response missing id")
			}
			if string(completedEvent.Response.Status) != "completed" {
				return fail("responses_stream", fmt.Sprintf("response.completed status is %q, want completed", completedEvent.Response.Status))
			}
			completed = true
			terminalReached = true
		case "response.incomplete":
			incompleteEvent := event.AsResponseIncomplete()
			if !incompleteEvent.JSON.Response.Valid() {
				return fail("responses_stream", "response.incomplete missing response object")
			}
			if isContentFilterIncompleteResponse(&incompleteEvent.Response) {
				contentFilterIncomplete = true
			} else {
				terminalFailure = true
			}
			terminalReached = true
		case "response.failed", "error":
			terminalFailure = true
			terminalReached = true
		default:
			return fail("responses_stream", fmt.Sprintf("unsupported stream event %q", event.Type))
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("responses stream failed: %w", err)
	}
	if terminalFailure {
		return fail("responses_stream", "stream ended with a failure event")
	}
	if !completed && !contentFilterIncomplete {
		return fail("responses_stream", "stream missing response.completed event")
	}
	if !hasOutput && !contentFilterIncomplete {
		return fail("responses_stream", "stream produced no output text or refusal")
	}
	return nil
}

func validateResponseTextDelta(suite string, delta responses.ResponseTextDeltaEvent) error {
	if !delta.JSON.ContentIndex.Valid() {
		return fail(suite, "response.output_text.delta missing content_index")
	}
	if delta.ItemID == "" {
		return fail(suite, "response.output_text.delta missing item_id")
	}
	if !delta.JSON.OutputIndex.Valid() {
		return fail(suite, "response.output_text.delta missing output_index")
	}
	if !delta.JSON.SequenceNumber.Valid() {
		return fail(suite, "response.output_text.delta missing sequence_number")
	}
	return nil
}

func validateResponseRefusalDelta(suite string, delta responses.ResponseRefusalDeltaEvent) error {
	if !delta.JSON.ContentIndex.Valid() {
		return fail(suite, "response.refusal.delta missing content_index")
	}
	if delta.ItemID == "" {
		return fail(suite, "response.refusal.delta missing item_id")
	}
	if !delta.JSON.OutputIndex.Valid() {
		return fail(suite, "response.refusal.delta missing output_index")
	}
	if !delta.JSON.SequenceNumber.Valid() {
		return fail(suite, "response.refusal.delta missing sequence_number")
	}
	return nil
}