package mockserver

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *Server) handleResponses(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Stream     bool              `json:"stream"`
		Store      *bool             `json:"store"`
		Background *bool             `json:"background"`
		Input      json.RawMessage   `json:"input"`
		Tools      []json.RawMessage `json:"tools"`
		Text       *struct {
			Format *struct {
				Type   string `json:"type"`
				Strict *bool  `json:"strict"`
			} `json:"format"`
		} `json:"text"`
	}
	_ = json.Unmarshal(body, &req)

	if len(req.Tools) > 0 {
		if req.Stream {
			writeResponsesToolCallStream(w)
			return
		}
		writeResponsesToolCallResponse(w)
		return
	}

	if req.Stream {
		s.writeResponsesTextStream(w)
		return
	}

	status := "completed"
	if req.Background != nil && *req.Background {
		status = "queued"
	}

	id := "resp-mock"
	if req.Store != nil && *req.Store {
		id = s.store.allocateID()
	}

	outputText := "pong"
	if req.Text != nil && req.Text.Format != nil && req.Text.Format.Type == "json_schema" {
		if req.Text.Format.Strict != nil && *req.Text.Format.Strict {
			outputText = `{"answer":"pong"}`
		}
	}

	payload := map[string]any{
		"id":         id,
		"object":     "response",
		"status":     status,
		"model":      "gpt-4o-mini",
		"created_at": 1700000000,
		"output": []map[string]any{
			{
				"id":     "msg-mock",
				"type":   "message",
				"role":   "assistant",
				"status": "completed",
				"content": []map[string]any{
					{
						"type": "output_text",
						"text": outputText,
					},
				},
			},
		},
	}

	if req.Store != nil && *req.Store {
		s.store.save(id, payload, inputItemsFromRequest(req.Input))
	}

	writeJSON(w, payload)
}

func (s *Server) writeResponsesTextStream(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	seq := 0
	writeResponseStreamEvent := func(payload map[string]any) {
		payload["sequence_number"] = seq
		seq++
		data, _ := json.Marshal(payload)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
	}

	writeResponseStreamEvent(map[string]any{
		"type": "response.created",
		"response": map[string]any{
			"id":         "resp-mock",
			"object":     "response",
			"status":     "in_progress",
			"model":      "gpt-4o-mini",
			"created_at": 1700000000,
		},
	})
	writeResponseStreamEvent(map[string]any{
		"type": "response.in_progress",
		"response": map[string]any{
			"id":         "resp-mock",
			"object":     "response",
			"status":     "in_progress",
			"model":      "gpt-4o-mini",
			"created_at": 1700000000,
		},
	})
	writeResponseStreamEvent(map[string]any{
		"type":         "response.output_item.added",
		"output_index": 0,
		"item": map[string]any{
			"id":     "msg-mock",
			"type":   "message",
			"role":   "assistant",
			"status": "in_progress",
		},
	})
	writeResponseStreamEvent(map[string]any{
		"type":          "response.content_part.added",
		"item_id":       "msg-mock",
		"output_index":  0,
		"content_index": 0,
		"part": map[string]any{
			"type": "output_text",
			"text": "",
		},
	})

	chunks := []string{"one", " two", " three"}
	for _, chunk := range chunks {
		writeResponseStreamEvent(map[string]any{
			"type":          "response.output_text.delta",
			"content_index": 0,
			"item_id":       "msg-mock",
			"output_index":  0,
			"logprobs":      []any{},
			"delta":         chunk,
		})
	}
	writeResponseStreamEvent(map[string]any{
		"type": "response.completed",
		"response": map[string]any{
			"id":         "resp-mock",
			"object":     "response",
			"status":     "completed",
			"model":      "gpt-4o-mini",
			"created_at": 1700000000,
		},
	})
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
}

func (s *Server) handleResponseGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	payload, ok := s.store.get(id)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Response not found",
				"type":    "invalid_request_error",
				"param":   "id",
				"code":    "not_found",
			},
		})
		return
	}
	writeJSON(w, payload)
}

func (s *Server) handleResponseDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.store.delete(id) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Response not found",
				"type":    "invalid_request_error",
				"param":   "id",
				"code":    "not_found",
			},
		})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleResponseCancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	payload, ok := s.store.updateStatus(id, "cancelled")
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Response not found",
				"type":    "invalid_request_error",
				"param":   "id",
				"code":    "not_found",
			},
		})
		return
	}
	writeJSON(w, payload)
}

func (s *Server) handleResponseInputItems(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	items, ok := s.store.inputItemsFor(id)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]any{
			"error": map[string]any{
				"message": "Response not found",
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