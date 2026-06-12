package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	EnvBaseURL         = "OPENAI_BASE_URL"
	EnvAPIKey          = "OPENAI_API_KEY"
	EnvModel           = "OPENAI_MODEL"
	EnvCompletionModel = "OPENAI_COMPLETION_MODEL"
	EnvEmbeddingModel  = "OPENAI_EMBEDDING_MODEL"
	EnvTestSuites      = "TEST_SUITES"
	EnvRequestTimeout  = "REQUEST_TIMEOUT"
)

// DefaultSuites are run when TEST_SUITES is unset or set to "all".
var DefaultSuites = []string{
	"models",
	"chat_completions",
	"chat_completions_stream",
	"embeddings",
}

// Config holds runtime settings for compatibility testing.
type Config struct {
	BaseURL         string
	APIKey          string
	Model           string
	CompletionModel string
	EmbeddingModel  string
	Suites          []string
	RequestTimeout  time.Duration
	ListSuites      bool
}

// Load parses configuration from environment variables and command-line flags.
func Load(args []string) (*Config, error) {
	defaultTimeout, err := envDurationOrDefault(EnvRequestTimeout, 2*time.Minute)
	if err != nil {
		return nil, err
	}

	fs := flag.NewFlagSet("openai-compatibility-tester", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	baseURL := fs.String("base-url", envOrDefault(EnvBaseURL, ""), "OpenAI-compatible API base URL")
	apiKey := fs.String("api-key", envOrDefault(EnvAPIKey, ""), "API key for the endpoint")
	model := fs.String("model", envOrDefault(EnvModel, "gpt-4o-mini"), "Model for chat and responses suites")
	completionModel := fs.String("completion-model", envOrDefault(EnvCompletionModel, ""), "Model for legacy completions suite (defaults to --model)")
	embeddingModel := fs.String("embedding-model", envOrDefault(EnvEmbeddingModel, "text-embedding-3-small"), "Model for embedding tests")
	suites := fs.String("suites", envOrDefault(EnvTestSuites, "all"), "Comma-separated suite names to run, or 'all'")
	timeout := fs.Duration("timeout", defaultTimeout, "Per-request timeout")
	listSuites := fs.Bool("list-suites", false, "List available test suites and exit")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &Config{
		BaseURL:        strings.TrimRight(strings.TrimSpace(*baseURL), "/"),
		APIKey:         strings.TrimSpace(*apiKey),
		Model:          strings.TrimSpace(*model),
		CompletionModel: strings.TrimSpace(*completionModel),
		EmbeddingModel: strings.TrimSpace(*embeddingModel),
		RequestTimeout: *timeout,
		ListSuites:     *listSuites,
	}

	if cfg.CompletionModel == "" {
		cfg.CompletionModel = cfg.Model
	}

	if *suites == "all" {
		cfg.Suites = append([]string(nil), DefaultSuites...)
	} else {
		for _, name := range strings.Split(*suites, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				cfg.Suites = append(cfg.Suites, name)
			}
		}
	}

	if cfg.ListSuites {
		return cfg, nil
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("%s or --base-url is required", EnvBaseURL)
	}
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return nil, err
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%s or --api-key is required", EnvAPIKey)
	}
	if len(cfg.Suites) == 0 {
		return nil, fmt.Errorf("at least one test suite must be selected")
	}

	return cfg, nil
}

func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s: invalid URL: %w", EnvBaseURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s: URL must use http or https scheme", EnvBaseURL)
	}
	if u.Host == "" {
		return fmt.Errorf("%s: URL must include a host", EnvBaseURL)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration %q: %w", key, value, err)
	}
	return d, nil
}