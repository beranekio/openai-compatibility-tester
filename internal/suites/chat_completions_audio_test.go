package suites

import (
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestValidateChatCompletionAudio(t *testing.T) {
	tests := []struct {
		name    string
		audio   openai.ChatCompletionAudio
		wantErr bool
	}{
		{
			name:    "valid audio",
			audio:   openai.ChatCompletionAudio{ID: "audio-1", Data: "YQ=="},
			wantErr: false,
		},
		{
			name:    "missing id",
			audio:   openai.ChatCompletionAudio{Data: "YQ=="},
			wantErr: true,
		},
		{
			name:    "empty data",
			audio:   openai.ChatCompletionAudio{ID: "audio-1", Data: ""},
			wantErr: true,
		},
		{
			name:    "invalid base64",
			audio:   openai.ChatCompletionAudio{ID: "audio-1", Data: "not-base64!!!"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChatCompletionAudio("chat_completions_audio", tt.audio)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateChatCompletionAudio() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}