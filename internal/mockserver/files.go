package mockserver

import (
	"io"
	"net/http"
)

func (s *Server) handleFileCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	purpose := r.FormValue("purpose")
	if purpose == "" {
		purpose = "user_data"
	}
	filename := header.Filename
	if filename == "" {
		filename = r.FormValue("filename")
	}
	if filename == "" {
		filename = "upload"
	}

	id := s.fileStore.allocateID()
	stored := storedFile{
		id:        id,
		bytes:     append([]byte(nil), data...),
		filename:  filename,
		purpose:   purpose,
		createdAt: 1700000000,
	}
	s.fileStore.save(id, stored)
	writeJSON(w, fileObjectPayload(stored))
}

func (s *Server) handleFileList(w http.ResponseWriter, _ *http.Request) {
	items := s.fileStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, file := range items {
		data[i] = fileObjectPayload(file)
		if i == 0 {
			firstID = file.id
		}
		lastID = file.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleFileGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	file, ok := s.fileStore.get(id)
	if !ok {
		writeNotFound(w, "File not found", "id")
		return
	}
	writeJSON(w, fileObjectPayload(file))
}

func (s *Server) handleFileDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.fileStore.delete(id) {
		writeNotFound(w, "File not found", "id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "file",
		"deleted": true,
	})
}

func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	file, ok := s.fileStore.get(id)
	if !ok {
		writeNotFound(w, "File not found", "id")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file.bytes)
}

func writeNotFound(w http.ResponseWriter, message, param string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	writeJSON(w, map[string]any{
		"error": map[string]any{
			"message": message,
			"type":    "invalid_request_error",
			"param":   param,
			"code":    "not_found",
		},
	})
}