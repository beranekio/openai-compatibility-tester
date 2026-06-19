package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (s *Server) handleContainerCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "missing name", http.StatusBadRequest)
		return
	}
	container := s.containerStore.create(req.Name)
	writeJSON(w, containerPayload(container))
}

func (s *Server) handleContainerList(w http.ResponseWriter, _ *http.Request) {
	items := s.containerStore.list()
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, container := range items {
		data[i] = containerPayload(container)
		if i == 0 {
			firstID = container.id
		}
		lastID = container.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}

func (s *Server) handleContainerGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	container, ok := s.containerStore.get(id)
	if !ok {
		writeNotFound(w, "Container not found", "container_id")
		return
	}
	writeJSON(w, containerPayload(container))
}

func (s *Server) handleContainerDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.containerStore.delete(id) {
		writeNotFound(w, "Container not found", "container_id")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleContainerFileCreate(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		s.handleContainerFileCreateMultipart(w, r, containerID)
		return
	}

	var req struct {
		FileID string `json:"file_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.FileID == "" {
		http.Error(w, "missing file_id", http.StatusBadRequest)
		return
	}
	file, ok := s.fileStore.get(req.FileID)
	if !ok {
		writeNotFound(w, "File not found", "file_id")
		return
	}
	created, ok := s.containerStore.addFile(containerID, file.bytes, "/"+file.filename, "user")
	if !ok {
		writeNotFound(w, "Container not found", "container_id")
		return
	}
	writeJSON(w, containerFilePayload(created))
}

func (s *Server) handleContainerFileCreateMultipart(w http.ResponseWriter, r *http.Request, containerID string) {
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

	path := header.Filename
	if path == "" {
		path = "upload"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	created, ok := s.containerStore.addFile(containerID, data, path, "user")
	if !ok {
		writeNotFound(w, "Container not found", "container_id")
		return
	}
	writeJSON(w, containerFilePayload(created))
}

func (s *Server) handleContainerFileList(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")
	files, ok := s.containerStore.listFiles(containerID)
	if !ok {
		writeNotFound(w, "Container not found", "container_id")
		return
	}
	data := make([]map[string]any, len(files))
	firstID := ""
	lastID := ""
	for i, file := range files {
		data[i] = containerFilePayload(file)
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

func (s *Server) handleContainerFileGet(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")
	fileID := r.PathValue("fileID")
	file, ok := s.containerStore.getFile(containerID, fileID)
	if !ok {
		writeNotFound(w, "Container file not found", "file_id")
		return
	}
	writeJSON(w, containerFilePayload(file))
}

func (s *Server) handleContainerFileDelete(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")
	fileID := r.PathValue("fileID")
	if !s.containerStore.deleteFile(containerID, fileID) {
		writeNotFound(w, "Container file not found", "file_id")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleContainerFileContent(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")
	fileID := r.PathValue("fileID")
	file, ok := s.containerStore.getFile(containerID, fileID)
	if !ok {
		writeNotFound(w, "Container file not found", "file_id")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file.bytes)
}