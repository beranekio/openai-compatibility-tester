package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedFile struct {
	id        string
	bytes     []byte
	filename  string
	purpose   string
	createdAt int64
}

type fileStore struct {
	mu    sync.Mutex
	next  int
	files map[string]storedFile
}

func newFileStore() *fileStore {
	return &fileStore{
		files: make(map[string]storedFile),
	}
}

func (s *fileStore) allocateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	return "file-mock-" + strconv.Itoa(s.next)
}

func (s *fileStore) save(id string, file storedFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[id] = file
}

func (s *fileStore) get(id string) (storedFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, ok := s.files[id]
	if !ok {
		return storedFile{}, false
	}
	return file, true
}

func (s *fileStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.files[id]; !ok {
		return false
	}
	delete(s.files, id)
	return true
}

func (s *fileStore) list() []storedFile {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]storedFile, 0, len(s.files))
	for _, file := range s.files {
		items = append(items, file)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].id < items[j].id
	})
	return items
}

func fileObjectPayload(file storedFile) map[string]any {
	return map[string]any{
		"id":         file.id,
		"object":     "file",
		"bytes":      len(file.bytes),
		"created_at": file.createdAt,
		"filename":   file.filename,
		"purpose":    file.purpose,
		"status":     "processed",
	}
}