package suites

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func TestResponsesToolsStream(t *testing.T) {
	type obj = map[string]any
	const firstArgs = `{"location":"San Francisco"}`
	const secondArgs = `{"location":"Oakland"}`
	call := func(id, args string) obj {
		return obj{"type": "function_call", "id": id, "call_id": "call_" + id, "name": "get_weather", "status": "completed", "arguments": args}
	}
	itemEvent := func(kind string, index int, item obj) obj {
		return obj{"type": "response.output_item." + kind, "output_index": index, "item": item}
	}
	delta := func(index int, id, part string) obj {
		return obj{"type": "response.function_call_arguments.delta", "item_id": id, "output_index": index, "delta": part}
	}
	argsDone := func(index int, id, args string) obj {
		return obj{"type": "response.function_call_arguments.done", "item_id": id, "output_index": index, "arguments": args}
	}
	completed := func(output ...obj) obj {
		return obj{"type": "response.completed", "response": obj{"id": "resp_test", "status": "completed", "output": output}}
	}
	incomplete := func(reason string) obj {
		return obj{"type": "response.incomplete", "response": obj{"id": "resp_test", "status": "incomplete", "incomplete_details": obj{"reason": reason}}}
	}
	first, second := call("fc_1", firstArgs), call("fc_2", secondArgs)
	reasoning := obj{"type": "reasoning", "id": "rs_1", "summary": []obj{}}
	message := obj{"type": "message", "id": "msg_1", "role": "assistant", "status": "completed", "content": []obj{{"type": "output_text", "text": "Hello"}}}
	refusal := obj{"type": "message", "id": "msg_2", "role": "assistant", "status": "completed", "content": []obj{{"type": "refusal", "refusal": "I can't help."}}}
	reasoningDone := itemEvent("done", 0, reasoning)
	firstDone := itemEvent("done", 0, first)
	firstArgsDone := argsDone(0, "fc_1", firstArgs)
	duplicateCallID := maps.Clone(second)
	duplicateCallID["call_id"] = first["call_id"]
	changedCallID := maps.Clone(first)
	changedCallID["call_id"] = "other_call_id"
	added := call("fc_1", "")
	added["status"] = "in_progress"

	tests := []struct {
		name    string
		events  []obj
		wantErr string
	}{
		{"reasoning and message with function calls", []obj{
			reasoningDone, itemEvent("added", 1, added),
			delta(1, "fc_1", `{"location":`), delta(2, "fc_2", secondArgs),
			delta(1, "fc_1", `"San Francisco"}`), argsDone(2, "fc_2", secondArgs), itemEvent("done", 2, second),
			argsDone(1, "fc_1", firstArgs), itemEvent("done", 1, first), itemEvent("done", 3, message),
			completed(reasoning, first, second, message),
		}, ""},
		{"refusal", []obj{itemEvent("done", 0, refusal), completed(refusal)}, ""},
		{"filtered incomplete", []obj{reasoningDone, incomplete("content_filter")}, ""},
		{"filtered partial call", []obj{argsDone(0, "fc_1", `{"location":`), incomplete("content_filter")}, ""},
		{"reasoning only", []obj{reasoningDone, completed(reasoning)}, "no function_call or refusal"},
		{"text only", []obj{itemEvent("done", 0, message), completed(message)}, "no function_call or refusal"},
		{"different item IDs", []obj{argsDone(0, "fc_2", secondArgs), firstDone, completed(first)}, "changed item_id"},
		{"different output indices", []obj{argsDone(1, "fc_1", firstArgs), firstDone, completed(first)}, "missing matching"},
		{"duplicate call ID", []obj{
			firstArgsDone, firstDone, argsDone(1, "fc_2", secondArgs), itemEvent("done", 1, duplicateCallID),
			completed(first, duplicateCallID),
		}, "duplicate function_call call_id"},
		{"different final call ID", []obj{firstArgsDone, firstDone, completed(changedCallID)}, "differs between"},
		{"missing item done", []obj{firstArgsDone, completed(first)}, "missing matching"},
		{"missing arguments done", []obj{firstDone, completed(first)}, "missing matching"},
		{"missing terminal output", []obj{firstArgsDone, firstDone, completed()}, "missing from completed response"},
		{"inconsistent final arguments", []obj{firstArgsDone, firstDone, completed(call("fc_1", secondArgs))}, "differs between"},
		{"inconsistent deltas", []obj{delta(0, "fc_1", secondArgs), firstArgsDone, firstDone, completed(first)}, "argument deltas"},
		{"only refusal fragment", []obj{
			{"type": "response.refusal.delta", "item_id": "msg_1", "output_index": 0, "content_index": 0, "delta": "No"},
			completed(message),
		}, "no function_call or refusal"},
		{"duplicate item done", []obj{firstArgsDone, firstDone, firstDone}, "duplicate"},
		{"error", []obj{{"type": "response.failed", "response": obj{"id": "resp_test", "status": "failed"}}}, "failure event"},
		{"insufficient tokens", []obj{incomplete("max_output_tokens")}, "failure event"},
		{"missing terminal", []obj{firstArgsDone, firstDone}, "missing terminal"},
		{"event after terminal", []obj{completed(refusal), reasoningDone}, "after terminal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				for i, fixture := range tt.events {
					event := maps.Clone(fixture)
					event["sequence_number"] = i
					data, err := json.Marshal(event)
					if err != nil {
						t.Error(err)
						return
					}
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
			}))
			defer server.Close()
			client := openai.NewClient(option.WithBaseURL(server.URL), option.WithAPIKey("test"), option.WithMaxRetries(0))
			err := (ResponsesToolsStream{}).Run(t.Context(), client, &config.Config{ResponsesModel: "test-model"})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}
