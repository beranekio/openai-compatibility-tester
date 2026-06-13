package config

import (
	"flag"
	"fmt"
	"net"
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
	EnvResponsesModel  = "OPENAI_RESPONSES_MODEL"
	EnvVisionModel     = "OPENAI_VISION_MODEL"
	EnvImageModel      = "OPENAI_IMAGE_MODEL"
	EnvTTSModel        = "OPENAI_TTS_MODEL"
	EnvWhisperModel    = "OPENAI_WHISPER_MODEL"
	EnvTestSuites      = "TEST_SUITES"
	EnvRequestTimeout  = "REQUEST_TIMEOUT"
	EnvAllowInsecureHTTP = "ALLOW_INSECURE_HTTP"

	// DefaultCompletionModel is used when the completions suite is selected without
	// an explicit completion model. Legacy /v1/completions expects instruct models.
	DefaultCompletionModel = "gpt-3.5-turbo-instruct"
)

// DefaultSuites are run when TEST_SUITES is unset or set to "all" or "default".
var DefaultSuites = []string{
	"models",
	"chat_completions",
	"chat_completions_stream",
	"responses",
	"responses_stream",
}

// ExtendedSuites adds commonly optional inference suites to the default set.
// Update this list when new opt-in suites ship (see issue #45).
var ExtendedSuites = []string{
	"models",
	"chat_completions",
	"chat_completions_stream",
	"responses",
	"responses_stream",
	"completions",
	"embeddings",
}

// FullSuites lists every registered suite name. Keep in sync with suites.All().
var FullSuites = []string{
	"models",
	"chat_completions",
	"chat_completions_stream",
	"completions",
	"embeddings",
	"responses",
	"responses_stream",
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
	ResponsesModel  string
	VisionModel     string
	ImageModel      string
	TTSModel        string
	WhisperModel    string
	Suites          []string
	RequestTimeout  time.Duration
	AllowInsecureHTTP bool
	ListSuites      bool
}

