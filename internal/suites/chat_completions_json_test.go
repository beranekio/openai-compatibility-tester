package suites

import (
	"strings"
	"testing"
)

func TestValidateStructuredAnswerJSON(t *testing.T) {
	if err := validateStructuredAnswerJSON("chat_completions_json", `{"answer":"pong"}`); err != nil {
		t.Fatalf("validateStructuredAnswerJSON() error = %v", err)
	}
	if err := validateStructuredAnswerJSON("chat_completions_json", `{"answer":"pong","unexpected":true}`); err == nil {
		t.Fatal("expected error for extra top-level field")
	}
	if err := validateStructuredAnswerJSON("chat_completions_json", `not-json`); err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("expected JSON parse error, got %v", err)
	}
}