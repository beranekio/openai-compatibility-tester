package mockserver

import (
	"encoding/json"
	"sort"
	"strconv"
	"sync"
)

type storedConversation struct {
	id        string
	metadata  map[string]any
	createdAt int64
	items     map[string]storedConversationItem
}

type storedConversationItem struct {
	id      string
	role    string
	text    string
	status  string
	created int
}

type conversationStore struct {
	mu            sync.Mutex
	next          int
	nextItem      int
	conversations map[string]storedConversation
}

func newConversationStore() *conversationStore {
	return &conversationStore{
		conversations: make(map[string]storedConversation),
	}
}

func (s *conversationStore) create(metadata map[string]any) storedConversation {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	conversation := storedConversation{
		id:        "conv_mock_" + strconv.Itoa(s.next),
		metadata:  cloneMap(metadata),
		createdAt: 1700000000,
		items:     make(map[string]storedConversationItem),
	}
	s.conversations[conversation.id] = conversation
	return cloneConversation(conversation)
}

func (s *conversationStore) get(id string) (storedConversation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[id]
	if !ok {
		return storedConversation{}, false
	}
	return cloneConversation(conversation), true
}

func (s *conversationStore) update(id string, metadata map[string]any) (storedConversation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[id]
	if !ok {
		return storedConversation{}, false
	}
	conversation.metadata = cloneMap(metadata)
	s.conversations[id] = conversation
	return cloneConversation(conversation), true
}

func (s *conversationStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.conversations[id]; !ok {
		return false
	}
	delete(s.conversations, id)
	return true
}

func (s *conversationStore) addItems(conversationID string, requests []conversationItemRequest) ([]storedConversationItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		return nil, false
	}
	items := make([]storedConversationItem, 0, len(requests))
	for _, req := range requests {
		s.nextItem++
		item := storedConversationItem{
			id:      "msg_mock_" + strconv.Itoa(s.nextItem),
			role:    req.role,
			text:    req.text,
			status:  "completed",
			created: s.nextItem,
		}
		if item.role == "" {
			item.role = "user"
		}
		conversation.items[item.id] = item
		items = append(items, item)
	}
	s.conversations[conversationID] = conversation
	return cloneConversationItems(items), true
}

func (s *conversationStore) getItem(conversationID, itemID string) (storedConversationItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		return storedConversationItem{}, false
	}
	item, ok := conversation.items[itemID]
	if !ok {
		return storedConversationItem{}, false
	}
	return item, true
}

func (s *conversationStore) listItems(conversationID string) ([]storedConversationItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		return nil, false
	}
	items := make([]storedConversationItem, 0, len(conversation.items))
	for _, item := range conversation.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].created < items[j].created
	})
	return items, true
}

func (s *conversationStore) deleteItem(conversationID, itemID string) (storedConversation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conversation, ok := s.conversations[conversationID]
	if !ok {
		return storedConversation{}, false
	}
	if _, ok := conversation.items[itemID]; !ok {
		return storedConversation{}, false
	}
	delete(conversation.items, itemID)
	s.conversations[conversationID] = conversation
	return cloneConversation(conversation), true
}

func cloneConversation(conversation storedConversation) storedConversation {
	conversation.metadata = cloneMap(conversation.metadata)
	conversation.items = cloneConversationItemMap(conversation.items)
	return conversation
}

func cloneConversationItemMap(src map[string]storedConversationItem) map[string]storedConversationItem {
	dst := make(map[string]storedConversationItem, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneConversationItems(src []storedConversationItem) []storedConversationItem {
	return append([]storedConversationItem(nil), src...)
}

func conversationPayload(conversation storedConversation) map[string]any {
	return map[string]any{
		"id":         conversation.id,
		"object":     "conversation",
		"created_at": conversation.createdAt,
		"metadata":   cloneMap(conversation.metadata),
	}
}

func conversationItemPayload(item storedConversationItem) map[string]any {
	return map[string]any{
		"id":     item.id,
		"type":   "message",
		"role":   item.role,
		"status": item.status,
		"content": []map[string]any{{
			"type": "input_text",
			"text": item.text,
		}},
	}
}

type conversationItemRequest struct {
	role string
	text string
}

func conversationItemRequestsFromRaw(rawItems []json.RawMessage) []conversationItemRequest {
	requests := make([]conversationItemRequest, 0, len(rawItems))
	for _, raw := range rawItems {
		var req struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			continue
		}
		requests = append(requests, conversationItemRequest{
			role: req.Role,
			text: conversationItemTextFromRaw(req.Content),
		})
	}
	return requests
}

func conversationItemTextFromRaw(raw json.RawMessage) string {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	var content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &content); err == nil {
		for _, item := range content {
			if item.Text != "" {
				return item.Text
			}
		}
	}
	return ""
}
