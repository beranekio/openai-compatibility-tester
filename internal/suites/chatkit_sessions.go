package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/openai/openai-go/v3"
)

const chatkitSessionUser = "compatibility-test-user"

// ChatKitSessions verifies Beta ChatKit session lifecycle via client.Beta.ChatKit.Sessions.*.
type ChatKitSessions struct{}

func (ChatKitSessions) Name() string { return "chatkit_sessions" }
func (ChatKitSessions) Description() string {
	return "Beta ChatKit sessions (POST /v1/chatkit/sessions, POST /v1/chatkit/sessions/{id}/cancel)"
}

func (ChatKitSessions) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := client.Beta.ChatKit.Sessions.New(ctx, openai.BetaChatKitSessionNewParams{
		User: chatkitSessionUser,
		Workflow: openai.ChatSessionWorkflowParam{
			ID: cfg.ChatKitWorkflowID,
		},
	})
	if err != nil {
		return fmt.Errorf("chatkit session create failed: %w", err)
	}
	if err := validateChatKitSession("chatkit_sessions", created, chatkitSessionUser, cfg.ChatKitWorkflowID, openai.ChatSessionStatusActive); err != nil {
		return err
	}
	sessionID := created.ID

	cancelled, err := client.Beta.ChatKit.Sessions.Cancel(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("chatkit session cancel failed: %w", err)
	}
	if err := validateChatKitSession("chatkit_sessions", cancelled, chatkitSessionUser, cfg.ChatKitWorkflowID, openai.ChatSessionStatusCancelled); err != nil {
		return err
	}
	if cancelled.ID != sessionID {
		return fail("chatkit_sessions", fmt.Sprintf("cancel id is %q, want %q", cancelled.ID, sessionID))
	}
	return nil
}

func validateChatKitSession(suite string, session *openai.ChatSession, wantUser, wantWorkflowID string, wantStatus openai.ChatSessionStatus) error {
	if session == nil {
		return fail(suite, "session is nil")
	}
	if session.ID == "" {
		return fail(suite, "session missing id")
	}
	if !session.JSON.ClientSecret.Valid() {
		return fail(suite, "session missing client_secret")
	}
	if session.ClientSecret == "" {
		return fail(suite, "session client_secret is empty")
	}
	if !session.JSON.ExpiresAt.Valid() {
		return fail(suite, "session missing expires_at")
	}
	if !session.JSON.MaxRequestsPer1Minute.Valid() {
		return fail(suite, "session missing max_requests_per_1_minute")
	}
	if !session.JSON.Object.Valid() {
		return fail(suite, "session missing object")
	}
	if string(session.Object) != "chatkit.session" {
		return fail(suite, fmt.Sprintf("session object is %q, want chatkit.session", session.Object))
	}
	if !session.JSON.Status.Valid() {
		return fail(suite, "session missing status")
	}
	if session.Status != wantStatus {
		return fail(suite, fmt.Sprintf("session status is %q, want %q", session.Status, wantStatus))
	}
	if !session.JSON.User.Valid() {
		return fail(suite, "session missing user")
	}
	if session.User != wantUser {
		return fail(suite, fmt.Sprintf("session user is %q, want %q", session.User, wantUser))
	}
	if !session.JSON.ChatKitConfiguration.Valid() {
		return fail(suite, "session missing chatkit_configuration")
	}
	if !session.ChatKitConfiguration.JSON.AutomaticThreadTitling.Valid() {
		return fail(suite, "session missing chatkit_configuration.automatic_thread_titling")
	}
	if !session.ChatKitConfiguration.JSON.FileUpload.Valid() {
		return fail(suite, "session missing chatkit_configuration.file_upload")
	}
	if !session.ChatKitConfiguration.JSON.History.Valid() {
		return fail(suite, "session missing chatkit_configuration.history")
	}
	if !session.JSON.RateLimits.Valid() {
		return fail(suite, "session missing rate_limits")
	}
	if !session.RateLimits.JSON.MaxRequestsPer1Minute.Valid() {
		return fail(suite, "session missing rate_limits.max_requests_per_1_minute")
	}
	if !session.JSON.Workflow.Valid() {
		return fail(suite, "session missing workflow")
	}
	if !session.Workflow.JSON.ID.Valid() {
		return fail(suite, "session missing workflow.id")
	}
	if session.Workflow.ID != wantWorkflowID {
		return fail(suite, fmt.Sprintf("session workflow id is %q, want %q", session.Workflow.ID, wantWorkflowID))
	}
	if !session.Workflow.JSON.Tracing.Valid() {
		return fail(suite, "session missing workflow.tracing")
	}
	return nil
}