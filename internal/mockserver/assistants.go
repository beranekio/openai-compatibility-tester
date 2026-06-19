package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (s *Server) handleAssistantCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model        string         `json:"model"`
		Name         string         `json:"name"`
		Description  string         `json:"description"`
		Instructions string         `json:"instructions"`
		Metadata     map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}
	assistant := s.assistantStore.create(req.Model, req.Name, req.Description, req.Instructions, req.Metadata)
	writeJSON(w, assistantPayload(assistant))
}

func (s *Server) handleAssistantList(w http.ResponseWriter, _ *http.Request) {
	items := s.assistantStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, assistant := range items {
		data[i] = assistantPayload(assistant)
		if i == 0 {
			firstID = assistant.id
		}
		lastID = assistant.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleAssistantGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	assistant, ok := s.assistantStore.get(id)
	if !ok {
		writeNotFound(w, "Assistant not found", "assistant_id")
		return
	}
	writeJSON(w, assistantPayload(assistant))
}

func (s *Server) handleAssistantUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name         *string        `json:"name"`
		Description  *string        `json:"description"`
		Instructions *string        `json:"instructions"`
		Metadata     map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	assistant, ok := s.assistantStore.update(id, req.Name, req.Description, req.Instructions, req.Metadata)
	if !ok {
		writeNotFound(w, "Assistant not found", "assistant_id")
		return
	}
	writeJSON(w, assistantPayload(assistant))
}

func (s *Server) handleAssistantDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.assistantStore.delete(id) {
		writeNotFound(w, "Assistant not found", "assistant_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "assistant.deleted",
		"deleted": true,
	})
}

func assistantPayload(assistant storedAssistant) map[string]any {
	return map[string]any{
		"id":            assistant.id,
		"object":        "assistant",
		"created_at":    assistant.createdAt,
		"name":          assistant.name,
		"description":   assistant.description,
		"instructions":  assistant.instructions,
		"model":         assistant.model,
		"metadata":      assistant.metadata,
		"tools":         []any{},
		"tool_resources": map[string]any{},
	}
}