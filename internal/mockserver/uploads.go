package mockserver

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *Server) handleUploadCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Bytes    int64  `json:"bytes"`
		Filename string `json:"filename"`
		MimeType string `json:"mime_type"`
		Purpose  string `json:"purpose"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Bytes <= 0 {
		http.Error(w, "bytes must be positive", http.StatusBadRequest)
		return
	}
	if req.Filename == "" {
		http.Error(w, "missing filename", http.StatusBadRequest)
		return
	}
	if req.MimeType == "" {
		http.Error(w, "missing mime_type", http.StatusBadRequest)
		return
	}
	if req.Purpose == "" {
		req.Purpose = "user_data"
	}

	id := s.uploadStore.allocateID()
	upload := storedUpload{
		id:        id,
		bytes:     req.Bytes,
		filename:  req.Filename,
		mimeType:  req.MimeType,
		purpose:   req.Purpose,
		createdAt: 1700000000,
		expiresAt: 1700003600,
		status:    "pending",
		parts:     make(map[string]storedUploadPart),
	}
	s.uploadStore.save(upload)
	writeJSON(w, uploadObjectPayload(upload))
}

func (s *Server) handleUploadPartCreate(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	upload, ok := s.uploadStore.get(uploadID)
	if !ok {
		writeNotFound(w, "Upload not found", "id")
		return
	}
	if upload.status != "pending" {
		http.Error(w, "upload is not pending", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	dataField, _, err := r.FormFile("data")
	if err != nil {
		http.Error(w, "missing data", http.StatusBadRequest)
		return
	}
	defer dataField.Close()

	data, err := io.ReadAll(dataField)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	partID := s.uploadStore.allocatePartID()
	part := storedUploadPart{
		id:        partID,
		uploadID:  uploadID,
		bytes:     append([]byte(nil), data...),
		createdAt: 1700000000,
	}

	ok = s.uploadStore.update(uploadID, func(upload *storedUpload) {
		upload.parts[partID] = part
	})
	if !ok {
		writeNotFound(w, "Upload not found", "id")
		return
	}

	writeJSON(w, uploadPartObjectPayload(part))
}

func (s *Server) handleUploadComplete(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	upload, ok := s.uploadStore.get(uploadID)
	if !ok {
		writeNotFound(w, "Upload not found", "id")
		return
	}
	if upload.status != "pending" {
		http.Error(w, "upload is not pending", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		PartIDs []string `json:"part_ids"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.PartIDs) == 0 {
		http.Error(w, "missing part_ids", http.StatusBadRequest)
		return
	}

	assembled := make([]byte, 0, upload.bytes)
	for _, partID := range req.PartIDs {
		part, ok := upload.parts[partID]
		if !ok {
			writeNotFound(w, "Part not found", "part_ids")
			return
		}
		assembled = append(assembled, part.bytes...)
	}
	if int64(len(assembled)) != upload.bytes {
		http.Error(w, "assembled bytes do not match upload bytes", http.StatusBadRequest)
		return
	}

	fileID := s.fileStore.allocateID()
	stored := storedFile{
		id:        fileID,
		bytes:     assembled,
		filename:  upload.filename,
		purpose:   upload.purpose,
		createdAt: 1700000000,
	}
	s.fileStore.save(fileID, stored)

	ok = s.uploadStore.update(uploadID, func(upload *storedUpload) {
		upload.status = "completed"
		upload.fileID = fileID
	})
	if !ok {
		writeNotFound(w, "Upload not found", "id")
		return
	}

	upload, _ = s.uploadStore.get(uploadID)
	payload := uploadObjectPayload(upload)
	payload["file"] = fileObjectPayload(stored)
	writeJSON(w, payload)
}

func (s *Server) handleUploadCancel(w http.ResponseWriter, r *http.Request) {
	uploadID := r.PathValue("id")
	ok := s.uploadStore.update(uploadID, func(upload *storedUpload) {
		upload.status = "cancelled"
	})
	if !ok {
		writeNotFound(w, "Upload not found", "id")
		return
	}
	upload, _ := s.uploadStore.get(uploadID)
	writeJSON(w, uploadObjectPayload(upload))
}