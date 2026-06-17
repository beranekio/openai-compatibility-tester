package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (s *Server) handleVectorStoreFileCreate(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	var req struct {
		FileID     string         `json:"file_id"`
		Attributes map[string]any `json:"attributes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := s.vectorStoreStore.get(vectorStoreID); !ok {
		writeNotFound(w, "Vector store not found", "vector_store_id")
		return
	}
	file, ok := s.fileStore.get(req.FileID)
	if !ok {
		writeNotFound(w, "File not found", "file_id")
		return
	}
	attached, _ := s.vectorStoreStore.attachFile(vectorStoreID, req.FileID, req.Attributes, "", "completed", int64(len(file.bytes)))
	writeJSON(w, vectorStoreFilePayload(attached))
}

func (s *Server) handleVectorStoreFileList(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	items, ok := s.vectorStoreStore.listFiles(vectorStoreID)
	if !ok {
		writeNotFound(w, "Vector store not found", "vector_store_id")
		return
	}
	writeVectorStoreFileList(w, items)
}

func (s *Server) handleVectorStoreFileGet(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	fileID := r.PathValue("fileID")
	file, ok := s.vectorStoreStore.getFile(vectorStoreID, fileID)
	if !ok {
		writeNotFound(w, "Vector store file not found", "file_id")
		return
	}
	writeJSON(w, vectorStoreFilePayload(file))
}

func (s *Server) handleVectorStoreFileDelete(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	fileID := r.PathValue("fileID")
	if !s.vectorStoreStore.deleteFile(vectorStoreID, fileID) {
		writeNotFound(w, "Vector store file not found", "file_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      fileID,
		"object":  "vector_store.file.deleted",
		"deleted": true,
	})
}

func writeVectorStoreFileList(w http.ResponseWriter, items []storedVectorStoreFile) {
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, file := range items {
		data[i] = vectorStoreFilePayload(file)
		if i == 0 {
			firstID = file.fileID
		}
		lastID = file.fileID
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}
