package mockserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (s *Server) handleConversationCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Metadata map[string]any    `json:"metadata"`
		Items    []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conversation := s.conversationStore.create(req.Metadata)
	if len(req.Items) > 0 {
		s.conversationStore.addItems(conversation.id, conversationItemRequestsFromRaw(req.Items))
		conversation, _ = s.conversationStore.get(conversation.id)
	}
	writeJSON(w, conversationPayload(conversation))
}

func (s *Server) handleConversationGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	conversation, ok := s.conversationStore.get(id)
	if !ok {
		writeNotFound(w, "Conversation not found", "conversation_id")
		return
	}
	writeJSON(w, conversationPayload(conversation))
}

func (s *Server) handleConversationUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conversation, ok := s.conversationStore.update(id, req.Metadata)
	if !ok {
		writeNotFound(w, "Conversation not found", "conversation_id")
		return
	}
	writeJSON(w, conversationPayload(conversation))
}

func (s *Server) handleConversationDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.conversationStore.delete(id) {
		writeNotFound(w, "Conversation not found", "conversation_id")
		return
	}
	writeJSON(w, map[string]any{
		"id":      id,
		"object":  "conversation.deleted",
		"deleted": true,
	})
}

func (s *Server) handleConversationItemCreate(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("id")
	var req struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	items, ok := s.conversationStore.addItems(conversationID, conversationItemRequestsFromRaw(req.Items))
	if !ok {
		writeNotFound(w, "Conversation not found", "conversation_id")
		return
	}
	writeConversationItemList(w, items)
}

func (s *Server) handleConversationItemList(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("id")
	items, ok := s.conversationStore.listItems(conversationID)
	if !ok {
		writeNotFound(w, "Conversation not found", "conversation_id")
		return
	}
	writeConversationItemList(w, items)
}

func (s *Server) handleConversationItemGet(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("id")
	itemID := r.PathValue("itemID")
	item, ok := s.conversationStore.getItem(conversationID, itemID)
	if !ok {
		writeNotFound(w, "Conversation item not found", "item_id")
		return
	}
	writeJSON(w, conversationItemPayload(item))
}

func (s *Server) handleConversationItemDelete(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("id")
	itemID := r.PathValue("itemID")
	conversation, ok := s.conversationStore.deleteItem(conversationID, itemID)
	if !ok {
		writeNotFound(w, "Conversation item not found", "item_id")
		return
	}
	writeJSON(w, conversationPayload(conversation))
}

func writeConversationItemList(w http.ResponseWriter, items []storedConversationItem) {
	data := make([]map[string]any, len(items))
	firstID := ""
	lastID := ""
	for i, item := range items {
		data[i] = conversationItemPayload(item)
		if i == 0 {
			firstID = item.id
		}
		lastID = item.id
	}
	writeJSON(w, map[string]any{
		"object":   "list",
		"data":     data,
		"first_id": firstID,
		"last_id":  lastID,
		"has_more": false,
	})
}
