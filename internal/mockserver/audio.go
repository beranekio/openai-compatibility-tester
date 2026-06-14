package mockserver

import (
	"encoding/json"
	"net/http"
	"strings"
)

// mockAudioBytes is a minimal non-empty audio payload for TTS mocks.
var mockAudioBytes = []byte{0xff, 0xfb, 0x90, 0x00}

func handleAudioSpeech(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "audio/mpeg")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(mockAudioBytes)
}

func handleAudioTranscriptions(w http.ResponseWriter, r *http.Request) {
	if transcriptionRequestWantsStream(r) {
		writeAudioTranscriptionStream(w)
		return
	}
	writeJSON(w, map[string]any{"text": "compatibility test"})
}

func handleAudioTranslations(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{"text": "compatibility test"})
}

func transcriptionRequestWantsStream(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(1 << 20); err == nil {
			if values, ok := r.MultipartForm.Value["stream"]; ok {
				for _, value := range values {
					if value == "true" {
						return true
					}
				}
			}
		}
	}

	return false
}

func writeAudioTranscriptionStream(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	delta, _ := json.Marshal(map[string]any{
		"type":  "transcript.text.delta",
		"delta": "compat",
	})
	_, _ = w.Write([]byte("data: " + string(delta) + "\n\n"))

	done, _ := json.Marshal(map[string]any{
		"type": "transcript.text.done",
		"text": "compatibility test",
	})
	_, _ = w.Write([]byte("data: " + string(done) + "\n\n"))
}