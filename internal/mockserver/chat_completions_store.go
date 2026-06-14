package mockserver

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type chatCompletionStore struct {
	mu          sync.Mutex
	nextID      int
	completions map[string]map[string]any
	messages    map[string][]map[string]any
}

func newChatCompletionStore() *chatCompletionStore {
	return &chatCompletionStore{
		completions: make(map[string]map[string]any),
		messages:    make(map[string][]map[string]any),
	}
}

func (s *chatCompletionStore) allocateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	return "chatcmpl-mock-" + strconv.Itoa(s.nextID)
}

func (s *chatCompletionStore) save(id string, payload map[string]any, messages []map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.completions[id] = payload
	if len(messages) > 0 {
		s.messages[id] = messages
	}
}

func (s *chatCompletionStore) get(id string) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, ok := s.completions[id]
	if !ok {
		return nil, false
	}
	return cloneMap(payload), true
}

func (s *chatCompletionStore) delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.completions[id]; !ok {
		return false
	}
	delete(s.completions, id)
	delete(s.messages, id)
	return true
}

func (s *chatCompletionStore) listAll() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]map[string]any, 0, len(s.completions))
	for _, payload := range s.completions {
		out = append(out, cloneMap(payload))
	}
	sort.Slice(out, func(i, j int) bool {
		idI, _ := out[i]["id"].(string)
		idJ, _ := out[j]["id"].(string)
		numI, _ := strconv.Atoi(strings.TrimPrefix(idI, "chatcmpl-mock-"))
		numJ, _ := strconv.Atoi(strings.TrimPrefix(idJ, "chatcmpl-mock-"))
		return numI < numJ
	})
	return out
}

func (s *chatCompletionStore) messagesFor(id string) ([]map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.completions[id]; !ok {
		return nil, false
	}
	items := s.messages[id]
	out := make([]map[string]any, len(items))
	for i, item := range items {
		out[i] = cloneMap(item)
	}
	return out, true
}

func chatMessagesFromRequest(messages []chatCompletionRequestMessage, assistantContent string, completionID string) []map[string]any {
	var out []map[string]any
	for i, msg := range messages {
		content, contentParts := chatMessageContentFromRaw(msg.Content)
		entry := map[string]any{
			"id":            completionID + "-msg-" + strconv.Itoa(i+1),
			"role":          msg.Role,
			"content":       content,
			"content_parts": contentParts,
		}
		out = append(out, entry)
	}
	out = append(out, map[string]any{
		"id":            completionID + "-msg-assistant",
		"role":          "assistant",
		"content":       assistantContent,
		"content_parts": nil,
	})
	return out
}

func chatMessageContentFromRaw(raw json.RawMessage) (string, any) {
	if len(raw) == 0 {
		return "", nil
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString, nil
	}
	var parts []map[string]any
	if err := json.Unmarshal(raw, &parts); err == nil && len(parts) > 0 {
		return "", parts
	}
	return "", nil
}

func mockChatCompletionPayload(id, content string) map[string]any {
	return map[string]any{
		"id":      id,
		"object":  "chat.completion",
		"created": 1700000000,
		"model":   "gpt-4o-mini",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     5,
			"completion_tokens": 1,
			"total_tokens":      6,
		},
	}
}