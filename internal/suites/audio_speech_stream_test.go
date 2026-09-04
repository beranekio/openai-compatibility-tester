package suites

import (
	"io"
	"net/http"
	"strings"
	"testing"
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
