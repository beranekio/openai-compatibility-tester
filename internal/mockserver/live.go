package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const liveMockSDPAnswer = "v=0\r\n" +
	"o=- 0 0 IN IP4 127.0.0.1\r\n" +
	"s=-\r\n" +
	"t=0 0\r\n" +
	"m=audio 9 UDP/TLS/RTP/SAVPF 111\r\n" +
	"c=IN IP4 0.0.0.0\r\n" +
	"a=mid:0\r\n"

func handleLiveSessionCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Session struct {
			Model string `json:"model"`
		} `json:"session"`
		Transport struct {
			Sdp  string `json:"sdp"`
			Type string `json:"type"`
		} `json:"transport"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Transport.Sdp == "" {
		http.Error(w, `{"error":{"message":"missing transport.sdp"}}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{
		"session": map[string]any{
			"id": "livesess_mock",
		},
		"transport": map[string]any{
			"type": "webrtc",
			"sdp":  liveMockSDPAnswer,
		},
	})
}

func handleLiveSessionHangup(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
