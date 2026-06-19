package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (s *Server) handleThreadCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	thread := s.threadStore.create(req.Metadata)
	writeJSON(w, threadPayload(thread))
}

func (s *Server) handleThreadGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	thread, ok := s.threadStore.get(id)
	if !ok {
		writeNotFound(w, "Thread not found", "thread_id")
		return
	}
	writeJSON(w, threadPayload(thread))
}

func (s *Server) handleThreadUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	thread, ok := s.threadStore.update(id, req.Metadata)
	if !ok {
		writeNotFound(w, "Thread not found", "thread_id")
		return
	}
	writeJSON(w, threadPayload(thread))
}

func (s *Server) handleThreadDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.threadStore.delete(id) {
		writeNotFound(w, "Thread not found", "thread_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "thread.deleted",
		"deleted": true,
	})
}

func (s *Server) handleThreadMessageCreate(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	var req struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	text := threadMessageTextFromContent(req.Content)
	if req.Role == "" {
		req.Role = "user"
	}
	message, ok := s.threadStore.addMessage(threadID, req.Role, text)
	if !ok {
		writeNotFound(w, "Thread not found", "thread_id")
		return
	}
	writeJSON(w, threadMessagePayload(message))
}

func (s *Server) handleThreadMessageList(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	messages, ok := s.threadStore.listMessages(threadID)
	if !ok {
		writeNotFound(w, "Thread not found", "thread_id")
		return
	}
	data := make([]map[string]any, len(messages))
	firstID := ""
	lastID := ""
	for i, message := range messages {
		data[i] = threadMessagePayload(message)
		if i == 0 {
			firstID = message.id
		}
		lastID = message.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleThreadMessageGet(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	messageID := r.PathValue("messageID")
	message, ok := s.threadStore.getMessage(threadID, messageID)
	if !ok {
		writeNotFound(w, "Message not found", "message_id")
		return
	}
	writeJSON(w, threadMessagePayload(message))
}

func (s *Server) handleThreadRunCreate(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	var req struct {
		AssistantID  string `json:"assistant_id"`
		Instructions string `json:"instructions"`
		Model        string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	assistant, ok := s.assistantStore.get(req.AssistantID)
	if !ok {
		writeNotFound(w, "Assistant not found", "assistant_id")
		return
	}
	model := req.Model
	if model == "" {
		model = assistant.model
	}
	instructions := req.Instructions
	if instructions == "" {
		instructions = assistant.instructions
	}
	run, ok := s.threadStore.createRun(threadID, req.AssistantID, model, instructions)
	if !ok {
		writeNotFound(w, "Thread not found", "thread_id")
		return
	}
	writeJSON(w, threadRunPayload(run))
}

func (s *Server) handleThreadRunGet(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	runID := r.PathValue("runID")
	run, ok := s.threadStore.getRun(threadID, runID)
	if !ok {
		writeNotFound(w, "Run not found", "run_id")
		return
	}
	writeJSON(w, threadRunPayload(run))
}

func threadPayload(thread storedThread) map[string]any {
	return map[string]any{
		"id":             thread.id,
		"object":         "thread",
		"created_at":     thread.createdAt,
		"metadata":       thread.metadata,
		"tool_resources": map[string]any{},
	}
}

func threadMessagePayload(message storedThreadMessage) map[string]any {
	payload := map[string]any{
		"id":          message.id,
		"object":      "thread.message",
		"created_at":  message.createdAt,
		"thread_id":   message.threadID,
		"role":        message.role,
		"content": []map[string]any{
			{
				"type": "text",
				"text": map[string]any{
					"value":       message.text,
					"annotations": []any{},
				},
			},
		},
		"attachments":        []any{},
		"metadata":           map[string]any{},
		"status":             message.status,
		"completed_at":       message.completedAt,
		"incomplete_at":      nil,
		"incomplete_details": nil,
	}
	if message.assistantID != "" {
		payload["assistant_id"] = message.assistantID
	} else {
		payload["assistant_id"] = nil
	}
	if message.runID != "" {
		payload["run_id"] = message.runID
	} else {
		payload["run_id"] = nil
	}
	return payload
}

func threadRunPayload(run storedThreadRun) map[string]any {
	return map[string]any{
		"id":                    run.id,
		"object":                "thread.run",
		"created_at":            run.createdAt,
		"thread_id":             run.threadID,
		"assistant_id":          run.assistantID,
		"status":                run.status,
		"started_at":            run.startedAt,
		"completed_at":          run.completedAt,
		"cancelled_at":          nil,
		"failed_at":             nil,
		"expires_at":            nil,
		"instructions":          run.instructions,
		"model":                 run.model,
		"metadata":              map[string]any{},
		"tools":                 []any{},
		"parallel_tool_calls":   true,
		"tool_choice":           "auto",
		"truncation_strategy":   map[string]any{"type": "auto"},
		"usage": map[string]any{
			"prompt_tokens":     5,
			"completion_tokens": 1,
			"total_tokens":      6,
		},
		"max_prompt_tokens":     0,
		"max_completion_tokens": 0,
		"last_error":            nil,
		"incomplete_details":    nil,
		"required_action":       nil,
		"response_format":       "auto",
	}
}

func threadMessageTextFromContent(content any) string {
	switch value := content.(type) {
	case string:
		return value
	case map[string]any:
		if text, ok := value["value"].(string); ok {
			return text
		}
		if textObj, ok := value["text"].(map[string]any); ok {
			if text, ok := textObj["value"].(string); ok {
				return text
			}
		}
	case []any:
		var parts []string
		for _, part := range value {
			if text := threadMessageTextFromContent(part); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "")
	}
	return ""
}