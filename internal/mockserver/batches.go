package mockserver

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *Server) handleBatchCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		InputFileID      string `json:"input_file_id"`
		Endpoint         string `json:"endpoint"`
		CompletionWindow string `json:"completion_window"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.InputFileID == "" {
		http.Error(w, "missing input_file_id", http.StatusBadRequest)
		return
	}
	if _, ok := s.fileStore.get(req.InputFileID); !ok {
		writeNotFound(w, "File not found", "input_file_id")
		return
	}
	if req.Endpoint == "" {
		req.Endpoint = "/v1/chat/completions"
	}
	if req.CompletionWindow == "" {
		req.CompletionWindow = "24h"
	}

	id := s.batchStore.allocateID()
	batch := storedBatch{
		id:               id,
		inputFileID:      req.InputFileID,
		endpoint:         req.Endpoint,
		completionWindow: req.CompletionWindow,
		status:           "validating",
		createdAt:        1700000000,
	}
	s.batchStore.save(id, batch)
	writeJSON(w, batchObjectPayload(batch))
}

func (s *Server) handleBatchGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batch, ok := s.batchStore.advanceStatus(id)
	if !ok {
		writeNotFound(w, "Batch not found", "id")
		return
	}
	writeJSON(w, batchObjectPayload(batch))
}

func (s *Server) handleBatchCancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batch, ok := s.batchStore.cancel(id)
	if !ok {
		writeNotFound(w, "Batch not found", "id")
		return
	}
	writeJSON(w, batchObjectPayload(batch))
}