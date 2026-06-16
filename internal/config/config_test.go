package config

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	cfg, err := Load([]string{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "https://example.com/v1" {
		t.Fatalf("BaseURL = %q, want https://example.com/v1", cfg.BaseURL)
	}
	if cfg.APIKey != "test-key" {
		t.Fatalf("APIKey = %q, want test-key", cfg.APIKey)
	}
	if cfg.CompletionModel != cfg.Model {
		t.Fatalf("CompletionModel = %q, want default %q", cfg.CompletionModel, cfg.Model)
	}
	if len(cfg.Suites) != len(DefaultSuites) {
		t.Fatalf("len(Suites) = %d, want %d", len(cfg.Suites), len(DefaultSuites))
	}
}

func TestLoadRequiresBaseURL(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{})
	if err == nil {
		t.Fatal("expected error when base URL is missing")
	}
}

func TestLoadRejectsInvalidBaseURL(t *testing.T) {
	t.Setenv(EnvBaseURL, "://host")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{})
	if err == nil {
		t.Fatal("expected error for malformed base URL")
	}
}

func TestLoadRejectsInvalidRequestTimeout(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvRequestTimeout, "2minutes")

	_, err := Load([]string{})
	if err == nil {
		t.Fatal("expected error for invalid request timeout")
	}
}

func TestLoadTimeoutFlagOverridesInvalidEnvironment(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvRequestTimeout, "typo")

	cfg, err := Load([]string{"--timeout", "30s"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.RequestTimeout.String() != "30s" {
		t.Fatalf("RequestTimeout = %s, want 30s", cfg.RequestTimeout)
	}
}

func TestLoadSingleDashTimeoutOverridesInvalidEnvironment(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvRequestTimeout, "typo")

	cfg, err := Load([]string{"-timeout=30s"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.RequestTimeout.String() != "30s" {
		t.Fatalf("RequestTimeout = %s, want 30s", cfg.RequestTimeout)
	}
}

func TestLoadListSuitesIgnoresInvalidRequestTimeout(t *testing.T) {
	t.Setenv(EnvRequestTimeout, "typo")

	cfg, err := Load([]string{"--list-suites"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.ListSuites {
		t.Fatal("expected ListSuites to be true")
	}
}

func TestLoadListSuitesIgnoresDuplicateSuites(t *testing.T) {
	t.Setenv(EnvTestSuites, "models,models")

	cfg, err := Load([]string{"--list-suites"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.ListSuites {
		t.Fatal("expected ListSuites to be true")
	}
}

func TestLoadRejectsBaseURLWithoutHostname(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--base-url", "http://:8080/v1"})
	if err == nil || !strings.Contains(err.Error(), "hostname") {
		t.Fatalf("expected hostname error, got %v", err)
	}
}

func TestLoadRejectsBaseURLWithInvalidPort(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--base-url", "http://host:99999/v1"})
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected port error, got %v", err)
	}
}

func TestLoadRejectsUnexpectedArguments(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--suites", "models", "typo"})
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("expected unexpected arguments error, got %v", err)
	}
}

func TestLoadRejectsBaseURLWithQuery(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--base-url", "https://host/v1?token=secret"})
	if err == nil || !strings.Contains(err.Error(), "query parameters") {
		t.Fatalf("expected query parameter error, got %v", err)
	}
}

func TestLoadRejectsNonPositiveTimeout(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--timeout", "0"})
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestLoadSelectedSuites(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")

	cfg, err := Load([]string{"--suites", "models,embeddings"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := []string{"models", "embeddings"}
	if len(cfg.Suites) != len(want) {
		t.Fatalf("len(Suites) = %d, want %d", len(cfg.Suites), len(want))
	}
	for i, suite := range want {
		if cfg.Suites[i] != suite {
			t.Fatalf("Suites[%d] = %q, want %q", i, cfg.Suites[i], suite)
		}
	}
}

func TestLoadRejectsDuplicateSuites(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--suites", "models,models"})
	if err == nil || !strings.Contains(err.Error(), "duplicate test suite") {
		t.Fatalf("expected duplicate suite error, got %v", err)
	}
}

func TestLoadAllowsModelsSuiteWithoutModelFlag(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "")

	_, err := Load([]string{"--suites", "models", "--model="})
	if err != nil {
		t.Fatalf("Load() error = %v, want models-only run without model", err)
	}
}

func TestLoadRejectsEmptyModelForChatSuite(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "")

	_, err := Load([]string{"--suites", "chat_completions", "--model="})
	if err == nil || !strings.Contains(err.Error(), EnvModel) {
		t.Fatalf("expected missing model error, got %v", err)
	}
}

func TestLoadAllowsModelsGetSuiteWithModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	cfg, err := Load([]string{"--suites", "models_get", "--model", "my-model"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Model != "my-model" {
		t.Fatalf("Model = %q, want my-model", cfg.Model)
	}
}

func TestLoadRejectsEmptyModelForModelsGetSuite(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "")

	_, err := Load([]string{"--suites", "models_get", "--model="})
	if err == nil || !strings.Contains(err.Error(), EnvModel) {
		t.Fatalf("expected missing model error, got %v", err)
	}
}

func TestLoadRejectsEmptyModelForBatchesCreateSuite(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "")

	_, err := Load([]string{"--suites", "batches_create", "--model="})
	if err == nil || !strings.Contains(err.Error(), EnvModel) {
		t.Fatalf("expected missing model error, got %v", err)
	}
}

func TestLoadAllowsBatchesCreateSuiteWithModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	cfg, err := Load([]string{"--suites", "batches_create", "--model", "my-model"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Model != "my-model" {
		t.Fatalf("Model = %q, want my-model", cfg.Model)
	}
}

func TestLoadRejectsEmptyCompletionModelFlag(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvCompletionModel, "")

	_, err := Load([]string{"--suites", "completions", "--completion-model="})
	if err == nil || !strings.Contains(err.Error(), EnvCompletionModel) {
		t.Fatalf("expected missing completion model error, got %v", err)
	}
}

func TestLoadHelpDoesNotExposeAPIKey(t *testing.T) {
	t.Setenv(EnvAPIKey, "super-secret-key")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = w

	outputDone := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputDone <- buf.String()
	}()

	_, err = Load([]string{"-h"})
	w.Close()
	os.Stderr = oldStderr
	output := <-outputDone

	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected ErrHelp, got %v", err)
	}
	if strings.Contains(output, "super-secret-key") {
		t.Fatalf("help output leaked API key: %q", output)
	}
}

func TestLoadRejectsWhitespaceOnlyCompletionModelFlag(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvCompletionModel, "")

	_, err := Load([]string{"--suites", "completions", "--completion-model", "   "})
	if err == nil || !strings.Contains(err.Error(), EnvCompletionModel) {
		t.Fatalf("expected missing completion model error, got %v", err)
	}
}

func TestLoadRejectsEmptyCompletionModelWhenLastFlagWins(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvCompletionModel, "")

	_, err := Load([]string{"--suites", "completions", "--completion-model=custom", "--completion-model="})
	if err == nil || !strings.Contains(err.Error(), EnvCompletionModel) {
		t.Fatalf("expected missing completion model error, got %v", err)
	}
}

