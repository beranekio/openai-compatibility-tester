package mockserver

import (
	"encoding/json"
	"strconv"
	"sync"
)

type responseStore struct {
	mu         sync.Mutex
	nextID     int
	responses  map[string]map[string]any
	inputItems map[string][]map[string]any
}

func newResponseStore() *responseStore {
	return &responseStore{
		responses:  make(map[string]map[string]any),
		inputItems: make(map[string][]map[string]any),
	}
}

func (s *responseStore) allocateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	return "resp-mock-" + strconv.Itoa(s.nextID)
}

func (s *responseStore) save(id string, payload map[string]any, input []map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responses[id] = payload
	if len(input) > 0 {
		s.inputItems[id] = input
	}
}

func (s *responseStore) get(id string) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, ok := s.responses[id]
	return payload, ok
}

func (s *responseStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.responses[id]; !ok {
		return false
	}
	delete(s.responses, id)
	delete(s.inputItems, id)
	return true
}

func (s *responseStore) updateStatus(id, status string) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, ok := s.responses[id]
	if !ok {
		return nil, false
	}
	updated := cloneMap(payload)
	updated["status"] = status
	s.responses[id] = updated
	return cloneMap(updated), true
}

func (s *responseStore) inputItemsFor(id string) ([]map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.responses[id]; !ok {
		return nil, false
	}
	items := s.inputItems[id]
	out := make([]map[string]any, len(items))
	for i, item := range items {
		out[i] = cloneMap(item)
	}
	return out, true
}

func inputItemsFromRequest(input json.RawMessage) []map[string]any {
	if len(input) == 0 {
		return nil
	}
	var asString string
	if err := json.Unmarshal(input, &asString); err == nil && asString != "" {
		return []map[string]any{{
			"id":   "item-input-1",
			"type": "message",
			"role": "user",
			"content": []map[string]any{{
				"type": "input_text",
				"text": asString,
			}},
		}}
	}
	return nil
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mockResponseUsage() map[string]any {
	return map[string]any{
		"input_tokens":  10,
		"output_tokens": 5,
		"total_tokens":  15,
		"input_tokens_details": map[string]any{
			"cached_tokens": 0,
		},
		"output_tokens_details": map[string]any{
			"reasoning_tokens": 0,
		},
	}
}