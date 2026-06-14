package mockserver

import "net/http"

// mockAudioBytes is a minimal non-empty audio payload for TTS mocks.
var mockAudioBytes = []byte{0xff, 0xfb, 0x90, 0x00}

func handleAudioSpeech(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "audio/mpeg")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(mockAudioBytes)
}