// Load parses configuration from environment variables and command-line flags.
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("openai-compatibility-tester", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	baseURL := fs.String("base-url", envOrDefault(EnvBaseURL, ""), "OpenAI-compatible API base URL")
	apiKey := fs.String("api-key", "", "API key for the endpoint (or set "+EnvAPIKey+")")
	model := fs.String("model", envOrDefault(EnvModel, "gpt-4o-mini"), "Model for chat completion suites")
	completionModel := fs.String("completion-model", envOrDefault(EnvCompletionModel, ""), "Model for legacy completions suite (defaults to "+DefaultCompletionModel+" when completions is selected)")
	embeddingModel := fs.String("embedding-model", envOrDefault(EnvEmbeddingModel, ""), "Model for embedding tests (required when embeddings suite is selected)")
	responsesModel := fs.String("responses-model", envOrDefault(EnvResponsesModel, ""), "Model for Responses API suites (defaults to --model)")
	visionModel := fs.String("vision-model", envOrDefault(EnvVisionModel, ""), "Model for vision chat suites (defaults to --model)")
	imageModel := fs.String("image-model", envOrDefault(EnvImageModel, ""), "Model for image generation suites")
	ttsModel := fs.String("tts-model", envOrDefault(EnvTTSModel, ""), "Model for text-to-speech suites")
	whisperModel := fs.String("whisper-model", envOrDefault(EnvWhisperModel, ""), "Model for speech-to-text suites")
	allowInsecureHTTP := fs.Bool("allow-insecure-http", envBoolOrDefault(EnvAllowInsecureHTTP, false), "Allow plaintext HTTP to non-loopback hosts")
	suiteList := fs.String("suites", envOrDefault(EnvTestSuites, "all"), "Comma-separated suite names, or preset: all, default, extended, full")
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
		EmbeddingModel:    strings.TrimSpace(*embeddingModel),
		ResponsesModel:    strings.TrimSpace(*responsesModel),
		VisionModel:       strings.TrimSpace(*visionModel),
		ImageModel:        strings.TrimSpace(*imageModel),
		TTSModel:          strings.TrimSpace(*ttsModel),
		WhisperModel:      strings.TrimSpace(*whisperModel),
		RequestTimeout:    *timeout,
		AllowInsecureHTTP: *allowInsecureHTTP,
		ListSuites:        *listSuites,
	}

	if cfg.ListSuites {
		return cfg, nil
	}

	if explicit, empty := apiKeyFlagExplicit(args); explicit && empty {
		return nil, fmt.Errorf("%s or --api-key is required", EnvAPIKey)
	}
	if cfg.APIKey == "" {
		cfg.APIKey = envOrDefault(EnvAPIKey, "")
	}

	suites, err := resolveSuiteSelection(*suiteList)
	if err != nil {
		return nil, err
	}
	cfg.Suites = suites

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
	if cfg.ResponsesModel == "" {
		cfg.ResponsesModel = cfg.Model
	}
	if cfg.VisionModel == "" {
		cfg.VisionModel = cfg.Model
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
	if err := validateBaseURL(cfg.BaseURL, cfg.AllowInsecureHTTP); err != nil {
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

func resolveSuiteSelection(raw string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "all", "default":
		return append([]string(nil), DefaultSuites...), nil
	case "extended":
		return append([]string(nil), ExtendedSuites...), nil
	case "full":
		return append([]string(nil), FullSuites...), nil
	}

	seen := make(map[string]struct{})
	var suites []string
	for _, name := range strings.Split(raw, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("duplicate test suite %q", name)
		}
		seen[name] = struct{}{}
		suites = append(suites, name)
	}
	return suites, nil
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
	var needsChat, needsResponses, needsCompletion, needsEmbedding bool
	var needsVision, needsImage, needsTTS, needsWhisper bool
	for _, name := range cfg.Suites {
		switch name {
		case "chat_completions", "chat_completions_stream":
			needsChat = true
		case "responses", "responses_stream":
			needsResponses = true
		case "completions":
			needsCompletion = true
		case "embeddings":
			needsEmbedding = true
		case "chat_completions_vision":
			needsVision = true
		case "images_generations", "images_edits", "images_variations":
			needsImage = true
		case "audio_speech":
			needsTTS = true
		case "audio_transcriptions", "audio_transcriptions_stream", "audio_translations":
			needsWhisper = true
		}
	}
	if needsChat && cfg.Model == "" {
		return fmt.Errorf("%s or --model is required for selected suites", EnvModel)
	}
	if needsResponses && cfg.ResponsesModel == "" {
		return fmt.Errorf("%s or --responses-model is required for selected suites", EnvResponsesModel)
	}
	if needsCompletion && cfg.CompletionModel == "" {
		return fmt.Errorf("%s or --completion-model is required for selected suites", EnvCompletionModel)
	}
	if needsEmbedding && cfg.EmbeddingModel == "" {
		return fmt.Errorf("%s or --embedding-model is required for selected suites", EnvEmbeddingModel)
	}
	if needsVision && cfg.VisionModel == "" {
		return fmt.Errorf("%s or --vision-model is required for selected suites", EnvVisionModel)
	}
	if needsImage && cfg.ImageModel == "" {
		return fmt.Errorf("%s or --image-model is required for selected suites", EnvImageModel)
	}
	if needsTTS && cfg.TTSModel == "" {
		return fmt.Errorf("%s or --tts-model is required for selected suites", EnvTTSModel)
	}
	if needsWhisper && cfg.WhisperModel == "" {
		return fmt.Errorf("%s or --whisper-model is required for selected suites", EnvWhisperModel)
	}
	return nil
}

func validateBaseURL(raw string, allowInsecureHTTP bool) error {
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
	if u.Scheme == "http" && !allowInsecureHTTP && !isLoopbackHost(u.Hostname()) {
		return fmt.Errorf("%s: plaintext HTTP is only permitted for loopback hosts unless --allow-insecure-http is set", EnvBaseURL)
	}
	return nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func apiKeyFlagExplicit(args []string) (explicit bool, valueEmpty bool) {
	for i, arg := range args {
		switch {
		case arg == "--api-key", arg == "-api-key":
			explicit = true
			if i+1 >= len(args) {
				valueEmpty = true
			} else {
				valueEmpty = strings.TrimSpace(args[i+1]) == ""
			}
		case strings.HasPrefix(arg, "--api-key="):
			explicit = true
			valueEmpty = strings.TrimSpace(strings.TrimPrefix(arg, "--api-key=")) == ""
		case strings.HasPrefix(arg, "-api-key="):
			explicit = true
			valueEmpty = strings.TrimSpace(strings.TrimPrefix(arg, "-api-key=")) == ""
		}
	}
	return explicit, valueEmpty
}

func completionModelFlagExplicit(args []string) (explicit bool, valueEmpty bool) {
	for i, arg := range args {
		switch {
		case arg == "--completion-model", arg == "-completion-model":
			explicit = true
			if i+1 >= len(args) {
				valueEmpty = true
			} else {
				valueEmpty = strings.TrimSpace(args[i+1]) == ""
			}
		case strings.HasPrefix(arg, "--completion-model="):
			explicit = true
			valueEmpty = strings.TrimSpace(strings.TrimPrefix(arg, "--completion-model=")) == ""
		case strings.HasPrefix(arg, "-completion-model="):
			explicit = true
			valueEmpty = strings.TrimSpace(strings.TrimPrefix(arg, "-completion-model=")) == ""
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

func envBoolOrDefault(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
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