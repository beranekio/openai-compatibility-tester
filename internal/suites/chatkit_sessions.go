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
	if err := validateChatKitSessionCreate("chatkit_sessions", created, chatkitSessionUser, cfg.ChatKitWorkflowID); err != nil {
		return err
	}
	sessionID := created.ID

	cancelled, err := client.Beta.ChatKit.Sessions.Cancel(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("chatkit session cancel failed: %w", err)
	}
	if err := validateChatKitSessionCancelled("chatkit_sessions", cancelled, cfg.ChatKitWorkflowID); err != nil {
		return err
	}
	if cancelled.ID != sessionID {
		return fail("chatkit_sessions", fmt.Sprintf("cancel id is %q, want %q", cancelled.ID, sessionID))
	}
	return nil
}

func validateChatKitSessionCreate(suite string, session *openai.ChatSession, wantUser, wantWorkflowID string) error {
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
	if session.JSON.RateLimits.Valid() {
		if !session.RateLimits.JSON.MaxRequestsPer1Minute.Valid() {
			return fail(suite, "session rate_limits missing max_requests_per_1_minute")
		}
		if session.RateLimits.MaxRequestsPer1Minute != session.MaxRequestsPer1Minute {
			return fail(suite, fmt.Sprintf(
				"session rate_limits.max_requests_per_1_minute is %d, want %d",
				session.RateLimits.MaxRequestsPer1Minute,
				session.MaxRequestsPer1Minute,
			))
		}
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
	if session.Status != openai.ChatSessionStatusActive {
		return fail(suite, fmt.Sprintf("session status is %q, want %q", session.Status, openai.ChatSessionStatusActive))
	}
	if !session.JSON.User.Valid() {
		return fail(suite, "session missing user")
	}
	if session.User != wantUser {
		return fail(suite, fmt.Sprintf("session user is %q, want %q", session.User, wantUser))
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
	return nil
}

func validateChatKitSessionCancelled(suite string, session *openai.ChatSession, wantWorkflowID string) error {
	if session == nil {
		return fail(suite, "session is nil")
	}
	if session.ID == "" {
		return fail(suite, "session missing id")
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
	if session.Status != openai.ChatSessionStatusCancelled {
		return fail(suite, fmt.Sprintf("session status is %q, want %q", session.Status, openai.ChatSessionStatusCancelled))
	}
	if session.JSON.Workflow.Valid() {
		if !session.Workflow.JSON.ID.Valid() {
			return fail(suite, "session workflow missing id")
		}
		if session.Workflow.ID != wantWorkflowID {
			return fail(suite, fmt.Sprintf("session workflow id is %q, want %q", session.Workflow.ID, wantWorkflowID))
		}
	}
	return nil
}