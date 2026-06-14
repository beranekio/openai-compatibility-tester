package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsMessages verifies GET /v1/chat/completions/{id}/messages.
type ChatCompletionsMessages struct{}

func (ChatCompletionsMessages) Name() string { return "chat_completions_messages" }
func (ChatCompletionsMessages) Description() string {
	return "Chat completion messages (POST /v1/chat/completions with store, then GET /v1/chat/completions/{id}/messages)"
}

func (ChatCompletionsMessages) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredChatCompletion(ctx, client, cfg)
	if err != nil {
		return err
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = client.Chat.Completions.Delete(cleanupCtx, created.ID)
	}()

	page, err := client.Chat.Completions.Messages.List(ctx, created.ID, openai.ChatCompletionMessageListParams{})
	if err != nil {
		return fmt.Errorf("chat completion messages list failed: %w", err)
	}
	if page == nil {
		return fail("chat_completions_messages", "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail("chat_completions_messages", "list missing has_more")
	}
	if page.HasMore {
		return fail("chat_completions_messages", "list has_more is true, want false for stored completion")
	}
	var envelope struct {
		Object  string `json:"object"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail("chat_completions_messages", "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail("chat_completions_messages", fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	if len(page.Data) == 0 {
		return fail("chat_completions_messages", "list data is empty")
	}
	if envelope.FirstID == "" {
		return fail("chat_completions_messages", "list missing first_id")
	}
	if envelope.LastID == "" {
		return fail("chat_completions_messages", "list missing last_id")
	}
	if envelope.FirstID != page.Data[0].ID {
		return fail("chat_completions_messages", fmt.Sprintf("list first_id is %q, want %q", envelope.FirstID, page.Data[0].ID))
	}
	if envelope.LastID != page.Data[len(page.Data)-1].ID {
		return fail("chat_completions_messages", fmt.Sprintf("list last_id is %q, want %q", envelope.LastID, page.Data[len(page.Data)-1].ID))
	}
	for _, msg := range page.Data {
		if msg.ID == "" {
			return fail("chat_completions_messages", "message missing id")
		}
		if string(msg.Role) == "" {
			return fail("chat_completions_messages", "message missing role")
		}
	}
	if !listContainsStoredChatInput(page.Data, storedChatCompletionInput) {
		return fail("chat_completions_messages", "list missing user message matching submitted prompt")
	}
	return nil
}

func listContainsStoredChatInput(messages []openai.ChatCompletionStoreMessage, want string) bool {
	for _, msg := range messages {
		if string(msg.Role) != "user" {
			continue
		}
		if msg.Content == want {
			return true
		}
		for _, part := range msg.ContentParts {
			if part.Type == "text" && part.AsTextContentPart().Text == want {
				return true
			}
		}
	}
	return false
}