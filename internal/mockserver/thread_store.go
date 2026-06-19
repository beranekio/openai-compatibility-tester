package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedThread struct {
	id        string
	metadata  map[string]any
	createdAt int64
	messages  map[string]storedThreadMessage
	runs      map[string]storedThreadRun
}

type storedThreadMessage struct {
	id          string
	threadID    string
	role        string
	text        string
	assistantID string
	runID       string
	status      string
	createdAt   int64
	completedAt int64
}

type storedThreadRun struct {
	id           string
	threadID     string
	assistantID  string
	model        string
	instructions string
	status       string
	createdAt    int64
	startedAt    int64
	completedAt  int64
}

type threadStore struct {
	mu       sync.Mutex
	next     int
	nextMsg  int
	nextRun  int
	threads  map[string]storedThread
}

func newThreadStore() *threadStore {
	return &threadStore{
		threads: make(map[string]storedThread),
	}
}

func (s *threadStore) create(metadata map[string]any) storedThread {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	thread := storedThread{
		id:        "thread_mock_" + strconv.Itoa(s.next),
		metadata:  cloneMap(metadata),
		createdAt: 1700000000,
		messages:  make(map[string]storedThreadMessage),
		runs:      make(map[string]storedThreadRun),
	}
	s.threads[thread.id] = thread
	return cloneThread(thread)
}

func (s *threadStore) get(id string) (storedThread, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[id]
	if !ok {
		return storedThread{}, false
	}
	return cloneThread(thread), true
}

func (s *threadStore) update(id string, metadata map[string]any) (storedThread, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[id]
	if !ok {
		return storedThread{}, false
	}
	thread.metadata = cloneMap(metadata)
	s.threads[id] = thread
	return cloneThread(thread), true
}

func (s *threadStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.threads[id]; !ok {
		return false
	}
	delete(s.threads, id)
	return true
}

func (s *threadStore) addMessage(threadID, role, text string) (storedThreadMessage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[threadID]
	if !ok {
		return storedThreadMessage{}, false
	}
	s.nextMsg++
	message := storedThreadMessage{
		id:          "msg_mock_" + strconv.Itoa(s.nextMsg),
		threadID:    threadID,
		role:        role,
		text:        text,
		status:      "completed",
		createdAt:   1700000000,
		completedAt: 1700000000,
	}
	thread.messages[message.id] = message
	s.threads[threadID] = thread
	return message, true
}

func (s *threadStore) listMessages(threadID string) ([]storedThreadMessage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[threadID]
	if !ok {
		return nil, false
	}
	ids := make([]string, 0, len(thread.messages))
	for id := range thread.messages {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	messages := make([]storedThreadMessage, len(ids))
	for i, id := range ids {
		messages[i] = thread.messages[id]
	}
	return messages, true
}

func (s *threadStore) getMessage(threadID, messageID string) (storedThreadMessage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[threadID]
	if !ok {
		return storedThreadMessage{}, false
	}
	message, ok := thread.messages[messageID]
	return message, ok
}

func (s *threadStore) createRun(threadID, assistantID, model, instructions string) (storedThreadRun, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[threadID]
	if !ok {
		return storedThreadRun{}, false
	}
	s.nextRun++
	run := storedThreadRun{
		id:           "run_mock_" + strconv.Itoa(s.nextRun),
		threadID:     threadID,
		assistantID:  assistantID,
		model:        model,
		instructions: instructions,
		status:       "completed",
		createdAt:    1700000000,
		startedAt:    1700000000,
		completedAt:  1700000001,
	}
	thread.runs[run.id] = run

	s.nextMsg++
	assistantMessage := storedThreadMessage{
		id:          "msg_mock_" + strconv.Itoa(s.nextMsg),
		threadID:    threadID,
		role:        "assistant",
		text:        assistantThreadAssistantReply,
		assistantID: assistantID,
		runID:       run.id,
		status:      "completed",
		createdAt:   1700000001,
		completedAt: 1700000001,
	}
	thread.messages[assistantMessage.id] = assistantMessage
	s.threads[threadID] = thread
	return run, true
}

func (s *threadStore) getRun(threadID, runID string) (storedThreadRun, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[threadID]
	if !ok {
		return storedThreadRun{}, false
	}
	run, ok := thread.runs[runID]
	return run, ok
}

func cloneThread(thread storedThread) storedThread {
	thread.metadata = cloneMap(thread.metadata)
	thread.messages = make(map[string]storedThreadMessage, len(thread.messages))
	for id, message := range thread.messages {
		thread.messages[id] = message
	}
	thread.runs = make(map[string]storedThreadRun, len(thread.runs))
	for id, run := range thread.runs {
		thread.runs[id] = run
	}
	return thread
}

const assistantThreadAssistantReply = "pong"