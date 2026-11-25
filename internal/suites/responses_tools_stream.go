package suites

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

	calls := make(responseStreamCalls)
	var completed *responses.Response
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
		case "response.output_item.added", "response.output_item.done":
			if err := validateOptionalResponsesStreamEvent("responses_tools_stream", event); err != nil {
				return err
			}
			if event.Item.Type != "function_call" {
				continue
			}
			call, err := calls.at(event.OutputIndex, event.Item.ID)
			if err != nil {
				return err
			}
			item := event.Item.AsFunctionCall()
			if event.Type == "response.output_item.added" {
				if call.added != nil || call.hasDelta || call.arguments != nil || call.done != nil {
					return fail("responses_tools_stream", "late or duplicate function_call item added")
				}
				call.added = &item
				call.fragments.WriteString(item.Arguments)
			} else {
				if call.done != nil {
					return fail("responses_tools_stream", "duplicate function_call item done")
				}
				call.done = &item
			}
		case "response.function_call_arguments.delta":
			delta := event.AsResponseFunctionCallArgumentsDelta()
			if err := validateResponseFunctionCallArgumentsDelta("responses_tools_stream", delta); err != nil {
				return err
			}
			call, err := calls.at(delta.OutputIndex, delta.ItemID)
			if err != nil {
				return err
			}
			if call.arguments != nil || call.done != nil {
				return fail("responses_tools_stream", "function_call arguments delta after completion")
			}
			call.hasDelta = true
			call.fragments.WriteString(delta.Delta)
		case "response.function_call_arguments.done":
			done := event.AsResponseFunctionCallArgumentsDone()
			if err := validateResponseFunctionCallArgumentsDone("responses_tools_stream", done); err != nil {
				return err
			}
			call, err := calls.at(done.OutputIndex, done.ItemID)
			if err != nil {
				return err
			}
			if call.arguments != nil || call.done != nil {
				return fail("responses_tools_stream", "late or duplicate function_call arguments done")
			}
			call.arguments = &done.Arguments
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
			completed = &completedEvent.Response
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
	if completed == nil && !contentFilterIncomplete {
		return fail("responses_tools_stream", "stream missing terminal event")
	}
	if contentFilterIncomplete {
		return nil
	}
	return calls.validateCompleted(completed)
}

type responseStreamCall struct {
	id        string
	added     *responses.ResponseFunctionToolCall
	fragments strings.Builder
	hasDelta  bool
	arguments *string
	done      *responses.ResponseFunctionToolCall
}

type responseStreamCalls map[int64]*responseStreamCall

func (calls responseStreamCalls) at(index int64, id string) (*responseStreamCall, error) {
	if index < 0 || id == "" {
		return nil, fail("responses_tools_stream", fmt.Sprintf("invalid function_call item_id %q or output_index %d", id, index))
	}
	call := calls[index]
	if call == nil {
		call = &responseStreamCall{id: id}
		calls[index] = call
	} else if call.id != id {
		return nil, fail("responses_tools_stream", fmt.Sprintf("function_call at output_index %d changed item_id from %q to %q", index, call.id, id))
	}
	return call, nil
}

func (calls responseStreamCalls) validateCompleted(response *responses.Response) error {
	seenIDs := make(map[string]bool)
	seenCallIDs := make(map[string]bool)
	for index, item := range response.Output {
		if item.Type != "function_call" {
			continue
		}
		final := item.AsFunctionCall()
		if err := validateResponseFunctionToolCall("responses_tools_stream", final); err != nil {
			return err
		}
		if seenIDs[final.ID] {
			return fail("responses_tools_stream", fmt.Sprintf("duplicate function_call item_id %q in completed response", final.ID))
		}
		if seenCallIDs[final.CallID] {
			return fail("responses_tools_stream", fmt.Sprintf("duplicate function_call call_id %q in completed response", final.CallID))
		}
		seenIDs[final.ID] = true
		seenCallIDs[final.CallID] = true
		call := calls[int64(index)]
		if call == nil || call.id != final.ID || call.done == nil || call.arguments == nil {
			return fail("responses_tools_stream", fmt.Sprintf("function_call %q at output_index %d missing matching arguments.done or output_item.done", final.ID, index))
		}
		if err := validateResponseFunctionToolCall("responses_tools_stream", *call.done); err != nil {
			return err
		}
		if call.done.CallID != final.CallID || call.done.Name != final.Name || call.done.Arguments != final.Arguments || *call.arguments != final.Arguments {
			return fail("responses_tools_stream", fmt.Sprintf("function_call %q differs between stream and completed response", final.ID))
		}
		if call.added != nil && (call.added.CallID != final.CallID || call.added.Name != final.Name) {
			return fail("responses_tools_stream", fmt.Sprintf("function_call %q changed identity after output_item.added", final.ID))
		}
		if call.hasDelta && call.fragments.String() != final.Arguments {
			return fail("responses_tools_stream", fmt.Sprintf("argument deltas for function_call %q differ from finalized arguments", final.ID))
		}
	}
	if len(seenIDs) != len(calls) {
		return fail("responses_tools_stream", "streamed function_call missing from completed response")
	}
	if len(calls) == 0 && responseOutputRefusal(response) == "" {
		return fail("responses_tools_stream", "completed response produced no function_call or refusal")
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
	// openai-go v3.61+ dropped Name from this event; tool name is validated on
	// response.output_item.done via validateResponseFunctionToolCall.
	if done.ItemID == "" {
		return fail(suite, "response.function_call_arguments.done missing item_id")
	}
	if !done.JSON.OutputIndex.Valid() {
		return fail(suite, "response.function_call_arguments.done missing output_index")
	}
	if !done.JSON.SequenceNumber.Valid() {
		return fail(suite, "response.function_call_arguments.done missing sequence_number")
	}
	if !done.JSON.Arguments.Valid() {
		return fail(suite, "response.function_call_arguments.done missing arguments")
	}
	return nil
}
