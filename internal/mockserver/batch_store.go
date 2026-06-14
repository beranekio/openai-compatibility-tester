package mockserver

import (
	"strconv"
	"sync"
)

type storedBatch struct {
	id               string
	inputFileID      string
	endpoint         string
	completionWindow string
	status           string
	createdAt        int64
}

type batchStore struct {
	mu     sync.Mutex
	next   int
	batches map[string]storedBatch
}

func newBatchStore() *batchStore {
	return &batchStore{
		batches: make(map[string]storedBatch),
	}
}

func (s *batchStore) allocateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	return "batch-mock-" + strconv.Itoa(s.next)
}

func (s *batchStore) save(id string, batch storedBatch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batches[id] = batch
}

func (s *batchStore) get(id string) (storedBatch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.batches[id]
	if !ok {
		return storedBatch{}, false
	}
	return batch, true
}

func (s *batchStore) setStatus(id, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.batches[id]
	if !ok {
		return false
	}
	batch.status = status
	s.batches[id] = batch
	return true
}

func batchObjectPayload(batch storedBatch) map[string]any {
	payload := map[string]any{
		"id":                batch.id,
		"object":            "batch",
		"endpoint":          batch.endpoint,
		"completion_window": batch.completionWindow,
		"input_file_id":     batch.inputFileID,
		"status":            batch.status,
		"created_at":        batch.createdAt,
		"request_counts": map[string]any{
			"total":     1,
			"completed": 0,
			"failed":    0,
		},
	}
	switch batch.status {
	case "completed":
		payload["request_counts"] = map[string]any{
			"total":     1,
			"completed": 1,
			"failed":    0,
		}
		payload["completed_at"] = batch.createdAt + 60
		payload["output_file_id"] = "file-mock-batch-output"
	case "cancelled":
		payload["request_counts"] = map[string]any{
			"total":     1,
			"completed": 0,
			"failed":    0,
		}
		payload["cancelled_at"] = batch.createdAt + 30
	}
	return payload
}