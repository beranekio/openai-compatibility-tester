package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (s *Server) handleVectorStoreCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string         `json:"name"`
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = "mock vector store"
	}
	store := s.vectorStoreStore.create(req.Name, req.Metadata)
	writeJSON(w, vectorStorePayload(store))
}

func (s *Server) handleVectorStoreList(w http.ResponseWriter, _ *http.Request) {
	items := s.vectorStoreStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, store := range items {
		data[i] = vectorStorePayload(store)
		if i == 0 {
			firstID = store.id
		}
		lastID = store.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleVectorStoreGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store, ok := s.vectorStoreStore.get(id)
	if !ok {
		writeNotFound(w, "Vector store not found", "vector_store_id")
		return
	}
	writeJSON(w, vectorStorePayload(store))
}

func (s *Server) handleVectorStoreUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name     *string        `json:"name"`
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	store, ok := s.vectorStoreStore.update(id, req.Name, req.Metadata)
	if !ok {
		writeNotFound(w, "Vector store not found", "vector_store_id")
		return
	}
	writeJSON(w, vectorStorePayload(store))
}

func (s *Server) handleVectorStoreDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.vectorStoreStore.delete(id) {
		writeNotFound(w, "Vector store not found", "vector_store_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "vector_store.deleted",
		"deleted": true,
	})
}

func (s *Server) handleVectorStoreSearch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.vectorStoreStore.get(id); !ok {
		writeNotFound(w, "Vector store not found", "vector_store_id")
		return
	}
	writeJSON(w, map[string]any{
		"object": "vector_store.search_results.page",
		"data":   []map[string]any{},
	})
}
