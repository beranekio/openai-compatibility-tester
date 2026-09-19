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

const (
	agentCreateName         = "compatibility-test-agent"
	agentUpdateName         = "compatibility-test-agent-updated"
	agentCreateInstructions = "Reply with exactly the word: pong."
	agentUpdateInstructions = "Reply with exactly the word: updated."
)

// Agents verifies Beta Agents API CRUD via client.Beta.Agents.*.
type Agents struct{}

func (Agents) Name() string { return "agents" }
func (Agents) Description() string {
	return "Agents API CRUD (POST/GET /v1/agents, GET/POST/DELETE /v1/agents/{id})"
}

func (Agents) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	deleted := false
	var agentID string
	defer func() {
		if agentID != "" && !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Beta.Agents.Delete(cleanupCtx, agentID)
		}
	}()

	created, err := client.Beta.Agents.New(ctx, openai.BetaAgentNewParams{
		Model:        cfg.Model,
		Name:         openai.String(agentCreateName),
		Instructions: openai.String(agentCreateInstructions),
		Metadata: map[string]string{
			"suite": "agents",
		},
	})
	if err != nil {
		return fmt.Errorf("agent create failed: %w", err)
	}
	if created != nil && created.ID != "" {
		agentID = created.ID
	}
	if err := validateAgentObject("agents", created); err != nil {
		return err
	}
	if created.Name != agentCreateName {
		return fail("agents", fmt.Sprintf("create name is %q, want %q", created.Name, agentCreateName))
	}
	if created.Instructions != agentCreateInstructions {
		return fail("agents", fmt.Sprintf("create instructions is %q, want %q", created.Instructions, agentCreateInstructions))
	}
	if err := validateAgentMetadata("agents", created.Metadata, map[string]string{"suite": "agents"}); err != nil {
		return err
	}

	got, err := client.Beta.Agents.Get(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent get failed: %w", err)
	}
	if err := validateAgentObject("agents", got); err != nil {
		return err
	}
	if got.ID != agentID {
		return fail("agents", fmt.Sprintf("get id is %q, want %q", got.ID, agentID))
	}
	if got.Name != agentCreateName {
		return fail("agents", fmt.Sprintf("get name is %q, want %q", got.Name, agentCreateName))
	}

	updated, err := client.Beta.Agents.Update(ctx, agentID, openai.BetaAgentUpdateParams{
		Name:         openai.String(agentUpdateName),
		Instructions: openai.String(agentUpdateInstructions),
		Metadata: map[string]string{
			"suite":  "agents",
			"status": "updated",
		},
	})
	if err != nil {
		return fmt.Errorf("agent update failed: %w", err)
	}
	if err := validateAgentObject("agents", updated); err != nil {
		return err
	}
	if updated.ID != agentID {
		return fail("agents", fmt.Sprintf("update id is %q, want %q", updated.ID, agentID))
	}
	if updated.Name != agentUpdateName {
		return fail("agents", fmt.Sprintf("update name is %q, want %q", updated.Name, agentUpdateName))
	}
	if updated.Instructions != agentUpdateInstructions {
		return fail("agents", fmt.Sprintf("update instructions is %q, want %q", updated.Instructions, agentUpdateInstructions))
	}
	if err := validateAgentMetadata("agents", updated.Metadata, map[string]string{
		"suite":  "agents",
		"status": "updated",
	}); err != nil {
		return err
	}

	listPage, err := client.Beta.Agents.List(ctx, openai.BetaAgentListParams{
		Limit: openai.Int(100),
	})
	if err != nil {
		return fmt.Errorf("agent list failed: %w", err)
	}
	found, err := agentListContains(listPage, agentID)
	if err != nil {
		return err
	}
	if !found {
		return fail("agents", "created agent missing from list response")
	}

	deletedResp, err := client.Beta.Agents.Delete(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent delete failed: %w", err)
	}
	if err := validateAgentDeleted("agents", deletedResp); err != nil {
		return err
	}
	if deletedResp.ID != agentID {
		return fail("agents", fmt.Sprintf("delete id is %q, want %q", deletedResp.ID, agentID))
	}

	_, getErr := client.Beta.Agents.Get(ctx, agentID)
	if getErr == nil {
		return fail("agents", "agent get after delete succeeded; agent still exists")
	}
	var apiError *openai.Error
	if !errors.As(getErr, &apiError) {
		return fmt.Errorf("agent get after delete failed: %w", getErr)
	}
	if apiError.StatusCode != http.StatusNotFound {
		return fail("agents", fmt.Sprintf("agent get after delete returned status %d, want 404", apiError.StatusCode))
	}
	deleted = true
	return nil
}

func validateAgentObject(suite string, agent *openai.Agent) error {
	if agent == nil {
		return fail(suite, "agent is nil")
	}
	if agent.ID == "" {
		return fail(suite, "agent missing id")
	}
	if !agent.JSON.CreatedAt.Valid() {
		return fail(suite, "agent missing created_at")
	}
	if !agent.JSON.UpdatedAt.Valid() {
		return fail(suite, "agent missing updated_at")
	}
	if !agent.JSON.Metadata.Valid() {
		return fail(suite, "agent missing metadata")
	}
	if !agent.JSON.Model.Valid() {
		return fail(suite, "agent missing model")
	}
	if agent.Model == "" {
		return fail(suite, "agent model is empty")
	}
	if !agent.JSON.Name.Valid() {
		return fail(suite, "agent missing name")
	}
	if !agent.JSON.Object.Valid() {
		return fail(suite, "agent missing object")
	}
	if string(agent.Object) != "agent" {
		return fail(suite, fmt.Sprintf("agent object is %q, want agent", agent.Object))
	}
	if !agent.JSON.Tools.Valid() {
		return fail(suite, "agent missing tools")
	}
	return nil
}

func validateAgentMetadata(suite string, metadata map[string]string, want map[string]string) error {
	for key, value := range want {
		got, ok := metadata[key]
		if !ok {
			return fail(suite, fmt.Sprintf("metadata missing key %q", key))
		}
		if got != value {
			return fail(suite, fmt.Sprintf("metadata[%q] is %q, want %q", key, got, value))
		}
	}
	return nil
}

func validateAgentDeleted(suite string, deleted *openai.AgentDeleted) error {
	if deleted == nil {
		return fail(suite, "delete response is nil")
	}
	if deleted.ID == "" {
		return fail(suite, "delete response missing id")
	}
	if !deleted.JSON.Deleted.Valid() {
		return fail(suite, "delete response missing deleted")
	}
	if !deleted.Deleted {
		return fail(suite, "delete response deleted is false")
	}
	if !deleted.JSON.Object.Valid() {
		return fail(suite, "delete response missing object")
	}
	if string(deleted.Object) != "agent.deleted" {
		return fail(suite, fmt.Sprintf("delete object is %q, want agent.deleted", deleted.Object))
	}
	return nil
}

func agentListContains(page *pagination.CursorPage[openai.Agent], agentID string) (bool, error) {
	for page != nil {
		if err := validateCursorListPage("agents", page, func(a *openai.Agent) string { return a.ID }); err != nil {
			return false, err
		}
		for i := range page.Data {
			if err := validateAgentObject("agents", &page.Data[i]); err != nil {
				return false, err
			}
			if page.Data[i].ID == agentID {
				return true, nil
			}
		}
		next, err := page.GetNextPage()
		if err != nil {
			return false, fmt.Errorf("agent list next page failed: %w", err)
		}
		page = next
	}
	return false, nil
}
