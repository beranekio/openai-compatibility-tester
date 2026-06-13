package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/mockserver"
)

func TestRunAllPassesAgainstMockServer(t *testing.T) {
	server := mockserver.New()
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:          server.BaseURL(),
		APIKey:           "test-key",
		Model:            "gpt-4o-mini",
		CompletionModel:  config.DefaultCompletionModel,
		VisionModel:      "gpt-4o-mini",
		EmbeddingModel:   "text-embedding-3-small",
		RequestTimeout:   30 * time.Second,
		Suites: []string{
			"models",
			"models_get",
			"chat_completions",
			"chat_completions_json",
			"chat_completions_vision",
			"chat_completions_tools",
			"chat_completions_tools_stream",
			"completions",
			"completions_stream",
			"embeddings",
			"embeddings_batch",
			"responses",
			"responses_stream",
			"responses_tools",
			"responses_tools_stream",
			"responses_json",
			"responses_get",
			"responses_delete",
			"responses_cancel",
			"responses_input_items",
		},
	}

	runner := New(cfg)
	runner.Output = &bytes.Buffer{}

	results, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if code := ExitCode(results); code != 0 {
		t.Fatalf("ExitCode() = %d, want 0; summary:\n%s", code, FormatSummary(results))
	}
}

func TestRunAllFailsOnIncompatibleEndpoint(t *testing.T) {
	server := mockserver.BrokenServer()
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:        server.BaseURL(),
		APIKey:         "test-key",
		Model:          "gpt-4o-mini",
		RequestTimeout: 30 * time.Second,
		Suites:         []string{"models", "chat_completions"},
	}

	runner := New(cfg)
	runner.Output = &bytes.Buffer{}

	results, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if code := ExitCode(results); code != 1 {
		t.Fatalf("ExitCode() = %d, want 1", code)
	}
}

func TestRunRejectsUnknownSuiteBeforeRequests(t *testing.T) {
	server := mockserver.New()
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:        server.BaseURL(),
		APIKey:         "test-key",
		Model:          "gpt-4o-mini",
		RequestTimeout: 30 * time.Second,
		Suites:         []string{"models", "not-a-suite"},
	}

	runner := New(cfg)
	results, err := runner.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "unknown test suite") {
		t.Fatalf("expected unknown suite error, got results=%v err=%v", results, err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no suite results before validation, got %d", len(results))
	}
}

func TestRunRejectsUnknownSuite(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "https://example.com/v1",
		APIKey:  "test-key",
		Suites:  []string{"not-a-suite"},
	}

	runner := New(cfg)
	_, err := runner.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "unknown test suite") {
		t.Fatalf("expected unknown suite error, got %v", err)
	}
}