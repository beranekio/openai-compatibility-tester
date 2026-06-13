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

	var hasArgumentDelta bool
	var functionCallDone bool
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
		case "response.function_call_arguments.delta":
			delta := event.AsResponseFunctionCallArgumentsDelta()
			if err := validateResponseFunctionCallArgumentsDelta("responses_tools_stream", delta); err != nil {
				return err
			}
			if delta.Delta != "" {
				hasArgumentDelta = true
			}
		case "response.function_call_arguments.done":
			done := event.AsResponseFunctionCallArgumentsDone()
			if err := validateResponseFunctionCallArgumentsDone("responses_tools_stream", done); err != nil {
				return err
			}
			functionCallDone = true
		case "response.completed":
			completedEvent := event.AsResponseCompleted()
			if !completedEvent.JSON.Response.Valid() {
				return fail("responses_tools_stream", "response.completed missing response object")
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
	if contentFilterIncomplete {
		return nil
	}
	if !completed {
		return fail("responses_tools_stream", "stream missing response.completed event")
	}
	if !functionCallDone && !hasArgumentDelta {
		return fail("responses_tools_stream", "stream produced no function call argument events")
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
	return nil
}

func validateResponseFunctionCallArgumentsDone(suite string, done responses.ResponseFunctionCallArgumentsDoneEvent) error {
	if done.ItemID == "" {
		return fail(suite, "response.function_call_arguments.done missing item_id")
	}
	if done.Name == "" {
		return fail(suite, "response.function_call_arguments.done missing name")
	}
	if done.Arguments == "" {
		return fail(suite, "response.function_call_arguments.done missing arguments")
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
	return nil
}