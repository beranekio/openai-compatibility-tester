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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")

	cfg, err := Load([]string{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "http://example.com/v1" {
		t.Fatalf("BaseURL = %q, want http://example.com/v1", cfg.BaseURL)
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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvRequestTimeout, "2minutes")

	_, err := Load([]string{})
	if err == nil {
		t.Fatal("expected error for invalid request timeout")
	}
}

func TestLoadTimeoutFlagOverridesInvalidEnvironment(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")
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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")
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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--timeout", "0"})
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestLoadSelectedSuites(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

	_, err := Load([]string{"--suites", "models,models"})
	if err == nil || !strings.Contains(err.Error(), "duplicate test suite") {
		t.Fatalf("expected duplicate suite error, got %v", err)
	}
}

func TestLoadAllowsModelsSuiteWithoutModelFlag(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "")

	_, err := Load([]string{"--suites", "models", "--model="})
	if err != nil {
		t.Fatalf("Load() error = %v, want models-only run without model", err)
	}
}

func TestLoadRejectsEmptyModelForChatSuite(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "")

	_, err := Load([]string{"--suites", "chat_completions", "--model="})
	if err == nil || !strings.Contains(err.Error(), EnvModel) {
		t.Fatalf("expected missing model error, got %v", err)
	}
}

func TestLoadRejectsEmptyCompletionModelFlag(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvCompletionModel, "")

	_, err := Load([]string{"--suites", "completions", "--completion-model", "   "})
	if err == nil || !strings.Contains(err.Error(), EnvCompletionModel) {
		t.Fatalf("expected missing completion model error, got %v", err)
	}
}

func TestLoadRejectsEmptyCompletionModelWhenLastFlagWins(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
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
	t.Setenv(EnvBaseURL, "http://example.com/v1")
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

func TestLoadRejectsEmptyAPIKeyFlagWhenEnvSet(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "production-secret")
	t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")

	_, err := Load([]string{"--api-key="})
	if err == nil || !strings.Contains(err.Error(), EnvAPIKey) {
		t.Fatalf("expected missing API key error, got %v", err)
	}
}

func TestLoadRejectsDefaultSuitesWithoutEmbeddingModel(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvEmbeddingModel, "")

	_, err := Load([]string{})
	if err == nil || !strings.Contains(err.Error(), EnvEmbeddingModel) {
		t.Fatalf("expected missing embedding model error, got %v", err)
	}
}

func TestLoadCompletionModelOverride(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")
	t.Setenv(EnvModel, "gpt-4o-mini")
	t.Setenv(EnvCompletionModel, "gpt-3.5-turbo-instruct")
	t.Setenv(EnvEmbeddingModel, "text-embedding-3-small")

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