package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/openai/openai-go/v3"
)

// Suite exercises one area of the OpenAI API through the official Go SDK.
type Suite interface {
	Name() string
	Description() string
	Run(ctx context.Context, client openai.Client, cfg *config.Config) error
}

// All returns every registered compatibility suite.
func All() []Suite {
	return []Suite{
		Models{},
		ModelsGet{},
		ChatCompletions{},
		ChatCompletionsStream{},
		Completions{},
		Embeddings{},
		Responses{},
		ResponsesStream{},
	}
}

// ByName maps suite names to implementations.
func ByName() map[string]Suite {
	registry := make(map[string]Suite, len(All()))
	for _, suite := range All() {
		registry[suite.Name()] = suite
	}
	return registry
}

// ModelRequirements describes which model settings selected suites need.
type ModelRequirements struct {
	Chat       bool
	Completion bool
	Embedding  bool
	Vision     bool
	Image      bool
	TTS        bool
	Whisper    bool
}

// RequiredModels returns which model settings must be configured for the suites.
func RequiredModels(names []string) ModelRequirements {
	var req ModelRequirements
	for _, name := range names {
		switch name {
		case "chat_completions", "chat_completions_stream", "models_get", "responses", "responses_stream":
			req.Chat = true
		case "completions":
			req.Completion = true
		case "embeddings":
			req.Embedding = true
		case "chat_completions_vision":
			req.Vision = true
		case "images_generations", "images_edits", "images_variations":
			req.Image = true
		case "audio_speech":
			req.TTS = true
		case "audio_transcriptions", "audio_transcriptions_stream", "audio_translations":
			req.Whisper = true
		}
	}
	return req
}

// ValidateNames reports whether every name is a registered suite.
func ValidateNames(names []string) error {
	registry := ByName()
	for _, name := range names {
		if _, ok := registry[name]; !ok {
			return fmt.Errorf("unknown test suite %q (use --list-suites to see options)", name)
		}
	}
	return nil
}

// Names returns sorted suite names for display.
func Names() []string {
	all := All()
	names := make([]string, len(all))
	for i, suite := range all {
		names[i] = suite.Name()
	}
	return names
}