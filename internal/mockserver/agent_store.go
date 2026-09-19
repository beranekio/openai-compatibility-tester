package mockserver

import (
	"sort"
	"strconv"
	"sync"
)

type storedAgent struct {
	id           string
	model        string
	name         string
	instructions string
	metadata     map[string]string
	createdAt    int64
	updatedAt    int64
}

type storedAgentSession struct {
	id        string
	agentID   string
	model     string
	name      string
	metadata  map[string]string
	createdAt int64
}

type agentStore struct {
	mu       sync.Mutex
	next     int
	agents   map[string]storedAgent
	sessions map[string]storedAgentSession
}

func newAgentStore() *agentStore {
	return &agentStore{
		agents:   make(map[string]storedAgent),
		sessions: make(map[string]storedAgentSession),
	}
}

func (s *agentStore) create(model, name, instructions string, metadata map[string]string) storedAgent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	agent := storedAgent{
		id:           "agt_mock_" + strconv.Itoa(s.next),
		model:        model,
		name:         name,
		instructions: instructions,
		metadata:     cloneStringMap(metadata),
		createdAt:    1700000000,
		updatedAt:    1700000000,
	}
	s.agents[agent.id] = agent
	return cloneStoredAgent(agent)
}

func (s *agentStore) get(id string) (storedAgent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[id]
	if !ok {
		return storedAgent{}, false
	}
	return cloneStoredAgent(agent), true
}

func (s *agentStore) update(id string, name, instructions *string, metadata map[string]string) (storedAgent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[id]
	if !ok {
		return storedAgent{}, false
	}
	if name != nil {
		agent.name = *name
	}
	if instructions != nil {
		agent.instructions = *instructions
	}
	if metadata != nil {
		agent.metadata = cloneStringMap(metadata)
	}
	agent.updatedAt = 1700000001
	s.agents[id] = agent
	return cloneStoredAgent(agent), true
}

func (s *agentStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[id]; !ok {
		return false
	}
	delete(s.agents, id)
	return true
}

func (s *agentStore) list() []storedAgent {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.agents))
	for id := range s.agents {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	items := make([]storedAgent, len(ids))
	for i, id := range ids {
		items[i] = cloneStoredAgent(s.agents[id])
	}
	return items
}

func (s *agentStore) createSession(agentID, model, name string, metadata map[string]string) storedAgentSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	session := storedAgentSession{
		id:        "sess_mock_" + strconv.Itoa(s.next),
		agentID:   agentID,
		model:     model,
		name:      name,
		metadata:  cloneStringMap(metadata),
		createdAt: 1700000000,
	}
	s.sessions[session.id] = session
	return cloneStoredAgentSession(session)
}

func (s *agentStore) getSession(id string) (storedAgentSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return storedAgentSession{}, false
	}
	return cloneStoredAgentSession(session), true
}

func (s *agentStore) deleteSession(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return false
	}
	delete(s.sessions, id)
	return true
}

func (s *agentStore) listSessions() []storedAgentSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.sessions))
	for id := range s.sessions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	items := make([]storedAgentSession, len(ids))
	for i, id := range ids {
		items[i] = cloneStoredAgentSession(s.sessions[id])
	}
	return items
}

func cloneStoredAgent(agent storedAgent) storedAgent {
	agent.metadata = cloneStringMap(agent.metadata)
	return agent
}

func cloneStoredAgentSession(session storedAgentSession) storedAgentSession {
	session.metadata = cloneStringMap(session.metadata)
	return session
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