func TestLoadRejectsBaseURLWithEncodedSlash(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--base-url", "https://host/proxy%2Ftenant/v1"})
	if err == nil || !strings.Contains(err.Error(), "%2F") {
		t.Fatalf("expected encoded slash error, got %v", err)
	}
}

func TestLoadCompletionsSuiteUsesInstructDefault(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvCompletionModel, "")

	cfg, err := Load([]string{"--suites", "completions"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CompletionModel != DefaultCompletionModel {
		t.Fatalf("CompletionModel = %q, want %q", cfg.CompletionModel, DefaultCompletionModel)
	}
}

func TestLoadListSuitesAllowsEmptyAPIKeyFlag(t *testing.T) {
	t.Setenv(EnvAPIKey, "production-secret")

	cfg, err := Load([]string{"--list-suites", "--api-key="})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.ListSuites {
		t.Fatal("expected ListSuites to be true")
	}
}

func TestLoadAcceptsDashPrefixedAPIKeyFlag(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "")

	cfg, err := Load([]string{"--api-key", "-dash-prefixed-token"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "-dash-prefixed-token" {
		t.Fatalf("APIKey = %q, want -dash-prefixed-token", cfg.APIKey)
	}
}

func TestLoadRejectsEmptyAPIKeyFlagWhenEnvSet(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "production-secret")
	_, err := Load([]string{"--api-key="})
	if err == nil || !strings.Contains(err.Error(), EnvAPIKey) {
		t.Fatalf("expected missing API key error, got %v", err)
	}
}

func TestLoadRejectsEmbeddingsSuiteWithoutEmbeddingModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "")

	_, err := Load([]string{"--suites", "embeddings"})
	if err == nil || !strings.Contains(err.Error(), EnvEmbeddingModel) {
		t.Fatalf("expected missing embedding model error, got %v", err)
	}
}

