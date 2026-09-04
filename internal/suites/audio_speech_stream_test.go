package suites

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestConsumeSpeechAudioStreamAcceptsLargeDelta(t *testing.T) {
	payload := `{"type":"speech.audio.delta","audio":"` + strings.Repeat("A", 70*1024) + `"}`
	body := "data: " + payload + "\n\ndata: {\"type\":\"speech.audio.done\"}\n\n"
	resp := &http.Response{
		Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:   io.NopCloser(strings.NewReader(body)),
	}
	if err := consumeSpeechAudioStream("audio_speech_stream", resp); err != nil {
		t.Fatalf("consumeSpeechAudioStream() error = %v", err)
	}
}

func TestConsumeSpeechAudioStreamRequiresDelta(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:   io.NopCloser(strings.NewReader("data: {\"type\":\"speech.audio.done\"}\n\n")),
	}
	err := consumeSpeechAudioStream("audio_speech_stream", resp)
	if err == nil || !strings.Contains(err.Error(), "speech.audio.delta") {
		t.Fatalf("consumeSpeechAudioStream() error = %v, want missing delta", err)
	}
}

func TestSpeechStreamFormatUsesAudioForLegacyTTS(t *testing.T) {
	tests := []struct {
		model string
		want  openai.AudioSpeechNewParamsStreamFormat
	}{
		{model: "tts-1", want: openai.AudioSpeechNewParamsStreamFormatAudio},
		{model: "tts-1-hd", want: openai.AudioSpeechNewParamsStreamFormatAudio},
		{model: " TTS-1 ", want: openai.AudioSpeechNewParamsStreamFormatAudio},
		{model: "gpt-4o-mini-tts", want: openai.AudioSpeechNewParamsStreamFormatSSE},
		{model: "gpt-4o-mini-tts-2025-12-15", want: openai.AudioSpeechNewParamsStreamFormatSSE},
	}
	for _, tt := range tests {
		if got := speechStreamFormat(tt.model); got != tt.want {
			t.Errorf("speechStreamFormat(%q) = %q, want %q", tt.model, got, tt.want)
		}
	}
}
