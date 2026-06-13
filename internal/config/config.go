package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
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

	// DefaultCompletionModel is used when the completions suite is selected without
	// an explicit completion model. Legacy /v1/completions expects instruct models.
	DefaultCompletionModel = "gpt-3.5-turbo-instruct"
)

// DefaultSuites are run when TEST_SUITES is unset or set to "all".
var DefaultSuites = []string{
	"models",
	"chat_completions",
	"chat_completions_stream",
	"embeddings",
}

var knownSuites = map[string]struct{}{
	"models":                   {},
	"chat_completions":         {},
	"chat_completions_stream":  {},
	"completions":              {},
	"embeddings":               {},
	"responses":                {},
	"responses_stream":         {},
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
	fs := flag.NewFlagSet("openai-compatibility-tester", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	baseURL := fs.String("base-url", envOrDefault(EnvBaseURL, ""), "OpenAI-compatible API base URL")
	apiKey := fs.String("api-key", envOrDefault(EnvAPIKey, ""), "API key for the endpoint")
	model := fs.String("model", envOrDefault(EnvModel, "gpt-4o-mini"), "Model for chat and responses suites")
	completionModel := fs.String("completion-model", envOrDefault(EnvCompletionModel, ""), "Model for legacy completions suite (defaults to --model)")
	embeddingModel := fs.String("embedding-model", envOrDefault(EnvEmbeddingModel, "text-embedding-3-small"), "Model for embedding tests")
	suiteList := fs.String("suites", envOrDefault(EnvTestSuites, "all"), "Comma-separated suite names to run, or 'all'")
	timeout := fs.Duration("timeout", 2*time.Minute, "Per-request timeout")
	listSuites := fs.Bool("list-suites", false, "List available test suites and exit")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if len(fs.Args()) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	cfg := &Config{
		BaseURL:         strings.TrimRight(strings.TrimSpace(*baseURL), "/"),
		APIKey:          strings.TrimSpace(*apiKey),
		Model:           strings.TrimSpace(*model),
		CompletionModel: strings.TrimSpace(*completionModel),
		EmbeddingModel:  strings.TrimSpace(*embeddingModel),
		RequestTimeout:  *timeout,
		ListSuites:      *listSuites,
	}

	if cfg.ListSuites {
		return cfg, nil
	}

	if *suiteList == "all" {
		cfg.Suites = append([]string(nil), DefaultSuites...)
	} else {
		seen := make(map[string]struct{})
		for _, name := range strings.Split(*suiteList, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				return nil, fmt.Errorf("duplicate test suite %q", name)
			}
			seen[name] = struct{}{}
			cfg.Suites = append(cfg.Suites, name)
		}
	}

	if explicit, empty := completionModelFlagExplicit(args); explicit && empty && suiteNeedsCompletion(cfg.Suites) {
		return nil, fmt.Errorf("%s or --completion-model is required for selected suites", EnvCompletionModel)
	}
	if cfg.CompletionModel == "" {
		if suiteNeedsCompletion(cfg.Suites) {
			cfg.CompletionModel = DefaultCompletionModel
		} else {
			cfg.CompletionModel = cfg.Model
		}
	}

	if !timeoutFlagExplicit(args) {
		envTimeout, err := envDurationOrDefault(EnvRequestTimeout, cfg.RequestTimeout)
		if err != nil {
			return nil, err
		}
		cfg.RequestTimeout = envTimeout
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
	if err := validateSuiteNames(cfg.Suites); err != nil {
		return nil, err
	}
	if cfg.RequestTimeout <= 0 {
		return nil, fmt.Errorf("request timeout must be greater than zero")
	}
	if err := validateModelsForSuites(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func suiteNeedsCompletion(names []string) bool {
	for _, name := range names {
		if name == "completions" {
			return true
		}
	}
	return false
}

func validateSuiteNames(names []string) error {
	for _, name := range names {
		if _, ok := knownSuites[name]; !ok {
			return fmt.Errorf("unknown test suite %q (use --list-suites to see options)", name)
		}
	}
	return nil
}

func validateModelsForSuites(cfg *Config) error {
	var needsChat, needsCompletion, needsEmbedding bool
	for _, name := range cfg.Suites {
		switch name {
		case "chat_completions", "chat_completions_stream", "responses", "responses_stream":
			needsChat = true
		case "completions":
			needsCompletion = true
		case "embeddings":
			needsEmbedding = true
		}
	}
	if needsChat && cfg.Model == "" {
		return fmt.Errorf("%s or --model is required for selected suites", EnvModel)
	}
	if needsCompletion && cfg.CompletionModel == "" {
		return fmt.Errorf("%s or --completion-model is required for selected suites", EnvCompletionModel)
	}
	if needsEmbedding && cfg.EmbeddingModel == "" {
		return fmt.Errorf("%s or --embedding-model is required for selected suites", EnvEmbeddingModel)
	}
	return nil
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
	if u.Hostname() == "" {
		return fmt.Errorf("%s: URL must include a hostname", EnvBaseURL)
	}
	if port := u.Port(); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil || p < 1 || p > 65535 {
			return fmt.Errorf("%s: URL port must be between 1 and 65535", EnvBaseURL)
		}
	}
	if u.RawQuery != "" {
		return fmt.Errorf("%s: query parameters in the base URL are not supported by the OpenAI Go SDK", EnvBaseURL)
	}
	if strings.Contains(strings.ToLower(raw), "%2f") {
		return fmt.Errorf("%s: encoded path separators (%%2F) are not supported by the OpenAI Go SDK", EnvBaseURL)
	}
	return nil
}

func completionModelFlagExplicit(args []string) (explicit bool, valueEmpty bool) {
	for i, arg := range args {
		switch {
		case arg == "--completion-model", arg == "-completion-model":
			explicit = true
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				valueEmpty = true
			} else {
				valueEmpty = args[i+1] == ""
			}
		case strings.HasPrefix(arg, "--completion-model="):
			explicit = true
			valueEmpty = strings.TrimPrefix(arg, "--completion-model=") == ""
		case strings.HasPrefix(arg, "-completion-model="):
			explicit = true
			valueEmpty = strings.TrimPrefix(arg, "-completion-model=") == ""
		}
	}
	return explicit, valueEmpty
}

func timeoutFlagExplicit(args []string) bool {
	for _, arg := range args {
		switch {
		case arg == "--timeout", arg == "-timeout":
			return true
		case strings.HasPrefix(arg, "--timeout="), strings.HasPrefix(arg, "-timeout="):
			return true
		}
	}
	return false
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