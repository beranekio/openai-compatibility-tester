package suites

import (
	"bytes"
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// AudioTranslations verifies POST /v1/audio/translations via client.Audio.Translations.New.
type AudioTranslations struct{}

func (AudioTranslations) Name() string        { return "audio_translations" }
func (AudioTranslations) Description() string { return "Audio translations (POST /v1/audio/translations)" }

func (AudioTranslations) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Audio.Translations.New(ctx, openai.AudioTranslationNewParams{
		File:           bytes.NewReader(smallWAVBytes()),
		Model:          openai.AudioModel(cfg.WhisperModel),
		ResponseFormat: openai.AudioTranslationNewParamsResponseFormatJSON,
	})
	if err != nil {
		return fmt.Errorf("audio translation request failed: %w", err)
	}
	if resp == nil {
		return fail("audio_translations", "response is nil")
	}
	if resp.Text == "" {
		return fail("audio_translations", "response missing text")
	}
	return nil
}