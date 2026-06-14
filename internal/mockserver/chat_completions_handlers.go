package mockserver

import (
	"net/http"
)

func (s *Server) handleChatCompletionGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	payload, ok := s.chatStore.get(id)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Chat completion not found",
				"type":    "invalid_request_error",
				"param":   "id",
				"code":    "not_found",
			},
		})
		return
	}
	writeJSON(w, payload)
}

func (s *Server) handleChatCompletionDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.chatStore.delete(id) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Chat completion not found",
				"type":    "invalid_request_error",
				"param":   "id",
				"code":    "not_found",
			},
		})
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "chat.completion.deleted",
		"deleted": true,
	})
}

func (s *Server) handleChatCompletionList(w http.ResponseWriter, r *http.Request) {
	items := s.chatStore.listAll()
	firstID := ""
	lastID := ""
	if len(items) > 0 {
		if idVal, ok := items[0]["id"].(string); ok {
			firstID = idVal
		}
		if idVal, ok := items[len(items)-1]["id"].(string); ok {
			lastID = idVal
		}
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     items,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleChatCompletionMessages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	items, ok := s.chatStore.messagesFor(id)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Chat completion not found",
				"type":    "invalid_request_error",
				"param":   "id",
				"code":    "not_found",
			},
		})
		return
	}
	firstID := ""
	lastID := ""
	if len(items) > 0 {
		if idVal, ok := items[0]["id"].(string); ok {
			firstID = idVal
		}
		if idVal, ok := items[len(items)-1]["id"].(string); ok {
			lastID = idVal
		}
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     items,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}