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

// ResponsesToolsStream verifies streaming tool calling on the Responses API.
type ResponsesToolsStream struct{}

func (ResponsesToolsStream) Name() string { return "responses_tools_stream" }
func (ResponsesToolsStream) Description() string {
	return "Streaming Responses API with tools (POST /v1/responses, stream=true, tool_choice required)"
}

func (ResponsesToolsStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("What is the weather in San Francisco?"),
		},
		Tools:      weatherResponseTools(),
		ToolChoice: requiredResponseToolChoice(),
		Store:      openai.Bool(false),
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("responses tools stream failed: %w", err)
	}
	if err := validateEventStreamContentType("responses_tools_stream", httpResp); err != nil {
		return err
	}

	var hasRefusal bool
	var functionCallDone bool
	var functionCallOutputItemDone bool
	var completed bool
	var contentFilterIncomplete bool
	var terminalFailure bool
	var terminalReached bool
	for stream.Next() {
		if terminalReached {
			return fail("responses_tools_stream", fmt.Sprintf("stream event %q after terminal event", stream.Current().Type))
		}

		event := stream.Current()
		switch event.Type {
		case "response.refusal.delta":
			delta := event.AsResponseRefusalDelta()
			if err := validateResponseRefusalDelta("responses_tools_stream", delta); err != nil {
				return err
			}
			if delta.Delta != "" {
				hasRefusal = true
			}
		case "response.function_call_arguments.delta":
			delta := event.AsResponseFunctionCallArgumentsDelta()
			if err := validateResponseFunctionCallArgumentsDelta("responses_tools_stream", delta); err != nil {
				return err
			}
		case "response.function_call_arguments.done":
			done := event.AsResponseFunctionCallArgumentsDone()
			if err := validateResponseFunctionCallArgumentsDone("responses_tools_stream", done); err != nil {
				return err
			}
			functionCallDone = true
		case "response.output_item.done":
			done := event.AsResponseOutputItemDone()
			if !done.JSON.Item.Valid() {
				return fail("responses_tools_stream", "response.output_item.done missing item")
			}
			if !done.JSON.OutputIndex.Valid() {
				return fail("responses_tools_stream", "response.output_item.done missing output_index")
			}
			if !done.JSON.SequenceNumber.Valid() {
				return fail("responses_tools_stream", "response.output_item.done missing sequence_number")
			}
			if done.Item.Type != "function_call" {
				return fail("responses_tools_stream", fmt.Sprintf("response.output_item.done item type is %q, want function_call", done.Item.Type))
			}
			if err := validateResponseFunctionToolCall("responses_tools_stream", done.Item.AsFunctionCall()); err != nil {
				return err
			}
			functionCallOutputItemDone = true
		case "response.completed":
			completedEvent := event.AsResponseCompleted()
			if !completedEvent.JSON.Response.Valid() {
				return fail("responses_tools_stream", "response.completed missing response object")
			}
			if !completedEvent.JSON.SequenceNumber.Valid() {
				return fail("responses_tools_stream", "response.completed missing sequence_number")
			}
			if completedEvent.Response.ID == "" {
				return fail("responses_tools_stream", "response.completed response missing id")
			}
			if string(completedEvent.Response.Status) != "completed" {
				return fail("responses_tools_stream", fmt.Sprintf("response.completed status is %q, want completed", completedEvent.Response.Status))
			}
			completed = true
			terminalReached = true
		case "response.incomplete":
			incompleteEvent := event.AsResponseIncomplete()
			if !incompleteEvent.JSON.Response.Valid() {
				return fail("responses_tools_stream", "response.incomplete missing response object")
			}
			if !incompleteEvent.JSON.SequenceNumber.Valid() {
				return fail("responses_tools_stream", "response.incomplete missing sequence_number")
			}
			if incompleteEvent.Response.ID == "" {
				return fail("responses_tools_stream", "response.incomplete response missing id")
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
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("responses tools stream failed: %w", err)
	}
	if terminalFailure {
		return fail("responses_tools_stream", "stream ended with a failure event")
	}
	if !completed && !contentFilterIncomplete {
		return fail("responses_tools_stream", "stream missing terminal event")
	}
	if contentFilterIncomplete {
		return nil
	}
	if hasRefusal {
		return nil
	}
	if !functionCallDone {
		return fail("responses_tools_stream", "stream missing response.function_call_arguments.done event")
	}
	if !functionCallOutputItemDone {
		return fail("responses_tools_stream", "stream missing response.output_item.done event")
	}
	return nil
}

func validateResponseFunctionCallArgumentsDelta(suite string, delta responses.ResponseFunctionCallArgumentsDeltaEvent) error {
	if delta.ItemID == "" {
		return fail(suite, "response.function_call_arguments.delta missing item_id")
	}
	if !delta.JSON.OutputIndex.Valid() {
		return fail(suite, "response.function_call_arguments.delta missing output_index")
	}
	if !delta.JSON.SequenceNumber.Valid() {
		return fail(suite, "response.function_call_arguments.delta missing sequence_number")
	}
	if !delta.JSON.Delta.Valid() {
		return fail(suite, "response.function_call_arguments.delta missing delta")
	}
	return nil
}

func validateResponseFunctionCallArgumentsDone(suite string, done responses.ResponseFunctionCallArgumentsDoneEvent) error {
	if done.ItemID == "" {
		return fail(suite, "response.function_call_arguments.done missing item_id")
	}
	if done.Name == "" {
		return fail(suite, "response.function_call_arguments.done missing name")
	}
	if done.Name != weatherToolName {
		return fail(suite, fmt.Sprintf("function name is %q, want %s", done.Name, weatherToolName))
	}
	if !done.JSON.OutputIndex.Valid() {
		return fail(suite, "response.function_call_arguments.done missing output_index")
	}
	if !done.JSON.SequenceNumber.Valid() {
		return fail(suite, "response.function_call_arguments.done missing sequence_number")
	}
	return validateWeatherToolArguments(suite, done.Arguments)
}
