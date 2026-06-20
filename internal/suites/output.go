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

func validateResponseEnvelope(suite string, resp *responses.Response) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if resp.ID == "" {
		return fail(suite, "response missing id")
	}
	if !resp.JSON.CreatedAt.Valid() {
		return fail(suite, "response missing created_at")
	}
	if resp.Model == "" {
		return fail(suite, "response missing model")
	}
	if string(resp.Object) != "response" {
		return fail(suite, fmt.Sprintf("response object is %q, want response", resp.Object))
	}
	return nil
}

func validateChatCompletionEnvelope(suite string, resp *openai.ChatCompletion) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if resp.ID == "" {
		return fail(suite, "response missing id")
	}
	if !resp.JSON.Created.Valid() {
		return fail(suite, "response missing created")
	}
	if resp.Model == "" {
		return fail(suite, "response missing model")
	}
	if string(resp.Object) != "chat.completion" {
		return fail(suite, fmt.Sprintf("response object is %q, want chat.completion", resp.Object))
	}
	return nil
}

func validateChatCompletionChoice(suite string, choice openai.ChatCompletionChoice) error {
	if !choice.JSON.Index.Valid() {
		return fail(suite, "choice missing index")
	}
	if choice.FinishReason == "" {
		return fail(suite, "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail(suite, fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if !hasChatMessageOutput(choice.Message) && !isContentFilterFinishReason(choice.FinishReason) {
		return fail(suite, "choice message has no content or refusal")
	}
	return nil
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
	return validateWeatherToolArguments(suite, call.arguments)
}

func hasResponseFunctionCalls(resp *responses.Response) bool {
	return len(responseFunctionCalls(resp)) > 0
}

func responseFunctionCalls(resp *responses.Response) []responses.ResponseFunctionToolCall {
	if resp == nil {
		return nil
	}
	var calls []responses.ResponseFunctionToolCall
	for _, item := range resp.Output {
		if item.Type == "function_call" {
			calls = append(calls, item.AsFunctionCall())
		}
	}
	return calls
}

func validateResponseFunctionToolCall(suite string, call responses.ResponseFunctionToolCall) error {
	if call.ID == "" {
		return fail(suite, "function_call missing id")
	}
	if call.CallID == "" {
		return fail(suite, "function_call missing call_id")
	}
	if !call.JSON.Status.Valid() {
		return fail(suite, "function_call missing status")
	}
	if call.Status != responses.ResponseFunctionToolCallStatusCompleted {
		return fail(suite, fmt.Sprintf("function_call status is %q, want completed", call.Status))
	}
	if call.Name == "" {
		return fail(suite, "function_call missing name")
	}
	if call.Name != weatherToolName {
		return fail(suite, fmt.Sprintf("function_call name is %q, want %s", call.Name, weatherToolName))
	}
	return validateWeatherToolArguments(suite, call.Arguments)
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
	return validateWeatherToolArguments(suite, fn.Function.Arguments)
}
