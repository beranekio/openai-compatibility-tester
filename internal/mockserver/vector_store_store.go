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
}

type vectorStoreStore struct {
	mu     sync.Mutex
	next   int
	stores map[string]storedVectorStore
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

func cloneVectorStore(store storedVectorStore) storedVectorStore {
	store.metadata = cloneMap(store.metadata)
	return store
}

func vectorStorePayload(store storedVectorStore) map[string]any {
	return map[string]any{
		"id":             store.id,
		"object":         "vector_store",
		"created_at":     store.createdAt,
		"name":           store.name,
		"metadata":       cloneMap(store.metadata),
		"status":         "completed",
		"usage_bytes":    0,
		"last_active_at": store.lastActiveAt,
		"file_counts": map[string]any{
			"in_progress": 0,
			"completed":   0,
			"failed":      0,
			"cancelled":   0,
			"total":       0,
		},
	}
}
