package config

import (
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

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

func TestLoadSelectedSuites(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
	t.Setenv(EnvAPIKey, "test-key")

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

func TestLoadCompletionModelOverride(t *testing.T) {
	t.Setenv(EnvBaseURL, "http://example.com/v1")
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