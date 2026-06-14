package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// AudioSpeech verifies POST /v1/audio/speech via client.Audio.Speech.New.
type AudioSpeech struct{}

func (AudioSpeech) Name() string        { return "audio_speech" }
func (AudioSpeech) Description() string { return "Text-to-speech (POST /v1/audio/speech)" }

func (AudioSpeech) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	_, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Model: openai.SpeechModel(cfg.TTSModel),
		Input: "compatibility test",
		Voice: openai.AudioSpeechNewParamsVoiceUnion{
			OfAudioSpeechNewsVoiceString2: openai.String("alloy"),
		},
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
	}, option.WithResponseInto(&httpResp))
	if err != nil {
		return fmt.Errorf("audio speech request failed: %w", err)
	}
	return validateBinaryHTTPResponse("audio_speech", httpResp, 1)
}