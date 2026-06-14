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
		ChatCompletionsJSON{},
		ChatCompletionsVision{},
		ChatCompletionsTools{},
		ChatCompletionsToolsStream{},
		Completions{},
		CompletionsStream{},
		Embeddings{},
		EmbeddingsBatch{},
		Responses{},
		ResponsesStream{},
		ResponsesTools{},
		ResponsesToolsStream{},
		ResponsesJSON{},
		ResponsesGet{},
		ResponsesDelete{},
		ResponsesCancel{},
		ResponsesInputItems{},
		ResponsesCompact{},
		ResponsesInputTokens{},
		Moderations{},
		ImagesGenerations{},
		ImagesEdits{},
		ImagesVariations{},
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
		case "chat_completions", "chat_completions_stream", "chat_completions_json", "chat_completions_tools", "chat_completions_tools_stream", "models_get", "responses", "responses_stream", "responses_tools", "responses_tools_stream", "responses_json", "responses_get", "responses_delete", "responses_cancel", "responses_input_items", "responses_compact", "responses_input_tokens":
			req.Chat = true
		case "completions", "completions_stream":
			req.Completion = true
		case "embeddings", "embeddings_batch":
			req.Embedding = true
		case "chat_completions_vision":
			req.Vision = true
		case "images_generations", "images_edits":
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