package mockserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Server provides a minimal OpenAI-compatible HTTP API for CI tests.
type Server struct {
	*httptest.Server
}

// New starts a mock OpenAI API server.
func New() *Server {
	mux := http.NewServeMux()
	s := &Server{}

	mux.HandleFunc("GET /v1/models", handleModels)
	mux.HandleFunc("GET /v1/models/{id}", handleModelGet)
	mux.HandleFunc("POST /v1/chat/completions", handleChatCompletions)
	mux.HandleFunc("POST /v1/completions", handleCompletions)
	mux.HandleFunc("POST /v1/embeddings", handleEmbeddings)
	mux.HandleFunc("POST /v1/responses", handleResponses)

	s.Server = httptest.NewServer(mux)
	return s
}

// BaseURL returns the API base URL including the /v1 prefix.
func (s *Server) BaseURL() string {
	return s.URL + "/v1"
}

func mockModel() map[string]any {
	return map[string]any{
		"id":       "gpt-4o-mini",
		"object":   "model",
		"created":  1700000000,
		"owned_by": "mock",
	}
}

func handleModels(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"object": "list",
		"data":   []map[string]any{mockModel()},
	})
}

func handleModelGet(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, mockModel())
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Stream         bool `json:"stream"`
		ResponseFormat *struct {
			Type string `json:"type"`
		} `json:"response_format"`
		Messages []struct {
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
		Tools []json.RawMessage `json:"tools"`
	}
	_ = json.Unmarshal(body, &req)

	if len(req.Tools) > 0 {
		if req.Stream {
			writeChatCompletionToolCallStream(w)
			return
		}
		writeChatCompletionToolCallResponse(w)
		return
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		chunks := []string{"one", " two", " three"}
		for _, chunk := range chunks {
			payload := map[string]any{
				"id":      "chatcmpl-mock",
				"object":  "chat.completion.chunk",
				"created": 1700000000,
				"model":   "gpt-4o-mini",
				"choices": []map[string]any{
					{
						"index": 0,
						"delta": map[string]any{"content": chunk},
					},
				},
			}
			data, _ := json.Marshal(payload)
			_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		}
		final, _ := json.Marshal(map[string]any{
			"id":      "chatcmpl-mock",
			"object":  "chat.completion.chunk",
			"created": 1700000000,
			"model":   "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"index":         0,
					"delta":         map[string]any{},
					"finish_reason": "stop",
				},
			},
		})
		_, _ = w.Write([]byte("data: " + string(final) + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		return
	}

	content := "pong"
	if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_schema" {
		content = `{"answer":"pong"}`
	} else if chatCompletionRequestHasImageURL(req.Messages) {
		content = "I see an image"
	}

	writeJSON(w, map[string]any{
		"id":      "chatcmpl-mock",
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
	})
}

func chatCompletionRequestHasImageURL(messages []struct {
	Content json.RawMessage `json:"content"`
}) bool {
	for _, msg := range messages {
		var parts []struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg.Content, &parts); err != nil {
			continue
		}
		for _, part := range parts {
			if part.Type == "image_url" {
				return true
			}
		}
	}
	return false
}

func handleCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Stream bool `json:"stream"`
	}
	_ = json.Unmarshal(body, &req)

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		chunks := []string{"hel", "lo"}
		for _, chunk := range chunks {
			payload := map[string]any{
				"id":      "cmpl-mock",
				"object":  "text_completion",
				"created": 1700000000,
				"model":   "gpt-4o-mini",
				"choices": []map[string]any{
					{
						"index": 0,
						"text":  chunk,
					},
				},
			}
			data, _ := json.Marshal(payload)
			_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		}
		final, _ := json.Marshal(map[string]any{
			"id":      "cmpl-mock",
			"object":  "text_completion",
			"created": 1700000000,
			"model":   "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"index":         0,
					"text":          "",
					"finish_reason": "stop",
				},
			},
		})
		_, _ = w.Write([]byte("data: " + string(final) + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		return
	}
	writeJSON(w, map[string]any{
		"id":      "cmpl-mock",
		"object":  "text_completion",
		"created": 1700000000,
		"model":   "gpt-4o-mini",
		"choices": []map[string]any{
			{
				"index":         0,
				"text":          "hello",
				"finish_reason": "stop",
			},
		},
	})
}

func handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Input json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := []map[string]any{
		{
			"object":    "embedding",
			"index":     0,
			"embedding": []float64{0.1, 0.2, 0.3},
		},
	}
	promptTokens := 3
	if len(req.Input) > 0 && req.Input[0] == '[' {
		data = append(data, map[string]any{
			"object":    "embedding",
			"index":     1,
			"embedding": []float64{0.4, 0.5, 0.6},
		})
		promptTokens = 6
	}

	writeJSON(w, map[string]any{
		"object": "list",
		"data":   data,
		"model":  "text-embedding-3-small",
		"usage": map[string]any{
			"prompt_tokens": promptTokens,
			"total_tokens":  promptTokens,
		},
	})
}

func handleResponses(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Stream bool              `json:"stream"`
		Tools  []json.RawMessage `json:"tools"`
	}
	_ = json.Unmarshal(body, &req)

	if len(req.Tools) > 0 {
		if req.Stream {
			writeResponsesToolCallStream(w)
			return
		}
		writeResponsesToolCallResponse(w)
		return
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		seq := 0
		writeResponseStreamEvent := func(payload map[string]any) {
			payload["sequence_number"] = seq
			seq++
			data, _ := json.Marshal(payload)
			_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		}

		writeResponseStreamEvent(map[string]any{
			"type": "response.created",
			"response": map[string]any{
				"id":         "resp-mock",
				"object":     "response",
				"status":     "in_progress",
				"model":      "gpt-4o-mini",
				"created_at": 1700000000,
			},
		})
		writeResponseStreamEvent(map[string]any{
			"type": "response.in_progress",
			"response": map[string]any{
				"id":         "resp-mock",
				"object":     "response",
				"status":     "in_progress",
				"model":      "gpt-4o-mini",
				"created_at": 1700000000,
			},
		})
		writeResponseStreamEvent(map[string]any{
			"type":          "response.output_item.added",
			"output_index":  0,
			"item": map[string]any{
				"id":     "msg-mock",
				"type":   "message",
				"role":   "assistant",
				"status": "in_progress",
			},
		})
		writeResponseStreamEvent(map[string]any{
			"type":          "response.content_part.added",
			"item_id":       "msg-mock",
			"output_index":  0,
			"content_index": 0,
			"part": map[string]any{
				"type": "output_text",
				"text": "",
			},
		})

		chunks := []string{"one", " two", " three"}
		for _, chunk := range chunks {
			writeResponseStreamEvent(map[string]any{
				"type":          "response.output_text.delta",
				"content_index": 0,
				"item_id":       "msg-mock",
				"output_index":  0,
				"logprobs":      []any{},
				"delta":         chunk,
			})
		}
		writeResponseStreamEvent(map[string]any{
			"type": "response.completed",
			"response": map[string]any{
				"id":         "resp-mock",
				"object":     "response",
				"status":     "completed",
				"model":      "gpt-4o-mini",
				"created_at": 1700000000,
			},
		})
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		return
	}

	writeJSON(w, map[string]any{
		"id":         "resp-mock",
		"object":     "response",
		"status":     "completed",
		"model":      "gpt-4o-mini",
		"created_at": 1700000000,
		"output": []map[string]any{
			{
				"id":     "msg-mock",
				"type":   "message",
				"role":   "assistant",
				"status": "completed",
				"content": []map[string]any{
					{
						"type": "output_text",
						"text": "pong",
					},
				},
			},
		},
	})
}

