package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedVectorStore struct {
	id           string
	name         string
	metadata     map[string]any
	createdAt    int64
	lastActiveAt int64
	files        map[string]storedVectorStoreFile
	batches      map[string]storedVectorStoreFileBatch
}

type storedVectorStoreFile struct {
	fileID        string
	vectorStoreID string
	batchID       string
	status        string
	attributes    map[string]any
	createdAt     int64
	usageBytes    int64
}

type storedVectorStoreFileBatch struct {
	id            string
	vectorStoreID string
	fileIDs       []string
	status        string
	createdAt     int64
}

type vectorStoreStore struct {
	mu        sync.Mutex
	next      int
	nextBatch int
	stores    map[string]storedVectorStore
}

func newVectorStoreStore() *vectorStoreStore {
	return &vectorStoreStore{
		stores: make(map[string]storedVectorStore),
	}
}

func (s *vectorStoreStore) create(name string, metadata map[string]any) storedVectorStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	store := storedVectorStore{
		id:           "vs_mock_" + strconv.Itoa(s.next),
		name:         name,
		metadata:     cloneMap(metadata),
		createdAt:    1700000000,
		lastActiveAt: 1700000000,
		files:        make(map[string]storedVectorStoreFile),
		batches:      make(map[string]storedVectorStoreFileBatch),
	}
	s.stores[store.id] = store
	return store
}

func (s *vectorStoreStore) get(id string) (storedVectorStore, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[id]
	if !ok {
		return storedVectorStore{}, false
	}
	return cloneVectorStore(store), true
}

func (s *vectorStoreStore) update(id string, name *string, metadata map[string]any) (storedVectorStore, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[id]
	if !ok {
		return storedVectorStore{}, false
	}
	if name != nil {
		store.name = *name
	}
	if metadata != nil {
		store.metadata = cloneMap(metadata)
	}
	s.stores[id] = store
	return cloneVectorStore(store), true
}

func (s *vectorStoreStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.stores[id]; !ok {
		return false
	}
	delete(s.stores, id)
	return true
}

func (s *vectorStoreStore) list() []storedVectorStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]storedVectorStore, 0, len(s.stores))
	for _, store := range s.stores {
		items = append(items, cloneVectorStore(store))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].id < items[j].id
	})
	return items
}

func (s *vectorStoreStore) attachFile(vectorStoreID, fileID string, attributes map[string]any, batchID, status string, usageBytes int64) (storedVectorStoreFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return storedVectorStoreFile{}, false
	}
	file := storedVectorStoreFile{
		fileID:        fileID,
		vectorStoreID: vectorStoreID,
		batchID:       batchID,
		status:        status,
		attributes:    cloneMap(attributes),
		createdAt:     1700000000,
		usageBytes:    usageBytes,
	}
	store.files[fileID] = file
	s.stores[vectorStoreID] = store
	return cloneVectorStoreFile(file), true
}

func (s *vectorStoreStore) getFile(vectorStoreID, fileID string) (storedVectorStoreFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return storedVectorStoreFile{}, false
	}
	file, ok := store.files[fileID]
	if !ok {
		return storedVectorStoreFile{}, false
	}
	return cloneVectorStoreFile(file), true
}

func (s *vectorStoreStore) deleteFile(vectorStoreID, fileID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return false
	}
	if _, ok := store.files[fileID]; !ok {
		return false
	}
	delete(store.files, fileID)
	s.stores[vectorStoreID] = store
	return true
}

func (s *vectorStoreStore) listFiles(vectorStoreID string) ([]storedVectorStoreFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return nil, false
	}
	items := make([]storedVectorStoreFile, 0, len(store.files))
	for _, file := range store.files {
		items = append(items, cloneVectorStoreFile(file))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].fileID < items[j].fileID
	})
	return items, true
}

func (s *vectorStoreStore) createFileBatch(vectorStoreID string, fileIDs []string) (storedVectorStoreFileBatch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return storedVectorStoreFileBatch{}, false
	}
	s.nextBatch++
	batch := storedVectorStoreFileBatch{
		id:            "vsfb_mock_" + strconv.Itoa(s.nextBatch),
		vectorStoreID: vectorStoreID,
		fileIDs:       append([]string(nil), fileIDs...),
		status:        "in_progress",
		createdAt:     1700000000,
	}
	store.batches[batch.id] = batch
	for _, fileID := range fileIDs {
		store.files[fileID] = storedVectorStoreFile{
			fileID:        fileID,
			vectorStoreID: vectorStoreID,
			batchID:       batch.id,
			status:        "in_progress",
			attributes:    map[string]any{},
			createdAt:     1700000000,
			usageBytes:    24,
		}
	}
	s.stores[vectorStoreID] = store
	return cloneVectorStoreFileBatch(batch), true
}

