package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const agentsBetaHeaderValue = "agents=v1"

func requireAgentsBetaHeader(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(r.Header.Get("OpenAI-Beta")) != agentsBetaHeaderValue {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "missing or invalid OpenAI-Beta header",
				"type":    "invalid_request_error",
			},
		})
		return false
	}
	return true
}

func (s *Server) handleAgentCreate(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	var req struct {
		Model        string            `json:"model"`
		Name         string            `json:"name"`
		Instructions string            `json:"instructions"`
		Metadata     map[string]string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}
	agent := s.agentStore.create(req.Model, req.Name, req.Instructions, req.Metadata)
	writeJSON(w, agentPayload(agent))
}

func (s *Server) handleAgentList(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	items := s.agentStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, agent := range items {
		data[i] = agentPayload(agent)
		if i == 0 {
			firstID = agent.id
		}
		lastID = agent.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleAgentGet(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	agent, ok := s.agentStore.get(id)
	if !ok {
		writeNotFound(w, "Agent not found", "agent_id")
		return
	}
	writeJSON(w, agentPayload(agent))
}

func (s *Server) handleAgentUpdate(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	var req struct {
		Name         *string           `json:"name"`
		Instructions *string           `json:"instructions"`
		Metadata     map[string]string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	agent, ok := s.agentStore.update(id, req.Name, req.Instructions, req.Metadata)
	if !ok {
		writeNotFound(w, "Agent not found", "agent_id")
		return
	}
	writeJSON(w, agentPayload(agent))
}

func (s *Server) handleAgentDelete(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	if !s.agentStore.delete(id) {
		writeNotFound(w, "Agent not found", "agent_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "agent.deleted",
		"deleted": true,
	})
}

func (s *Server) handleAgentSessionCreate(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	var req struct {
		AgentID  string            `json:"agent_id"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	model := "gpt-4o-mini"
	name := ""
	if req.AgentID != "" {
		if agent, ok := s.agentStore.get(req.AgentID); ok {
			model = agent.model
			name = agent.name
		}
	}
	session := s.agentStore.createSession(req.AgentID, model, name, req.Metadata)
	writeJSON(w, agentSessionPayload(session))
}

func (s *Server) handleAgentSessionList(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	items := s.agentStore.listSessions()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, session := range items {
		data[i] = agentSessionPayload(session)
		if i == 0 {
			firstID = session.id
		}
		lastID = session.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleAgentSessionGet(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	session, ok := s.agentStore.getSession(id)
	if !ok {
		writeNotFound(w, "Session not found", "session_id")
		return
	}
	writeJSON(w, agentSessionPayload(session))
}

func (s *Server) handleAgentSessionDelete(w http.ResponseWriter, r *http.Request) {
	if !requireAgentsBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	if !s.agentStore.deleteSession(id) {
		writeNotFound(w, "Session not found", "session_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "agent.session.deleted",
		"deleted": true,
	})
}

func agentPayload(agent storedAgent) map[string]any {
	return map[string]any{
		"id":           agent.id,
		"object":       "agent",
		"created_at":   agent.createdAt,
		"updated_at":   agent.updatedAt,
		"model":        agent.model,
		"name":         agent.name,
		"instructions": agent.instructions,
		"metadata":     agent.metadata,
		"tools":        []any{},
		"service_tier": "auto",
		"multi_agent":  map[string]any{"enabled": false, "max_concurrent_subagents": 0},
		"reasoning":    map[string]any{"effort": "none", "summary": "auto"},
		"text":         map[string]any{"format": map[string]any{"type": "text"}, "verbosity": "medium"},
	}
}

func agentSessionPayload(session storedAgentSession) map[string]any {
	return map[string]any{
		"id":               session.id,
		"object":           "agent.session",
		"created_at":       session.createdAt,
		"last_active_at":   session.createdAt,
		"status":           "idle",
		"error":            "",
		"metadata":         session.metadata,
		"required_actions": []any{},
		"vault_ids":        []any{},
		"environment":      map[string]any{"type": "none"},
		"usage": map[string]any{
			"input_tokens":          0,
			"output_tokens":         0,
			"total_tokens":          0,
			"input_tokens_details":  map[string]any{"cached_tokens": 0},
			"output_tokens_details": map[string]any{"reasoning_tokens": 0},
		},
		"agent": map[string]any{
			"id":           session.agentID,
			"model":        session.model,
			"name":         session.name,
			"instructions": "",
			"tools":        []any{},
			"service_tier": "auto",
			"multi_agent":  map[string]any{"enabled": false, "max_concurrent_subagents": 0},
			"reasoning":    map[string]any{"effort": "none", "summary": "auto"},
			"text":         map[string]any{"format": map[string]any{"type": "text"}, "verbosity": "medium"},
		},
	}
}
