package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedAssistant struct {
	id           string
	model        string
	name         string
	description  string
	instructions string
	metadata     map[string]any
	createdAt    int64
}

type assistantStore struct {
	mu         sync.Mutex
	next       int
	assistants map[string]storedAssistant
}

func newAssistantStore() *assistantStore {
	return &assistantStore{
		assistants: make(map[string]storedAssistant),
	}
}

func (s *assistantStore) create(model, name, description, instructions string, metadata map[string]any) storedAssistant {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	assistant := storedAssistant{
		id:           "asst_mock_" + strconv.Itoa(s.next),
		model:        model,
		name:         name,
		description:  description,
		instructions: instructions,
		metadata:     cloneMap(metadata),
		createdAt:    1700000000,
	}
	s.assistants[assistant.id] = assistant
	return cloneAssistant(assistant)
}

func (s *assistantStore) get(id string) (storedAssistant, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	assistant, ok := s.assistants[id]
	if !ok {
		return storedAssistant{}, false
	}
	return cloneAssistant(assistant), true
}

func (s *assistantStore) update(id string, name, description, instructions *string, metadata map[string]any) (storedAssistant, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	assistant, ok := s.assistants[id]
	if !ok {
		return storedAssistant{}, false
	}
	if name != nil {
		assistant.name = *name
	}
	if description != nil {
		assistant.description = *description
	}
	if instructions != nil {
		assistant.instructions = *instructions
	}
	if metadata != nil {
		assistant.metadata = cloneMap(metadata)
	}
	s.assistants[id] = assistant
	return cloneAssistant(assistant), true
}

func (s *assistantStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assistants[id]; !ok {
		return false
	}
	delete(s.assistants, id)
	return true
}

func (s *assistantStore) list() []storedAssistant {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.assistants))
	for id := range s.assistants {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	items := make([]storedAssistant, len(ids))
	for i, id := range ids {
		items[i] = cloneAssistant(s.assistants[id])
	}
	return items
}

func cloneAssistant(assistant storedAssistant) storedAssistant {
	assistant.metadata = cloneMap(assistant.metadata)
	return assistant
}