package mockserver

import (
	"io"
	"net/http"
)

func handleResponseCompact(w http.ResponseWriter, r *http.Request) {
	_, _ = io.ReadAll(r.Body)
	writeJSON(w, map[string]any{
		"id":         "resp-compact-mock",
		"object":     "response.compaction",
		"created_at": 1700000000,
		"output": []map[string]any{
			{
				"id":                 "compact-mock",
				"type":               "compaction",
				"encrypted_content": "encrypted-summary",
			},
		},
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

func handleResponseInputTokens(w http.ResponseWriter, r *http.Request) {
	_, _ = io.ReadAll(r.Body)
	writeJSON(w, map[string]any{
		"object":       "response.input_tokens",
		"input_tokens": 12,
	})
}