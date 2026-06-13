package mockserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func handleResponseCompact(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Input json.RawMessage `json:"input"`
	}
	_ = json.Unmarshal(body, &req)

	output := make([]map[string]any, 0, len(userMessagesFromCompactInput(req.Input))+1)
	for i, text := range userMessagesFromCompactInput(req.Input) {
		output = append(output, map[string]any{
			"id":     fmt.Sprintf("msg-user-%d", i+1),
			"type":   "message",
			"role":   "user",
			"status": "completed",
			"content": []map[string]any{{
				"type": "input_text",
				"text": text,
			}},
		})
	}
	output = append(output, map[string]any{
		"id":                 "compact-mock",
		"type":               "compaction",
		"encrypted_content": "encrypted-summary",
	})

	writeJSON(w, map[string]any{
		"id":         "resp-compact-mock",
		"object":     "response.compaction",
		"created_at": 1700000000,
		"output":     output,
		"usage": map[string]any{
			"input_tokens":  10,
			"output_tokens": 5,
			"total_tokens":  15,
			"input_tokens_details": map[string]any{
				"cached_tokens": 0,
			},
			"output_tokens_details": map[string]any{
				"reasoning_tokens": 0,
			},
		},
	})
}

func userMessagesFromCompactInput(input json.RawMessage) []string {
	if len(input) == 0 {
		return []string{"Summarize this conversation."}
	}
	var asString string
	if err := json.Unmarshal(input, &asString); err == nil && asString != "" {
		return []string{asString}
	}

	var items []json.RawMessage
	if err := json.Unmarshal(input, &items); err != nil {
		return nil
	}

	texts := make([]string, 0, len(items))
	for _, raw := range items {
		var item struct {
			Type    string          `json:"type"`
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}
		if item.Role != "user" {
			continue
		}
		if item.Type != "" && item.Type != "message" {
			continue
		}
		var contentString string
		if err := json.Unmarshal(item.Content, &contentString); err == nil && contentString != "" {
			texts = append(texts, contentString)
			continue
		}
		var contentItems []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(item.Content, &contentItems); err != nil {
			continue
		}
		for _, part := range contentItems {
			if part.Text != "" {
				texts = append(texts, part.Text)
				break
			}
		}
	}
	return texts
}

func handleResponseInputTokens(w http.ResponseWriter, r *http.Request) {
	_, _ = io.ReadAll(r.Body)
	writeJSON(w, map[string]any{
		"object":       "response.input_tokens",
		"input_tokens": 12,
	})
}