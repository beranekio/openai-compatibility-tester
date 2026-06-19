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
		ChatCompletionsStreamUsage{},
		ChatCompletionsLogprobs{},
		ChatCompletionsJSON{},
		ChatCompletionsVision{},
		ChatCompletionsReasoning{},
		ChatCompletionsAudio{},
		ChatCompletionsTools{},
		ChatCompletionsToolsStream{},
		ChatCompletionsMultiTurn{},
		ChatCompletionsGet{},
		ChatCompletionsList{},
		ChatCompletionsDelete{},
		ChatCompletionsMessages{},
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
		AudioSpeech{},
		AudioTranscriptions{},
		AudioTranscriptionsStream{},
		AudioTranslations{},
		Files{},
		Uploads{},
		BatchesCreate{},
		BatchesGet{},
		BatchesCancel{},
		Conversations{},
		VectorStores{},
		VectorStoreFiles{},
		VectorStoreFileBatches{},
		RealtimeClientSecrets{},
		Containers{},
		ContainerFiles{},
		Videos{},
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
	Chat          bool
	Completion    bool
	Embedding     bool
	Vision        bool
	Reasoning     bool
	Image         bool
	TTS           bool
	Whisper       bool
	Transcription bool
	Realtime      bool
	Video         bool
}

// RequiredModels returns which model settings must be configured for the suites.
func RequiredModels(names []string) ModelRequirements {
	var req ModelRequirements
	for _, name := range names {
		switch name {
		case "chat_completions", "chat_completions_stream", "chat_completions_stream_usage", "chat_completions_logprobs", "chat_completions_json", "chat_completions_audio", "chat_completions_tools", "chat_completions_tools_stream", "chat_completions_multi_turn", "chat_completions_get", "chat_completions_list", "chat_completions_delete", "chat_completions_messages", "models_get", "responses", "responses_stream", "responses_tools", "responses_tools_stream", "responses_json", "responses_get", "responses_delete", "responses_cancel", "responses_input_items", "responses_compact", "responses_input_tokens", "batches_create", "batches_get", "batches_cancel":
			req.Chat = true
		case "completions", "completions_stream":
			req.Completion = true
		case "embeddings", "embeddings_batch":
			req.Embedding = true
		case "chat_completions_vision":
			req.Vision = true
		case "chat_completions_reasoning":
			req.Reasoning = true
		case "images_generations", "images_edits":
			req.Image = true
		case "audio_speech":
			req.TTS = true
		case "audio_transcriptions", "audio_translations":
			req.Whisper = true
		case "audio_transcriptions_stream":
			req.Transcription = true
		case "realtime_client_secrets":
			req.Realtime = true
		case "videos":
			req.Video = true
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
