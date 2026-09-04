package suites

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/realtime"
)

// RealtimeTranscriptionClientSecrets verifies POST /v1/realtime/client_secrets
// with a transcription session.
type RealtimeTranscriptionClientSecrets struct{}

func (RealtimeTranscriptionClientSecrets) Name() string {
	return "realtime_transcription_client_secrets"
}
func (RealtimeTranscriptionClientSecrets) Description() string {
	return "Realtime transcription client secret creation (POST /v1/realtime/client_secrets). WebSocket session behavior is not exercised."
}

func (RealtimeTranscriptionClientSecrets) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	created, err := client.Realtime.ClientSecrets.New(ctx, realtime.ClientSecretNewParams{
		Session: realtime.ClientSecretNewParamsSessionUnion{
			OfTranscription: &realtime.RealtimeTranscriptionSessionCreateRequestParam{
				Type: "transcription",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("realtime transcription client secret create failed: %w", err)
	}
	return validateRealtimeTranscriptionClientSecretResponse("realtime_transcription_client_secrets", created)
}

func validateRealtimeTranscriptionClientSecretResponse(suite string, resp *realtime.ClientSecretNewResponse) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if strings.TrimSpace(resp.Value) == "" {
		return fail(suite, "client secret value is empty")
	}
	if resp.ExpiresAt <= 0 {
		return fail(suite, fmt.Sprintf("expires_at is %d, want positive unix timestamp", resp.ExpiresAt))
	}
	if resp.ExpiresAt < time.Now().Unix() {
		return fail(suite, "expires_at is in the past")
	}
	session := resp.Session.AsTranscription()
	if session.ID == "" {
		return fail(suite, "session id is empty")
	}
	if session.Object != "realtime.transcription_session" {
		return fail(suite, fmt.Sprintf("session object is %q, want realtime.transcription_session", session.Object))
	}
	if string(session.Type) != "transcription" {
		return fail(suite, fmt.Sprintf("session type is %q, want transcription", session.Type))
	}
	return nil
}
