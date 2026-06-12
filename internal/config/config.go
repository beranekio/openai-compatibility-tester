package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	EnvBaseURL        = "OPENAI_BASE_URL"
	EnvAPIKey         = "OPENAI_API_KEY"
	EnvModel          = "OPENAI_MODEL"
	EnvEmbeddingModel = "OPENAI_EMBEDDING_MODEL"
	EnvTestSuites     = "TEST_SUITES"
	EnvRequestTimeout = "REQUEST_TIMEOUT"
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
	BaseURL        string
	APIKey         string
	Model          string
	EmbeddingModel string
	Suites         []string
	RequestTimeout time.Duration
	ListSuites     bool
}

// Load parses configuration from environment variables and command-line flags.
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("openai-compatibility-tester", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	baseURL := fs.String("base-url", envOrDefault(EnvBaseURL, ""), "OpenAI-compatible API base URL")
	apiKey := fs.String("api-key", envOrDefault(EnvAPIKey, ""), "API key for the endpoint")
	model := fs.String("model", envOrDefault(EnvModel, "gpt-4o-mini"), "Model for chat/completion/responses tests")
	embeddingModel := fs.String("embedding-model", envOrDefault(EnvEmbeddingModel, "text-embedding-3-small"), "Model for embedding tests")
	suites := fs.String("suites", envOrDefault(EnvTestSuites, "all"), "Comma-separated suite names to run, or 'all'")
	timeout := fs.Duration("timeout", envDurationOrDefault(EnvRequestTimeout, 2*time.Minute), "Per-request timeout")
	listSuites := fs.Bool("list-suites", false, "List available test suites and exit")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &Config{
		BaseURL:        strings.TrimRight(strings.TrimSpace(*baseURL), "/"),
		APIKey:         strings.TrimSpace(*apiKey),
		Model:          strings.TrimSpace(*model),
		EmbeddingModel: strings.TrimSpace(*embeddingModel),
		RequestTimeout: *timeout,
		ListSuites:     *listSuites,
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
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%s or --api-key is required", EnvAPIKey)
	}
	if len(cfg.Suites) == 0 {
		return nil, fmt.Errorf("at least one test suite must be selected")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := envOrDefault(key, "")
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}