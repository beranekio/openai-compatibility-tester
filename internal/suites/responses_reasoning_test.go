package suites

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3/responses"
)

func TestValidateResponsesReasoningResultRejectsNonContentFilterIncomplete(t *testing.T) {
	raw := `{
		"id": "resp-1",
		"object": "response",
		"created_at": 1700000000,
		"model": "o3-mini",
		"status": "incomplete",
		"incomplete_details": {"reason": "max_output_tokens"},
		"output": [{
			"id": "msg-1",
			"type": "message",
			"role": "assistant",
			"status": "incomplete",
			"content": [{"type": "output_text", "text": "pong"}]
		}]
	}`
	var resp responses.Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	err := validateResponsesReasoningResult("responses_reasoning", &resp)
	if err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("validateResponsesReasoningResult() error = %v, want incomplete status failure", err)
	}
}

func TestValidateResponsesReasoningResultAcceptsCompletedOutput(t *testing.T) {
	raw := `{
		"id": "resp-1",
		"object": "response",
		"created_at": 1700000000,
		"model": "o3-mini",
		"status": "completed",
		"output": [{
			"id": "msg-1",
			"type": "message",
			"role": "assistant",
			"status": "completed",
			"content": [{"type": "output_text", "text": "pong"}]
		}]
	}`
	var resp responses.Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if err := validateResponsesReasoningResult("responses_reasoning", &resp); err != nil {
		t.Fatalf("validateResponsesReasoningResult() error = %v", err)
	}
}
