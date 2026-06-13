package suites

import (
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

func responseOutputRefusal(resp *responses.Response) string {
	if resp == nil {
		return ""
	}
	var b strings.Builder
	for _, item := range resp.Output {
		for _, content := range item.Content {
			if content.Type == "refusal" {
				b.WriteString(content.Refusal)
			}
		}
	}
	return b.String()
}

func hasResponseOutput(resp *responses.Response) bool {
	if resp == nil {
		return false
	}
	return resp.OutputText() != "" || responseOutputRefusal(resp) != ""
}

func hasChatMessageOutput(msg openai.ChatCompletionMessage) bool {
	return msg.Content != "" || msg.Refusal != ""
}

func isContentFilterFinishReason(reason string) bool {
	return reason == "content_filter"
}

func isContentFilterIncompleteResponse(resp *responses.Response) bool {
	if resp == nil {
		return false
	}
	return string(resp.Status) == "incomplete" &&
		resp.JSON.IncompleteDetails.Valid() &&
		resp.IncompleteDetails.Reason == "content_filter"
}

func hasToolCalls(calls []openai.ChatCompletionMessageToolCallUnion) bool {
	return len(calls) > 0
}

func isToolCallsFinishReason(reason string) bool {
	return reason == "tool_calls"
}

type accumulatedToolCall struct {
	id        string
	callType  string
	name      string
	arguments string
}

func accumulateDeltaToolCalls(acc map[int]*accumulatedToolCall, calls []openai.ChatCompletionChunkChoiceDeltaToolCall) {
	for _, call := range calls {
		entry := acc[int(call.Index)]
		if entry == nil {
			entry = &accumulatedToolCall{}
			acc[int(call.Index)] = entry
		}
		if call.ID != "" {
			entry.id = call.ID
		}
		if call.Type != "" {
			entry.callType = call.Type
		}
		if call.Function.Name != "" {
			entry.name = call.Function.Name
		}
		if call.Function.Arguments != "" {
			entry.arguments += call.Function.Arguments
		}
	}
}

func validateAccumulatedToolCall(suite string, call *accumulatedToolCall) error {
	if call == nil {
		return fail(suite, "tool call missing")
	}
	if call.id == "" {
		return fail(suite, "tool call missing id")
	}
	if call.callType != "" && call.callType != "function" {
		return fail(suite, fmt.Sprintf("tool call type is %q, want function", call.callType))
	}
	if call.name == "" {
		return fail(suite, "tool call function missing name")
	}
	if call.name != weatherToolName {
		return fail(suite, fmt.Sprintf("tool call function name is %q, want %s", call.name, weatherToolName))
	}
	if call.arguments == "" {
		return fail(suite, "tool call function missing arguments")
	}
	return nil
}

func validateFunctionToolCall(suite string, call openai.ChatCompletionMessageToolCallUnion) error {
	fn := call.AsFunction()
	if fn.ID == "" {
		return fail(suite, "tool call missing id")
	}
	if fn.Type != "function" {
		return fail(suite, fmt.Sprintf("tool call type is %q, want function", fn.Type))
	}
	if fn.Function.Name == "" {
		return fail(suite, "tool call function missing name")
	}
	if fn.Function.Name != weatherToolName {
		return fail(suite, fmt.Sprintf("tool call function name is %q, want %s", fn.Function.Name, weatherToolName))
	}
	if fn.Function.Arguments == "" {
		return fail(suite, "tool call function missing arguments")
	}
	return nil
}