package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (s *Server) handleVectorStoreFileBatchCreate(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	var req struct {
		FileIDs []string `json:"file_ids"`
		Files   []struct {
			FileID string `json:"file_id"`
		} `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, file := range req.Files {
		req.FileIDs = append(req.FileIDs, file.FileID)
	}
	if _, ok := s.vectorStoreStore.get(vectorStoreID); !ok {
		writeNotFound(w, "Vector store not found", "vector_store_id")
		return
	}
	for _, fileID := range req.FileIDs {
		if _, ok := s.fileStore.get(fileID); !ok {
			writeNotFound(w, "File not found", "file_id")
			return
		}
	}
	batch, _ := s.vectorStoreStore.createFileBatch(vectorStoreID, req.FileIDs)
	writeJSON(w, vectorStoreFileBatchPayload(batch))
}

func (s *Server) handleVectorStoreFileBatchGet(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	batchID := r.PathValue("batchID")
	batch, ok := s.vectorStoreStore.getFileBatch(vectorStoreID, batchID)
	if !ok {
		writeNotFound(w, "Vector store file batch not found", "batch_id")
		return
	}
	writeJSON(w, vectorStoreFileBatchPayload(batch))
}

func (s *Server) handleVectorStoreFileBatchCancel(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	batchID := r.PathValue("batchID")
	batch, ok := s.vectorStoreStore.cancelFileBatch(vectorStoreID, batchID)
	if !ok {
		writeNotFound(w, "Vector store file batch not found", "batch_id")
		return
	}
	writeJSON(w, vectorStoreFileBatchPayload(batch))
}

func (s *Server) handleVectorStoreFileBatchListFiles(w http.ResponseWriter, r *http.Request) {
	vectorStoreID := r.PathValue("id")
	batchID := r.PathValue("batchID")
	items, ok := s.vectorStoreStore.listBatchFiles(vectorStoreID, batchID)
	if !ok {
		writeNotFound(w, "Vector store file batch not found", "batch_id")
		return
	}
	writeVectorStoreFileList(w, items)
}
