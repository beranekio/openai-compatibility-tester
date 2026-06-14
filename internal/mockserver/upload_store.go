package mockserver

import (
	"strconv"
	"sync"
)

type storedUploadPart struct {
	id        string
	uploadID  string
	bytes     []byte
	createdAt int64
}

type storedUpload struct {
	id        string
	bytes     int64
	filename  string
	mimeType  string
	purpose   string
	createdAt int64
	expiresAt int64
	status    string
	fileID    string
	parts     map[string]storedUploadPart
}

type uploadStore struct {
	mu       sync.Mutex
	next     int
	nextPart int
	uploads  map[string]storedUpload
}

func newUploadStore() *uploadStore {
	return &uploadStore{
		uploads: make(map[string]storedUpload),
	}
}

func (s *uploadStore) allocateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	return "upload-mock-" + strconv.Itoa(s.next)
}

func (s *uploadStore) allocatePartID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextPart++
	return "part-mock-" + strconv.Itoa(s.nextPart)
}

func (s *uploadStore) save(upload storedUpload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploads[upload.id] = upload
}

func (s *uploadStore) get(id string) (storedUpload, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	upload, ok := s.uploads[id]
	if !ok {
		return storedUpload{}, false
	}
	return upload, true
}

func (s *uploadStore) update(id string, fn func(*storedUpload)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	upload, ok := s.uploads[id]
	if !ok {
		return false
	}
	fn(&upload)
	s.uploads[id] = upload
	return true
}

func uploadObjectPayload(upload storedUpload) map[string]any {
	payload := map[string]any{
		"id":         upload.id,
		"object":     "upload",
		"bytes":      upload.bytes,
		"created_at": upload.createdAt,
		"expires_at": upload.expiresAt,
		"filename":   upload.filename,
		"purpose":    upload.purpose,
		"status":     upload.status,
	}
	return payload
}

func uploadPartObjectPayload(part storedUploadPart) map[string]any {
	return map[string]any{
		"id":         part.id,
		"object":     "upload.part",
		"created_at": part.createdAt,
		"upload_id":  part.uploadID,
	}
}