func writeResponsesToolCallResponse(w http.ResponseWriter) {
	writeJSON(w, map[string]any{
		"id":         "resp-mock-tools",
		"object":     "response",
		"status":     "completed",
		"model":      "gpt-4o-mini",
		"created_at": 1700000000,
		"output": []map[string]any{
			{
				"id":        "fc-mock",
				"type":      "function_call",
				"status":    "completed",
				"call_id":   "call_mock_weather",
				"name":      "get_weather",
				"arguments": `{"location":"San Francisco, CA"}`,
			},
		},
	})
}

func writeResponsesToolCallStream(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	seq := 0
	writeEvent := func(payload map[string]any) {
		payload["sequence_number"] = seq
		seq++
		data, _ := json.Marshal(payload)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
	}

	writeEvent(map[string]any{
		"type":          "response.output_item.added",
		"output_index":  0,
		"item": map[string]any{
			"id":        "fc-mock",
			"type":      "function_call",
			"status":    "in_progress",
			"call_id":   "call_mock_weather",
			"name":      "get_weather",
			"arguments": "",
		},
	})

	argChunks := []string{`{"location"`, `:"San Francisco, CA"`, `}`}
	for _, chunk := range argChunks {
		writeEvent(map[string]any{
			"type":          "response.function_call_arguments.delta",
			"item_id":       "fc-mock",
			"output_index":  0,
			"delta":         chunk,
		})
	}
	writeEvent(map[string]any{
		"type":          "response.function_call_arguments.done",
		"item_id":       "fc-mock",
		"output_index":  0,
		"name":          "get_weather",
		"arguments":     `{"location":"San Francisco, CA"}`,
	})
	writeEvent(map[string]any{
		"type": "response.completed",
		"response": map[string]any{
			"id":         "resp-mock-tools",
			"object":     "response",
			"status":     "completed",
			"model":      "gpt-4o-mini",
			"created_at": 1700000000,
		},
	})
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
}

func writeChatCompletionToolCallResponse(w http.ResponseWriter) {
	writeJSON(w, map[string]any{
		"id":      "chatcmpl-mock-tools",
		"object":  "chat.completion",
		"created": 1700000000,
		"model":   "gpt-4o-mini",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role": "assistant",
					"tool_calls": []map[string]any{
						{
							"id":   "call_mock_weather",
							"type": "function",
							"function": map[string]any{
								"name":      "get_weather",
								"arguments": `{"location":"San Francisco, CA"}`,
							},
						},
					},
				},
				"finish_reason": "tool_calls",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     12,
			"completion_tokens": 18,
			"total_tokens":      30,
		},
	})
}

func writeChatCompletionToolCallStream(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	chunks := []map[string]any{
		{
			"id":      "chatcmpl-mock-tools",
			"object":  "chat.completion.chunk",
			"created": 1700000000,
			"model":   "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"role": "assistant",
						"tool_calls": []map[string]any{
							{
								"index": 0,
								"id":    "call_mock_weather",
								"type":  "function",
								"function": map[string]any{
									"name": "get_weather",
								},
							},
						},
					},
				},
			},
		},
		{
			"id":      "chatcmpl-mock-tools",
			"object":  "chat.completion.chunk",
			"created": 1700000000,
			"model":   "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"tool_calls": []map[string]any{
							{
								"index": 0,
								"function": map[string]any{
									"arguments": `{"location":"San Francisco, CA"}`,
								},
							},
						},
					},
				},
			},
		},
		{
			"id":      "chatcmpl-mock-tools",
			"object":  "chat.completion.chunk",
			"created": 1700000000,
			"model":   "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"index":         0,
					"delta":         map[string]any{},
					"finish_reason": "tool_calls",
				},
			},
		},
	}

	for _, payload := range chunks {
		data, _ := json.Marshal(payload)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
	}
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

// BrokenServer returns a server that responds with invalid payloads.
func BrokenServer() *Server {
	mux := http.NewServeMux()
	s := &Server{}

	mux.HandleFunc("GET /v1/models", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, strings.TrimSpace(`{"object":"list","data":[]}`))
	})
	mux.HandleFunc("POST /v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"message":"incompatible"}}`, http.StatusBadRequest)
	})

	s.Server = httptest.NewServer(mux)
	return s
}