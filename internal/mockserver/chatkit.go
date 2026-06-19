package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const chatKitBetaHeaderValue = "chatkit_beta=v1"

func requireChatKitBetaHeader(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(r.Header.Get("OpenAI-Beta")) != chatKitBetaHeaderValue {
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

func (s *Server) handleChatKitSessionCreate(w http.ResponseWriter, r *http.Request) {
	if !requireChatKitBetaHeader(w, r) {
		return
	}
	var req struct {
		User     string `json:"user"`
		Workflow struct {
			ID string `json:"id"`
		} `json:"workflow"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.User == "" {
		req.User = chatkitSeedUser
	}
	if req.Workflow.ID == "" {
		req.Workflow.ID = "wf_mock_compat_test"
	}
	session := s.chatKitStore.createSession(req.User, req.Workflow.ID)
	writeJSON(w, chatKitSessionPayload(session))
}

func (s *Server) handleChatKitSessionCancel(w http.ResponseWriter, r *http.Request) {
	if !requireChatKitBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	session, ok := s.chatKitStore.cancelSession(id)
	if !ok {
		writeNotFound(w, "ChatKit session not found", "session_id")
		return
	}
	writeJSON(w, chatKitSessionCancelPayload(session))
}

func (s *Server) handleChatKitThreadGet(w http.ResponseWriter, r *http.Request) {
	if !requireChatKitBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	thread, ok := s.chatKitStore.getThread(id)
	if !ok {
		writeNotFound(w, "ChatKit thread not found", "thread_id")
		return
	}
	writeJSON(w, chatKitThreadPayload(thread))
}

func (s *Server) handleChatKitThreadList(w http.ResponseWriter, r *http.Request) {
	if !requireChatKitBetaHeader(w, r) {
		return
	}
	user := r.URL.Query().Get("user")
	threads := s.chatKitStore.listThreads(user)
	data := make([]map[string]any, len(threads))
	firstID := ""
	lastID := ""
	for i, thread := range threads {
		data[i] = chatKitThreadPayload(thread)
		if i == 0 {
			firstID = thread.id
		}
		lastID = thread.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleChatKitThreadDelete(w http.ResponseWriter, r *http.Request) {
	if !requireChatKitBetaHeader(w, r) {
		return
	}
	id := r.PathValue("id")
	if !s.chatKitStore.deleteThread(id) {
		writeNotFound(w, "ChatKit thread not found", "thread_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"deleted": true,
		"object":  "chatkit.thread.deleted",
	})
}

func (s *Server) handleChatKitThreadListItems(w http.ResponseWriter, r *http.Request) {
	if !requireChatKitBetaHeader(w, r) {
		return
	}
	threadID := r.PathValue("id")
	thread, ok := s.chatKitStore.getThread(threadID)
	if !ok {
		writeNotFound(w, "ChatKit thread not found", "thread_id")
		return
	}
	data := make([]map[string]any, len(thread.items))
	firstID := ""
	lastID := ""
	for i, item := range thread.items {
		data[i] = chatKitThreadItemPayload(threadID, item)
		if i == 0 {
			firstID = item.id
		}
		lastID = item.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func chatKitSessionCancelPayload(session storedChatKitSession) map[string]any {
	return map[string]any{
		"id":     session.id,
		"object": "chatkit.session",
		"status": session.status,
		"workflow": map[string]any{
			"id": session.workflowID,
		},
	}
}

func chatKitSessionPayload(session storedChatKitSession) map[string]any {
	expiresAt := session.createdAt + 600
	maxRequests := int64(10)
	return map[string]any{
		"id":            session.id,
		"client_secret": "cks_mock_secret_" + session.id,
		"expires_at":    expiresAt,
		"max_requests_per_1_minute": maxRequests,
		"object":                    "chatkit.session",
		"status":                    session.status,
		"user":                      session.user,
		"chatkit_configuration": map[string]any{
			"automatic_thread_titling": map[string]any{"enabled": true},
			"file_upload": map[string]any{
				"enabled":       false,
				"max_file_size": 512,
				"max_files":     10,
			},
			"history": map[string]any{
				"enabled":        true,
				"recent_threads": nil,
			},
		},
		"rate_limits": map[string]any{
			"max_requests_per_1_minute": maxRequests,
		},
		"workflow": map[string]any{
			"id":              session.workflowID,
			"state_variables": nil,
			"tracing":         map[string]any{"enabled": true},
			"version":         nil,
		},
	}
}

func chatKitThreadPayload(thread storedChatKitThread) map[string]any {
	return map[string]any{
		"id":         thread.id,
		"created_at": thread.createdAt,
		"object":     "chatkit.thread",
		"status": map[string]any{
			"type": "active",
		},
		"title": thread.title,
		"user":  thread.user,
	}
}

func chatKitThreadItemPayload(threadID string, item storedChatKitThreadItem) map[string]any {
	createdAt := item.createdAt
	if createdAt == 0 {
		createdAt = time.Now().Unix()
	}
	return map[string]any{
		"id":         item.id,
		"object":     "chatkit.thread_item",
		"thread_id":  threadID,
		"type":       "chatkit.user_message",
		"created_at": createdAt,
		"attachments": []any{},
		"content": []map[string]any{{
			"type": "input_text",
			"text": item.text,
		}},
		"inference_options": map[string]any{
			"model": nil,
			"tool_choice": map[string]any{
				"id": nil,
			},
		},
	}
}

