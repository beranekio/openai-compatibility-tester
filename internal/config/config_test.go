package config

import (
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