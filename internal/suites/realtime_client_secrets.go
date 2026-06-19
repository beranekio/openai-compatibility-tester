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

// RealtimeClientSecrets verifies POST /v1/realtime/client_secrets via client.Realtime.ClientSecrets.New.
type RealtimeClientSecrets struct{}

func (RealtimeClientSecrets) Name() string { return "realtime_client_secrets" }
func (RealtimeClientSecrets) Description() string {
	return "Realtime API client secret creation (POST /v1/realtime/client_secrets). WebSocket session behavior is not exercised."
}

func (RealtimeClientSecrets) Run(ctx context.Context, client openai.Client, _ *config.Config) error {
	created, err := client.Realtime.ClientSecrets.New(ctx, realtime.ClientSecretNewParams{
		Session: realtime.ClientSecretNewParamsSessionUnion{
			OfRealtime: &realtime.RealtimeSessionCreateRequestParam{
				Model: realtime.RealtimeSessionCreateRequestModelGPTRealtime,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("realtime client secret create failed: %w", err)
	}
	if err := validateRealtimeClientSecretResponse("realtime_client_secrets", created); err != nil {
		return err
	}
	return nil
}

func validateRealtimeClientSecretResponse(suite string, resp *realtime.ClientSecretNewResponse) error {
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
	session := resp.Session.AsRealtime()
	if session.ID == "" {
		return fail(suite, "session id is empty")
	}
	if string(session.Object) != "realtime.session" {
		return fail(suite, fmt.Sprintf("session object is %q, want realtime.session", session.Object))
	}
	if string(session.Type) != "realtime" {
		return fail(suite, fmt.Sprintf("session type is %q, want realtime", session.Type))
	}
	return nil
}