package mockserver

import (
	"slices"
	"strconv"
	"sync"
	"time"
)

const chatkitSeedUser = "compatibility-test-user"

type storedChatKitSession struct {
	id         string
	user       string
	workflowID string
	status     string
	createdAt  int64
}

type storedChatKitThread struct {
	id        string
	user      string
	title     string
	createdAt int64
	items     []storedChatKitThreadItem
}

type storedChatKitThreadItem struct {
	id        string
	text      string
	createdAt int64
}

type chatKitStore struct {
	mu       sync.Mutex
	nextSess int
	nextThr  int
	nextItem int
	sessions map[string]storedChatKitSession
	threads  map[string]storedChatKitThread
}

func newChatKitStore() *chatKitStore {
	s := &chatKitStore{
		sessions: make(map[string]storedChatKitSession),
		threads:  make(map[string]storedChatKitThread),
	}
	s.seedThread()
	return s
}

func (s *chatKitStore) seedThread() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextThr++
	threadID := "cthr_mock_" + strconv.Itoa(s.nextThr)
	s.nextItem++
	itemID := "cthi_mock_" + strconv.Itoa(s.nextItem)
	now := time.Now().Unix()
	s.threads[threadID] = storedChatKitThread{
		id:        threadID,
		user:      chatkitSeedUser,
		title:     "Compatibility test thread",
		createdAt: now,
		items: []storedChatKitThreadItem{{
			id:        itemID,
			text:      "Remember this compatibility test item.",
			createdAt: now,
		}},
	}
}

func (s *chatKitStore) createSession(user, workflowID string) storedChatKitSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSess++
	session := storedChatKitSession{
		id:         "cksess_mock_" + strconv.Itoa(s.nextSess),
		user:       user,
		workflowID: workflowID,
		status:     "active",
		createdAt:  time.Now().Unix(),
	}
	s.sessions[session.id] = session
	return session
}

func (s *chatKitStore) cancelSession(id string) (storedChatKitSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return storedChatKitSession{}, false
	}
	session.status = "cancelled"
	s.sessions[id] = session
	return session, true
}

func (s *chatKitStore) getThread(id string) (storedChatKitThread, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[id]
	if !ok {
		return storedChatKitThread{}, false
	}
	return cloneChatKitThread(thread), true
}

func (s *chatKitStore) listThreads(user string) []storedChatKitThread {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.threads))
	for id := range s.threads {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var items []storedChatKitThread
	for _, id := range ids {
		thread := s.threads[id]
		if user != "" && thread.user != user {
			continue
		}
		items = append(items, cloneChatKitThread(thread))
	}
	return items
}

func (s *chatKitStore) deleteThread(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.threads[id]; !ok {
		return false
	}
	delete(s.threads, id)
	return true
}

func cloneChatKitThread(thread storedChatKitThread) storedChatKitThread {
	cloned := thread
	cloned.items = slices.Clone(thread.items)
	return cloned
}