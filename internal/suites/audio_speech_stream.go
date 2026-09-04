package suites

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// AudioSpeechStream verifies streaming POST /v1/audio/speech.
type AudioSpeechStream struct{}

func (AudioSpeechStream) Name() string { return "audio_speech_stream" }
func (AudioSpeechStream) Description() string {
	return "Streaming text-to-speech (POST /v1/audio/speech)"
}

func (AudioSpeechStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	streamFormat := speechStreamFormat(cfg.TTSModel)
	opts := []option.RequestOption{option.WithResponseInto(&httpResp)}
	if streamFormat == openai.AudioSpeechNewParamsStreamFormatSSE {
		opts = append([]option.RequestOption{option.WithHeader("Accept", "text/event-stream")}, opts...)
	}
	_, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Model: openai.SpeechModel(cfg.TTSModel),
		Input: "compatibility test",
		Voice: openai.AudioSpeechNewParamsVoiceUnion{
			OfAudioSpeechNewsVoiceString2: openai.String("alloy"),
		},
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
		StreamFormat:   streamFormat,
	}, opts...)
	if err != nil {
		return fmt.Errorf("audio speech stream request failed: %w", err)
	}
	if httpResp == nil {
		return fail("audio_speech_stream", "response is nil")
	}

	mediaType, _, parseErr := mime.ParseMediaType(httpResp.Header.Get("Content-Type"))
	if parseErr != nil {
		mediaType = strings.TrimSpace(httpResp.Header.Get("Content-Type"))
	}
	if mediaType == "text/event-stream" {
		return consumeSpeechAudioStream("audio_speech_stream", httpResp)
	}
	return validateBinaryHTTPResponse("audio_speech_stream", httpResp, 1)
}

// speechStreamFormat selects stream_format for POST /v1/audio/speech.
// The API does not support sse for tts-1 / tts-1-hd.
func speechStreamFormat(model string) openai.AudioSpeechNewParamsStreamFormat {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "tts-1", "tts-1-hd":
		return openai.AudioSpeechNewParamsStreamFormatAudio
	default:
		return openai.AudioSpeechNewParamsStreamFormatSSE
	}
}

func consumeSpeechAudioStream(suite string, resp *http.Response) error {
	if err := validateEventStreamContentType(suite, resp); err != nil {
		return err
	}
	if resp.Body == nil {
		return fail(suite, "stream response body is nil")
	}
	defer resp.Body.Close()

	var terminalReached, sawDelta bool
	scanner := bufio.NewScanner(resp.Body)
	// Real TTS SSE data lines carry base64 audio and exceed the default 64KiB token cap.
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var ev struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			return fail(suite, fmt.Sprintf("stream event is not valid JSON: %v", err))
		}
		if terminalReached {
			return fail(suite, fmt.Sprintf("stream event %q after terminal speech.audio.done event", ev.Type))
		}
		switch ev.Type {
		case "speech.audio.delta":
			sawDelta = true
		case "speech.audio.done":
			terminalReached = true
		default:
			return fail(suite, fmt.Sprintf("unexpected stream event type %q", ev.Type))
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s stream failed: %w", suite, err)
	}
	if !sawDelta {
		return fail(suite, "stream ended without speech.audio.delta event")
	}
	if !terminalReached {
		return fail(suite, "stream ended without speech.audio.done event")
	}
	return nil
}