func (s *vectorStoreStore) getFileBatch(vectorStoreID, batchID string) (storedVectorStoreFileBatch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return storedVectorStoreFileBatch{}, false
	}
	batch, ok := store.batches[batchID]
	if !ok {
		return storedVectorStoreFileBatch{}, false
	}
	return cloneVectorStoreFileBatch(batch), true
}

func (s *vectorStoreStore) cancelFileBatch(vectorStoreID, batchID string) (storedVectorStoreFileBatch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return storedVectorStoreFileBatch{}, false
	}
	batch, ok := store.batches[batchID]
	if !ok {
		return storedVectorStoreFileBatch{}, false
	}
	batch.status = "cancelled"
	store.batches[batchID] = batch
	for _, fileID := range batch.fileIDs {
		file := store.files[fileID]
		file.status = "cancelled"
		store.files[fileID] = file
	}
	s.stores[vectorStoreID] = store
	return cloneVectorStoreFileBatch(batch), true
}

func (s *vectorStoreStore) listBatchFiles(vectorStoreID, batchID string) ([]storedVectorStoreFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, ok := s.stores[vectorStoreID]
	if !ok {
		return nil, false
	}
	batch, ok := store.batches[batchID]
	if !ok {
		return nil, false
	}
	items := make([]storedVectorStoreFile, 0, len(batch.fileIDs))
	for _, fileID := range batch.fileIDs {
		if file, ok := store.files[fileID]; ok {
			items = append(items, cloneVectorStoreFile(file))
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].fileID < items[j].fileID
	})
	return items, true
}

func cloneVectorStore(store storedVectorStore) storedVectorStore {
	store.metadata = cloneMap(store.metadata)
	store.files = make(map[string]storedVectorStoreFile, len(store.files))
	for id, file := range store.files {
		store.files[id] = cloneVectorStoreFile(file)
	}
	store.batches = make(map[string]storedVectorStoreFileBatch, len(store.batches))
	for id, batch := range store.batches {
		store.batches[id] = cloneVectorStoreFileBatch(batch)
	}
	return store
}

func cloneVectorStoreFile(file storedVectorStoreFile) storedVectorStoreFile {
	file.attributes = cloneMap(file.attributes)
	return file
}

func cloneVectorStoreFileBatch(batch storedVectorStoreFileBatch) storedVectorStoreFileBatch {
	batch.fileIDs = append([]string(nil), batch.fileIDs...)
	return batch
}

func vectorStorePayload(store storedVectorStore) map[string]any {
	cancelled := 0
	completed := 0
	failed := 0
	inProgress := 0
	usageBytes := int64(0)
	for _, file := range store.files {
		usageBytes += file.usageBytes
		switch file.status {
		case "cancelled":
			cancelled++
		case "failed":
			failed++
		case "in_progress":
			inProgress++
		default:
			completed++
		}
	}
	return map[string]any{
		"id":             store.id,
		"object":         "vector_store",
		"created_at":     store.createdAt,
		"name":           store.name,
		"metadata":       cloneMap(store.metadata),
		"status":         "completed",
		"usage_bytes":    usageBytes,
		"last_active_at": store.lastActiveAt,
		"file_counts": map[string]any{
			"in_progress": inProgress,
			"completed":   completed,
			"failed":      failed,
			"cancelled":   cancelled,
			"total":       len(store.files),
		},
	}
}

func vectorStoreFilePayload(file storedVectorStoreFile) map[string]any {
	return map[string]any{
		"id":              file.fileID,
		"object":          "vector_store.file",
		"created_at":      file.createdAt,
		"vector_store_id": file.vectorStoreID,
		"status":          file.status,
		"usage_bytes":     file.usageBytes,
		"last_error":      nil,
		"attributes":      cloneMap(file.attributes),
	}
}

func vectorStoreFileBatchPayload(batch storedVectorStoreFileBatch) map[string]any {
	cancelled := 0
	completed := 0
	inProgress := 0
	switch batch.status {
	case "cancelled":
		cancelled = len(batch.fileIDs)
	case "completed":
		completed = len(batch.fileIDs)
	default:
		inProgress = len(batch.fileIDs)
	}
	return map[string]any{
		"id":              batch.id,
		"object":          "vector_store.files_batch",
		"created_at":      batch.createdAt,
		"vector_store_id": batch.vectorStoreID,
		"status":          batch.status,
		"file_counts": map[string]any{
			"in_progress": inProgress,
			"completed":   completed,
			"failed":      0,
			"cancelled":   cancelled,
			"total":       len(batch.fileIDs),
		},
	}
}
