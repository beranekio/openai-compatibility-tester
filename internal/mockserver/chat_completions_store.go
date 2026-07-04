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

// updateMetadata merges the given metadata into the stored completion payload
// and returns a clone of the updated payload. It builds a fresh metadata map
// rather than mutating the existing one in place, so previously returned
// (shallow-cloned) payloads keep their nested metadata untouched and free of
// data races.
func (s *chatCompletionStore) updateMetadata(id string, metadata map[string]string) (map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, ok := s.completions[id]
	if !ok {
		return nil, false
	}
	merged := make(map[string]string, len(metadata))
	switch existing := payload["metadata"].(type) {
	case map[string]string:
		for k, v := range existing {
			merged[k] = v
		}
	case map[string]any:
		for k, v := range existing {
			if s, ok := v.(string); ok {
				merged[k] = s
			}
		}
	}
	for k, v := range metadata {
		merged[k] = v
	}
	payload["metadata"] = merged
	return cloneMap(payload), true
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