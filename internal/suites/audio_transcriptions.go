package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// AudioTranscriptions verifies POST /v1/audio/transcriptions via client.Audio.Transcriptions.New.
type AudioTranscriptions struct{}

func (AudioTranscriptions) Name() string        { return "audio_transcriptions" }
func (AudioTranscriptions) Description() string { return "Audio transcriptions (POST /v1/audio/transcriptions)" }

func (AudioTranscriptions) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:           smallWAVReader(),
		Model:          openai.AudioModel(cfg.WhisperModel),
		ResponseFormat: openai.AudioResponseFormatJSON,
	})
	if err != nil {
		return fmt.Errorf("audio transcription request failed: %w", err)
	}
	if resp == nil {
		return fail("audio_transcriptions", "response is nil")
	}
	if resp.Text == "" {
		return fail("audio_transcriptions", "response missing text")
	}
	return nil
}