func TestLoadRejectsEmbeddingsBatchSuiteWithoutEmbeddingModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "")

	_, err := Load([]string{"--suites", "embeddings_batch"})
	if err == nil || !strings.Contains(err.Error(), EnvEmbeddingModel) {
		t.Fatalf("expected missing embedding model error, got %v", err)
	}
}

func TestLoadAllowsImagesVariationsSuiteWithoutImageModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvImageModel, "")

	cfg, err := Load([]string{"--suites", "images_variations"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Suites) != 1 || cfg.Suites[0] != "images_variations" {
		t.Fatalf("Suites = %v, want [images_variations]", cfg.Suites)
	}
}

func TestLoadAllowsConversationsSuiteWithoutModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "")

	cfg, err := Load([]string{"--suites", "conversations", "--model="})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Suites) != 1 || cfg.Suites[0] != "conversations" {
		t.Fatalf("Suites = %v, want [conversations]", cfg.Suites)
	}
}

func TestLoadRejectsImagesEditsSuiteWithoutImageModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvImageModel, "")

	_, err := Load([]string{"--suites", "images_edits"})
	if err == nil || !strings.Contains(err.Error(), EnvImageModel) {
		t.Fatalf("expected missing image model error, got %v", err)
	}
}

func TestLoadRejectsAudioTranscriptionsStreamWithoutTranscriptionModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvTranscriptionModel, "")

	_, err := Load([]string{"--suites", "audio_transcriptions_stream"})
	if err == nil || !strings.Contains(err.Error(), EnvTranscriptionModel) {
		t.Fatalf("expected missing transcription model error, got %v", err)
	}
}

func TestLoadAllowsAudioTranscriptionsWithoutTranscriptionModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvWhisperModel, "whisper-1")
	t.Setenv(EnvTranscriptionModel, "")

	cfg, err := Load([]string{"--suites", "audio_transcriptions"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Suites) != 1 || cfg.Suites[0] != "audio_transcriptions" {
		t.Fatalf("Suites = %v, want [audio_transcriptions]", cfg.Suites)
	}
}

