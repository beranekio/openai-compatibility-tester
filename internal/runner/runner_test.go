package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		BaseURL:            server.BaseURL(),
		APIKey:             "test-key",
		AdminAPIKey:        "test-admin-key",
		Model:              "gpt-4o-mini",
		ResponsesModel:     "gpt-4o-mini",
		CompletionModel:    config.DefaultCompletionModel,
		VisionModel:        "gpt-4o-mini",
		ReasoningModel:     "o3-mini",
		ImageModel:         "dall-e-2",
		VideoModel:         "sora-2",
		TTSModel:           "tts-1",
		WhisperModel:       "whisper-1",
		TranscriptionModel: "gpt-4o-mini-transcribe",
		EmbeddingModel:      "text-embedding-3-small",
		RealtimeModel:       "gpt-realtime",
		ChatKitWorkflowID:   config.DefaultChatKitWorkflowID,
		ChatKitTestThreadID: "cthr_mock_1",
		RequestTimeout:      30 * time.Second,
		Suites: []string{
			"models",
			"models_get",
			"chat_completions",
			"chat_completions_stream_usage",
			"chat_completions_logprobs",
			"chat_completions_json",
			"chat_completions_vision",
			"chat_completions_reasoning",
			"chat_completions_audio",
			"chat_completions_tools",
			"chat_completions_tools_stream",
			"chat_completions_multi_turn",
			"chat_completions_get",
			"chat_completions_list",
			"chat_completions_delete",
			"chat_completions_messages",
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
			"responses_compact",
			"responses_input_tokens",
			"moderations",
			"images_generations",
			"images_edits",
			"images_variations",
			"audio_speech",
			"audio_transcriptions",
			"audio_transcriptions_stream",
			"audio_translations",
			"files",
			"uploads",
			"batches_create",
			"batches_get",
			"batches_cancel",
			"conversations",
			"vector_stores",
			"vector_store_files",
			"vector_store_file_batches",
			"realtime_client_secrets",
			"containers",
			"container_files",
			"videos",
			"skills",
			"skill_versions",
			"fine_tuning",
			"chatkit_sessions",
			"chatkit_threads",
			"assistants",
			"assistants_threads",
			"error_responses",
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

func TestErrorResponsesPassesAgainstMockServer(t *testing.T) {
	server := mockserver.New()
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:        server.BaseURL(),
		APIKey:         "test-key",
		Model:          "gpt-4o-mini",
		RequestTimeout: 30 * time.Second,
		Suites:         []string{"error_responses"},
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

func TestErrorResponsesFailsOnBrokenServer(t *testing.T) {
	server := mockserver.BrokenServer()
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:        server.BaseURL(),
		APIKey:         "test-key",
		Model:          "gpt-4o-mini",
		RequestTimeout: 30 * time.Second,
		Suites:         []string{"error_responses"},
	}

	runner := New(cfg)
	runner.Output = &bytes.Buffer{}

	results, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if code := ExitCode(results); code != 1 {
		t.Fatalf("ExitCode() = %d, want 1; summary:\n%s", code, FormatSummary(results))
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

func TestRunPassesWithOrgAndProjectConfigured(t *testing.T) {
	server := mockserver.New()
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:        server.BaseURL(),
		APIKey:         "test-key",
		OrgID:          "org-smoke-test",
		ProjectID:      "proj-smoke-test",
		Model:          "gpt-4o-mini",
		RequestTimeout: 30 * time.Second,
		Suites:         []string{"models"},
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

func TestRunnerSendsOrgAndProjectHeaders(t *testing.T) {
	const (
		wantOrg     = "org-header-test"
		wantProject = "proj-header-test"
	)

	headers := make(chan struct {
		org     string
		project string
	}, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/models", func(w http.ResponseWriter, r *http.Request) {
		headers <- struct {
			org     string
			project string
		}{
			org:     r.Header.Get("OpenAI-Organization"),
			project: r.Header.Get("OpenAI-Project"),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]any{{
				"id":       "gpt-4o-mini",
				"object":   "model",
				"created":  1700000000,
				"owned_by": "mock",
			}},
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:        server.URL + "/v1",
		APIKey:         "test-key",
		OrgID:          wantOrg,
		ProjectID:      wantProject,
		Model:          "gpt-4o-mini",
		RequestTimeout: 30 * time.Second,
		Suites:         []string{"models"},
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
	var got struct {
		org     string
		project string
	}
	select {
	case got = <-headers:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for models request headers")
	}
	if got.org != wantOrg {
		t.Fatalf("OpenAI-Organization = %q, want %q", got.org, wantOrg)
	}
	if got.project != wantProject {
		t.Fatalf("OpenAI-Project = %q, want %q", got.project, wantProject)
	}
}

func TestRunnerClearsOrgAndProjectHeadersWhenUnset(t *testing.T) {
	t.Setenv("OPENAI_ORG_ID", "org-env")
	t.Setenv("OPENAI_PROJECT_ID", "proj-env")

	headers := make(chan struct {
		org     string
		project string
	}, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/models", func(w http.ResponseWriter, r *http.Request) {
		headers <- struct {
			org     string
			project string
		}{
			org:     r.Header.Get("OpenAI-Organization"),
			project: r.Header.Get("OpenAI-Project"),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]any{{
				"id":       "gpt-4o-mini",
				"object":   "model",
				"created":  1700000000,
				"owned_by": "mock",
			}},
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	cfg := &config.Config{
		BaseURL:        server.URL + "/v1",
		APIKey:         "test-key",
		Model:          "gpt-4o-mini",
		RequestTimeout: 30 * time.Second,
		Suites:         []string{"models"},
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

	var got struct {
		org     string
		project string
	}
	select {
	case got = <-headers:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for models request headers")
	}
	if got.org != "" {
		t.Fatalf("OpenAI-Organization = %q, want empty", got.org)
	}
	if got.project != "" {
		t.Fatalf("OpenAI-Project = %q, want empty", got.project)
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
