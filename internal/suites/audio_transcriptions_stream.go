package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// AudioTranscriptionsStream verifies streaming POST /v1/audio/transcriptions.
type AudioTranscriptionsStream struct{}

func (AudioTranscriptionsStream) Name() string { return "audio_transcriptions_stream" }
func (AudioTranscriptionsStream) Description() string {
	return "Streaming audio transcriptions (POST /v1/audio/transcriptions, stream=true)"
}

func (AudioTranscriptionsStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Audio.Transcriptions.NewStreaming(ctx, openai.AudioTranscriptionNewParams{
		File:           smallWAVReader(),
		Model:          openai.AudioModel(cfg.TranscriptionModel),
		ResponseFormat: openai.AudioResponseFormatJSON,
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("audio transcription stream failed: %w", err)
	}
	if err := validateEventStreamContentType("audio_transcriptions_stream", httpResp); err != nil {
		return err
	}

	var hasDelta bool
	var hasDone bool
	var terminalReached bool
	for stream.Next() {
		if terminalReached {
			return fail("audio_transcriptions_stream", fmt.Sprintf("stream event %q after terminal event", stream.Current().Type))
		}

		event := stream.Current()
		switch event.Type {
		case "transcript.text.delta":
			delta := event.AsTranscriptTextDelta()
			if delta.Delta == "" {
				return fail("audio_transcriptions_stream", "transcript.text.delta missing delta")
			}
			hasDelta = true
		case "transcript.text.done":
			done := event.AsTranscriptTextDone()
			if done.Text == "" {
				return fail("audio_transcriptions_stream", "transcript.text.done missing text")
			}
			hasDone = true
			terminalReached = true
		default:
			return fail("audio_transcriptions_stream", fmt.Sprintf("unexpected stream event type %q", event.Type))
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("audio transcription stream failed: %w", err)
	}
	if !hasDelta {
		return fail("audio_transcriptions_stream", "stream ended without transcript.text.delta event")
	}
	if !hasDone {
		return fail("audio_transcriptions_stream", "stream ended without transcript.text.done event")
	}
	return nil
}