func TestLoadAllowsLoopbackHTTP(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://127.0.0.1:4010/v1")
	t.Setenv(EnvAPIKey, "test-key")

	cfg, err := Load([]string{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BaseURL != "http://127.0.0.1:4010/v1" {
		t.Fatalf("BaseURL = %q, want loopback HTTP URL", cfg.BaseURL)
	}
}

func TestLoadRejectsPlaintextHTTPForRemoteHost(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--base-url", "http://example.com/v1"})
	if err == nil || !strings.Contains(err.Error(), "plaintext HTTP") {
		t.Fatalf("expected plaintext HTTP error, got %v", err)
	}
}

func TestLoadAllowsPlaintextHTTPWithExplicitOptIn(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvAPIKey, "test-key")

	cfg, err := Load([]string{"--base-url", "http://example.com/v1", "--allow-insecure-http"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.AllowInsecureHTTP {
		t.Fatal("expected AllowInsecureHTTP to be true")
	}
}

func TestLoadResponsesModelOverride(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "chat-model")
	t.Setenv(EnvResponsesModel, "responses-model")

	cfg, err := Load([]string{"--suites", "responses"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ResponsesModel != "responses-model" {
		t.Fatalf("ResponsesModel = %q, want responses-model", cfg.ResponsesModel)
	}
	if cfg.Model != "chat-model" {
		t.Fatalf("Model = %q, want chat-model", cfg.Model)
	}
}

func TestLoadResponsesToolsSuitesUseResponsesModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "chat-model")
	t.Setenv(EnvResponsesModel, "responses-model")

	cfg, err := Load([]string{"--suites", "responses_tools,responses_tools_stream"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ResponsesModel != "responses-model" {
		t.Fatalf("ResponsesModel = %q, want responses-model", cfg.ResponsesModel)
	}
	if len(cfg.Suites) != 2 {
		t.Fatalf("len(Suites) = %d, want 2", len(cfg.Suites))
	}
}

func TestLoadResponsesToolsSuitesInExtendedAndFullPresets(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")
	t.Setenv(EnvImageModel, "dall-e-2")
	t.Setenv(EnvTTSModel, "tts-1")
	t.Setenv(EnvWhisperModel, "whisper-1")
	t.Setenv(EnvTranscriptionModel, "gpt-4o-mini-transcribe")
	t.Setenv(EnvReasoningModel, "o3-mini")

	for _, preset := range []string{"extended", "full"} {
		cfg, err := Load([]string{"--suites", preset})
		if err != nil {
			t.Fatalf("Load(%q) error = %v", preset, err)
		}
		seen := make(map[string]struct{}, len(cfg.Suites))
		for _, name := range cfg.Suites {
			seen[name] = struct{}{}
		}
		for _, name := range []string{"responses_tools", "responses_tools_stream"} {
			if _, ok := seen[name]; !ok {
				t.Fatalf("preset %q missing suite %q", preset, name)
			}
		}
	}
}

func TestLoadCompletionModelOverride(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "gpt-4o-mini")
	t.Setenv(EnvCompletionModel, "gpt-3.5-turbo-instruct")

	cfg, err := Load([]string{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CompletionModel != "gpt-3.5-turbo-instruct" {
		t.Fatalf("CompletionModel = %q, want gpt-3.5-turbo-instruct", cfg.CompletionModel)
	}
	if !strings.Contains(cfg.Model, "gpt-4o-mini") {
		t.Fatalf("Model = %q, want gpt-4o-mini", cfg.Model)
	}
}

func TestLoadCompletionsStreamSuiteUsesInstructDefault(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvCompletionModel, "")

	cfg, err := Load([]string{"--suites", "completions_stream"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CompletionModel != DefaultCompletionModel {
		t.Fatalf("CompletionModel = %q, want %q", cfg.CompletionModel, DefaultCompletionModel)
	}
}

func TestLoadSuitePresets(t *testing.T) {
	tests := []struct {
		name       string
		preset     string
		wantSuites []string
		setup      func(t *testing.T)
	}{
		{
			name:       "extended preset",
			preset:     "extended",
			wantSuites: ExtendedSuites,
			setup: func(t *testing.T) {
				t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")
				t.Setenv(EnvImageModel, "dall-e-2")
				t.Setenv(EnvTTSModel, "tts-1")
				t.Setenv(EnvWhisperModel, "whisper-1")
				t.Setenv(EnvTranscriptionModel, "gpt-4o-mini-transcribe")
				t.Setenv(EnvReasoningModel, "o3-mini")
			},
		},
		{
			name:       "full preset",
			preset:     "full",
			wantSuites: FullSuites,
			setup: func(t *testing.T) {
				t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")
				t.Setenv(EnvImageModel, "dall-e-2")
				t.Setenv(EnvTTSModel, "tts-1")
				t.Setenv(EnvWhisperModel, "whisper-1")
				t.Setenv(EnvTranscriptionModel, "gpt-4o-mini-transcribe")
				t.Setenv(EnvReasoningModel, "o3-mini")
			},
		},
		{
			name:       "default preset alias",
			preset:     "default",
			wantSuites: DefaultSuites,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvBaseURL, "https://example.com/v1")
			t.Setenv(EnvAPIKey, "test-key")
			if tt.setup != nil {
				tt.setup(t)
			}

			cfg, err := Load([]string{"--suites", tt.preset})
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if len(cfg.Suites) != len(tt.wantSuites) {
				t.Fatalf("len(Suites) = %d, want %d", len(cfg.Suites), len(tt.wantSuites))
			}
			for i, name := range tt.wantSuites {
				if cfg.Suites[i] != name {
					t.Fatalf("Suites[%d] = %q, want %q", i, cfg.Suites[i], name)
				}
			}
		})
	}
}

func TestLoadRejectsExplicitlyEmptySuiteSelection(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--suites", ""})
	if err == nil || !strings.Contains(err.Error(), "at least one test suite must be selected") {
		t.Fatalf("expected empty suite selection error, got %v", err)
	}
}

func TestLoadVisionModelDefaultsToChatModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "gpt-4o")

	cfg, err := Load([]string{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.VisionModel != "gpt-4o" {
		t.Fatalf("VisionModel = %q, want gpt-4o", cfg.VisionModel)
	}
}

func TestLoadRejectsReasoningSuiteWithoutReasoningModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvReasoningModel, "")

	_, err := Load([]string{"--suites", "chat_completions_reasoning"})
	if err == nil || !strings.Contains(err.Error(), EnvReasoningModel) {
		t.Fatalf("expected missing reasoning model error, got %v", err)
	}
}

func TestLoadAllowsReasoningSuiteWithExplicitReasoningModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvReasoningModel, "o3-mini")

	cfg, err := Load([]string{"--suites", "chat_completions_reasoning"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ReasoningModel != "o3-mini" {
		t.Fatalf("ReasoningModel = %q, want o3-mini", cfg.ReasoningModel)
	}
}
