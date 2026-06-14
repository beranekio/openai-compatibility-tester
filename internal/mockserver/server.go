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
	store     *responseStore
	chatStore *chatCompletionStore
}

// New starts a mock OpenAI API server.
func New() *Server {
	mux := http.NewServeMux()
	s := &Server{
		store:     newResponseStore(),
		chatStore: newChatCompletionStore(),
	}

	mux.HandleFunc("GET /v1/models", handleModels)
	mux.HandleFunc("GET /v1/models/{id}", handleModelGet)
	mux.HandleFunc("POST /v1/chat/completions", s.handleChatCompletions)
	mux.HandleFunc("GET /v1/chat/completions", s.handleChatCompletionList)
	mux.HandleFunc("GET /v1/chat/completions/{id}", s.handleChatCompletionGet)
	mux.HandleFunc("DELETE /v1/chat/completions/{id}", s.handleChatCompletionDelete)
	mux.HandleFunc("GET /v1/chat/completions/{id}/messages", s.handleChatCompletionMessages)
	mux.HandleFunc("POST /v1/completions", handleCompletions)
	mux.HandleFunc("POST /v1/embeddings", handleEmbeddings)
	mux.HandleFunc("POST /v1/responses", s.handleResponses)
	mux.HandleFunc("GET /v1/responses/{id}", s.handleResponseGet)
	mux.HandleFunc("DELETE /v1/responses/{id}", s.handleResponseDelete)
	mux.HandleFunc("POST /v1/responses/{id}/cancel", s.handleResponseCancel)
	mux.HandleFunc("GET /v1/responses/{id}/input_items", s.handleResponseInputItems)
	mux.HandleFunc("POST /v1/responses/compact", handleResponseCompact)
	mux.HandleFunc("POST /v1/responses/input_tokens", handleResponseInputTokens)
	mux.HandleFunc("POST /v1/moderations", handleModerations)
	mux.HandleFunc("POST /v1/images/generations", handleImagesGenerations)
	mux.HandleFunc("POST /v1/images/edits", handleImagesEdits)
	mux.HandleFunc("POST /v1/images/variations", handleImagesVariations)
	mux.HandleFunc("POST /v1/audio/speech", handleAudioSpeech)
	mux.HandleFunc("POST /v1/audio/transcriptions", handleAudioTranscriptions)
	mux.HandleFunc("POST /v1/audio/translations", handleAudioTranslations)

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

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Model           string `json:"model"`
		Stream          bool   `json:"stream"`
		Store           *bool  `json:"store"`
		ReasoningEffort string `json:"reasoning_effort"`
		ResponseFormat  *struct {
			Type string `json:"type"`
		} `json:"response_format"`
		Modalities []string                       `json:"modalities"`
		Messages   []chatCompletionRequestMessage `json:"messages"`
		Tools []json.RawMessage `json:"tools"`
	}
	_ = json.Unmarshal(body, &req)

	if len(req.Tools) > 0 && !chatCompletionRequestIsMultiTurn(req.Messages) {
		if req.Stream {
			writeChatCompletionToolCallStream(w)
			return
		}
		writeChatCompletionToolCallResponse(w)
		return
	}

	if chatCompletionRequestHasAudioModalities(req.Modalities) {
		writeChatCompletionAudioResponse(w)
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
	if chatCompletionRequestIsMultiTurn(req.Messages) {
		content = "72"
	} else if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_schema" {
		content = `{"answer":"pong"}`
	} else if chatCompletionRequestHasImageURL(req.Messages) {
		content = "I see an image"
	}

	usage := map[string]any{
		"prompt_tokens":     5,
		"completion_tokens": 1,
		"total_tokens":      6,
	}
	if req.ReasoningEffort != "" || isReasoningChatModel(req.Model) {
		usage["completion_tokens_details"] = map[string]any{
			"reasoning_tokens": 3,
		}
	}

	id := "chatcmpl-mock"
	if req.Store != nil && *req.Store {
		id = s.chatStore.allocateID()
	}

	payload := map[string]any{
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
		"usage": usage,
	}
	if req.Store != nil && *req.Store {
		s.chatStore.save(id, payload, chatMessagesFromRequest(req.Messages, content, id))
	}
	writeJSON(w, payload)
}

func chatCompletionRequestHasAudioModalities(modalities []string) bool {
	for _, modality := range modalities {
		if modality == "audio" {
			return true
		}
	}
	return false
}

func writeChatCompletionAudioResponse(w http.ResponseWriter) {
	writeJSON(w, map[string]any{
		"id":      "chatcmpl-mock-audio",
		"object":  "chat.completion",
		"created": 1700000000,
		"model":   "gpt-4o-mini",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": nil,
					"audio": map[string]any{
						"id":         "audio-mock",
						"data":       mockChatCompletionWAVBase64(),
						"expires_at": 1700003600,
						"transcript": "pong",
					},
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

type chatCompletionRequestMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

func chatCompletionRequestIsMultiTurn(messages []chatCompletionRequestMessage) bool {
	for _, msg := range messages {
		if msg.Role == "developer" || msg.Role == "tool" {
			return true
		}
	}
	return false
}

func chatCompletionRequestHasImageURL(messages []chatCompletionRequestMessage) bool {
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

func isReasoningChatModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o4")
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
		"type":         "response.output_item.added",
		"output_index": 0,
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
			"type":         "response.function_call_arguments.delta",
			"item_id":      "fc-mock",
			"output_index": 0,
			"delta":        chunk,
		})
	}
	writeEvent(map[string]any{
		"type":         "response.function_call_arguments.done",
		"item_id":      "fc-mock",
		"output_index": 0,
		"name":         "get_weather",
		"arguments":    `{"location":"San Francisco, CA"}`,
	})
	writeEvent(map[string]any{
		"type":         "response.output_item.done",
		"output_index": 0,
		"item": map[string]any{
			"id":        "fc-mock",
			"type":      "function_call",
			"status":    "completed",
			"call_id":   "call_mock_weather",
			"name":      "get_weather",
			"arguments": `{"location":"San Francisco, CA"}`,
		},
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
