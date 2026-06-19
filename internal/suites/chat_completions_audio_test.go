package suites

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/beranekio/openai-compatibility-tester/internal/testutil"
	"github.com/openai/openai-go/v3"
)

func TestValidateChatCompletionAudio(t *testing.T) {
	validData := base64.StdEncoding.EncodeToString(testutil.SmallWAVBytes())

	tests := []struct {
		name    string
		fields  map[string]any
		wantErr bool
	}{
		{
			name: "valid audio",
			fields: map[string]any{
				"id":         "audio-1",
				"data":       validData,
				"expires_at": 1700003600,
				"transcript": "pong",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			fields: map[string]any{
				"data":       validData,
				"expires_at": 1700003600,
				"transcript": "pong",
			},
			wantErr: true,
		},
		{
			name: "empty data",
			fields: map[string]any{
				"id":         "audio-1",
				"data":       "",
				"expires_at": 1700003600,
				"transcript": "pong",
			},
			wantErr: true,
		},
		{
			name: "missing expires_at",
			fields: map[string]any{
				"id":         "audio-1",
				"data":       validData,
				"transcript": "pong",
			},
			wantErr: true,
		},
		{
			name: "missing transcript",
			fields: map[string]any{
				"id":         "audio-1",
				"data":       validData,
				"expires_at": 1700003600,
			},
			wantErr: true,
		},
		{
			name: "invalid base64",
			fields: map[string]any{
				"id":         "audio-1",
				"data":       "not-base64!!!",
				"expires_at": 1700003600,
				"transcript": "pong",
			},
			wantErr: true,
		},
		{
			name: "non-wav payload",
			fields: map[string]any{
				"id":         "audio-1",
				"data":       "YQ==",
				"expires_at": 1700003600,
				"transcript": "pong",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audio := unmarshalChatCompletionAudio(t, tt.fields)
			err := validateChatCompletionAudio("chat_completions_audio", audio)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateChatCompletionAudio() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasChatCompletionAudioOutput(t *testing.T) {
	audio := unmarshalChatCompletionMessage(t, map[string]any{
		"role":    "assistant",
		"content": nil,
		"audio": map[string]any{
			"id":         "audio-1",
			"data":       base64.StdEncoding.EncodeToString(testutil.SmallWAVBytes()),
			"expires_at": 1700003600,
			"transcript": "pong",
		},
	})
	if !hasChatCompletionAudioOutput(audio) {
		t.Fatal("expected audio-only message to count as output")
	}
	if hasChatMessageOutput(audio) {
		t.Fatal("expected audio-only message to have no text output")
	}
}

func unmarshalChatCompletionMessage(t *testing.T, fields map[string]any) openai.ChatCompletionMessage {
	t.Helper()
	payload, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal message fields: %v", err)
	}
	var msg openai.ChatCompletionMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	return msg
}

func unmarshalChatCompletionAudio(t *testing.T, fields map[string]any) openai.ChatCompletionAudio {
	t.Helper()
	payload, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal audio fields: %v", err)
	}
	var audio openai.ChatCompletionAudio
	if err := json.Unmarshal(payload, &audio); err != nil {
		t.Fatalf("unmarshal audio: %v", err)
	}
	return audio
}
