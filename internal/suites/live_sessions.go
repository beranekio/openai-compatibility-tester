package suites

import (
	"context"
	"fmt"
	"strings"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/live"
)

// Minimal SDP offer used only to exercise POST /v1/live/sessions request parsing.
// Real endpoints may reject it; the mock server accepts it and returns an SDP answer.
const liveSessionSDPOffer = "v=0\r\n" +
	"o=- 0 0 IN IP4 127.0.0.1\r\n" +
	"s=-\r\n" +
	"t=0 0\r\n" +
	"a=group:BUNDLE 0\r\n" +
	"m=audio 9 UDP/TLS/RTP/SAVPF 111\r\n" +
	"c=IN IP4 0.0.0.0\r\n" +
	"a=rtcp:9 IN IP4 0.0.0.0\r\n" +
	"a=ice-ufrag:mock\r\n" +
	"a=ice-pwd:mockpasswordmockpassword\r\n" +
	"a=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\r\n" +
	"a=setup:actpass\r\n" +
	"a=mid:0\r\n" +
	"a=sendrecv\r\n" +
	"a=rtcp-mux\r\n"

// LiveSessions verifies POST /v1/live/sessions via client.Live.New.
type LiveSessions struct{}

func (LiveSessions) Name() string { return "live_sessions" }
func (LiveSessions) Description() string {
	return "Live API session create (POST /v1/live/sessions). WebRTC/SIP media is not exercised."
}

// Run creates a WebRTC Live session. There is no REST delete for that transport:
// Sessions.Hangup ends a SIP call (POST /v1/live/sessions/{id}/hangup), not this path.
func (LiveSessions) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := client.Live.New(ctx, live.LiveNewParams{
		Session: live.MediaSessionConfigParam{
			Model: live.MediaSessionConfigModel(cfg.LiveModel),
		},
		Transport: live.LiveNewParamsTransport{
			Sdp: liveSessionSDPOffer,
		},
	})
	if err != nil {
		return fmt.Errorf("live session create failed: %w", err)
	}
	if created == nil {
		return fail("live_sessions", "response is nil")
	}
	if !created.JSON.Session.Valid() {
		return fail("live_sessions", "response missing session")
	}
	if strings.TrimSpace(created.Session.ID) == "" {
		return fail("live_sessions", "session id is empty")
	}
	if !created.JSON.Transport.Valid() {
		return fail("live_sessions", "response missing transport")
	}
	if string(created.Transport.Type) != "webrtc" {
		return fail("live_sessions", fmt.Sprintf("transport type is %q, want webrtc", created.Transport.Type))
	}
	if strings.TrimSpace(created.Transport.Sdp) == "" {
		return fail("live_sessions", "transport sdp is empty")
	}
	return nil
}
