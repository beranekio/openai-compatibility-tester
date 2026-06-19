package mockserver

import (
	"net/http"
	"time"
)

func handleRealtimeClientSecretCreate(w http.ResponseWriter, _ *http.Request) {
	now := time.Now().Unix()
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