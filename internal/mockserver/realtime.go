package mockserver

import (
	"encoding/json"
	"net/http"
	"time"
)

func handleRealtimeClientSecretCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Session struct {
			Type string `json:"type"`
		} `json:"session"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	now := time.Now().Unix()
	if req.Session.Type == "transcription" {
		writeJSON(w, map[string]any{
			"expires_at": now + 600,
			"value":      "ek_mock_transcription_client_secret",
			"session": map[string]any{
				"id":     "sess_mock_transcription",
				"object": "realtime.transcription_session",
				"type":   "transcription",
			},
		})
		return
	}
	writeJSON(w, map[string]any{
		"expires_at": now + 600,
		"value":      "ek_mock_realtime_client_secret",
		"session": map[string]any{
			"id":     "sess_mock_realtime",
			"object": "realtime.session",
			"type":   "realtime",
			"model":  "gpt-realtime",
		},
	})
}
