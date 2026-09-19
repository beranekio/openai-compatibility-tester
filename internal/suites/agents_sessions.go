package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

// AgentsSessions verifies Agents session create/get/list/delete with environment type none.
type AgentsSessions struct{}

func (AgentsSessions) Name() string { return "agents_sessions" }
func (AgentsSessions) Description() string {
	return "Agents sessions (POST/GET /v1/agents/sessions, GET/DELETE /v1/agents/sessions/{id})"
}

func (AgentsSessions) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var agentID string
	var sessionID string
	sessionDeleted := false
	agentDeleted := false
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if sessionID != "" && !sessionDeleted {
			_, _ = client.Beta.Agents.Sessions.Delete(cleanupCtx, sessionID)
		}
		if agentID != "" && !agentDeleted {
			_, _ = client.Beta.Agents.Delete(cleanupCtx, agentID)
		}
	}()

	agent, err := client.Beta.Agents.New(ctx, openai.BetaAgentNewParams{
		Model:        cfg.Model,
		Name:         openai.String(agentCreateName),
		Instructions: openai.String(agentCreateInstructions),
		Metadata: map[string]string{
			"suite": "agents_sessions",
		},
	})
	if err != nil {
		return fmt.Errorf("agent create failed: %w", err)
	}
	if err := validateAgentObject("agents_sessions", agent); err != nil {
		return err
	}
	agentID = agent.ID

	created, err := client.Beta.Agents.Sessions.New(ctx, openai.BetaAgentSessionNewParams{
		AgentID: openai.String(agentID),
		Environment: openai.EnvironmentParamUnion{
			OfParamNone: &openai.EnvironmentParamNone{},
		},
		Metadata: map[string]string{
			"suite": "agents_sessions",
		},
	})
	if err != nil {
		return fmt.Errorf("agent session create failed: %w", err)
	}
	if err := validateAgentSessionObject("agents_sessions", created); err != nil {
		return err
	}
	sessionID = created.ID
	if created.Agent.ID != "" && created.Agent.ID != agentID {
		return fail("agents_sessions", fmt.Sprintf("session agent id is %q, want %q", created.Agent.ID, agentID))
	}

	got, err := client.Beta.Agents.Sessions.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("agent session get failed: %w", err)
	}
	if err := validateAgentSessionObject("agents_sessions", got); err != nil {
		return err
	}
	if got.ID != sessionID {
		return fail("agents_sessions", fmt.Sprintf("get id is %q, want %q", got.ID, sessionID))
	}

	listPage, err := client.Beta.Agents.Sessions.List(ctx, openai.BetaAgentSessionListParams{
		Limit: openai.Int(100),
	})
	if err != nil {
		return fmt.Errorf("agent session list failed: %w", err)
	}
	found, err := agentSessionListContains(listPage, sessionID)
	if err != nil {
		return err
	}
	if !found {
		return fail("agents_sessions", "created session missing from list response")
	}

	deletedResp, err := client.Beta.Agents.Sessions.Delete(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("agent session delete failed: %w", err)
	}
	if deletedResp == nil || deletedResp.ID == "" {
		return fail("agents_sessions", "delete response missing id")
	}
	if !deletedResp.JSON.Deleted.Valid() || !deletedResp.Deleted {
		return fail("agents_sessions", "delete response deleted is false")
	}
	if string(deletedResp.Object) != "agent.session.deleted" {
		return fail("agents_sessions", fmt.Sprintf("delete object is %q, want agent.session.deleted", deletedResp.Object))
	}
	if deletedResp.ID != sessionID {
		return fail("agents_sessions", fmt.Sprintf("delete id is %q, want %q", deletedResp.ID, sessionID))
	}

	_, getErr := client.Beta.Agents.Sessions.Get(ctx, sessionID)
	if getErr == nil {
		return fail("agents_sessions", "session get after delete succeeded; session still exists")
	}
	var apiError *openai.Error
	if !errors.As(getErr, &apiError) {
		return fmt.Errorf("agent session get after delete failed: %w", getErr)
	}
	if apiError.StatusCode != http.StatusNotFound {
		return fail("agents_sessions", fmt.Sprintf("session get after delete returned status %d, want 404", apiError.StatusCode))
	}
	sessionDeleted = true

	if _, err := client.Beta.Agents.Delete(ctx, agentID); err != nil {
		return fmt.Errorf("agent delete failed: %w", err)
	}
	agentDeleted = true
	return nil
}

func validateAgentSessionObject(suite string, session *openai.AgentSession) error {
	if session == nil {
		return fail(suite, "session is nil")
	}
	if session.ID == "" {
		return fail(suite, "session missing id")
	}
	if !session.JSON.CreatedAt.Valid() {
		return fail(suite, "session missing created_at")
	}
	if !session.JSON.Object.Valid() {
		return fail(suite, "session missing object")
	}
	if string(session.Object) != "agent.session" {
		return fail(suite, fmt.Sprintf("session object is %q, want agent.session", session.Object))
	}
	if !session.JSON.Status.Valid() {
		return fail(suite, "session missing status")
	}
	switch session.Status {
	case openai.AgentSessionStatusIdle, openai.AgentSessionStatusInProgress, openai.AgentSessionStatusRequiresAction:
	default:
		return fail(suite, fmt.Sprintf("session status is %q, want idle, in_progress, or requires_action", session.Status))
	}
	if !session.JSON.Environment.Valid() {
		return fail(suite, "session missing environment")
	}
	return nil
}

func agentSessionListContains(page *pagination.CursorPage[openai.AgentSession], sessionID string) (bool, error) {
	for page != nil {
		if err := validateCursorListPage("agents_sessions", page, func(s *openai.AgentSession) string { return s.ID }); err != nil {
			return false, err
		}
		for i := range page.Data {
			if err := validateAgentSessionObject("agents_sessions", &page.Data[i]); err != nil {
				return false, err
			}
			if page.Data[i].ID == sessionID {
				return true, nil
			}
		}
		next, err := page.GetNextPage()
		if err != nil {
			return false, fmt.Errorf("agent session list next page failed: %w", err)
		}
		page = next
	}
	return false, nil
}
