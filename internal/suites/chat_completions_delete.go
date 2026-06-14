package suites

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsDelete verifies DELETE /v1/chat/completions/{id}.
type ChatCompletionsDelete struct{}

func (ChatCompletionsDelete) Name() string { return "chat_completions_delete" }
func (ChatCompletionsDelete) Description() string {
	return "Chat completion delete (POST /v1/chat/completions with store, then DELETE /v1/chat/completions/{id})"
}

func (ChatCompletionsDelete) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredChatCompletion(ctx, client, cfg)
	if err != nil {
		return err
	}
	deleted := false
	defer func() {
		if !deleted {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = client.Chat.Completions.Delete(cleanupCtx, created.ID)
		}
	}()

	got, err := client.Chat.Completions.Get(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("get before delete failed: %w", err)
	}
	if err := validateChatCompletionEnvelope("chat_completions_delete", got); err != nil {
		return err
	}

	resp, err := client.Chat.Completions.Delete(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("chat completion delete failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_delete", "delete response is nil")
	}
	if resp.ID != created.ID {
		return fail("chat_completions_delete", fmt.Sprintf("delete id is %q, want %q", resp.ID, created.ID))
	}
	if !resp.Deleted {
		return fail("chat_completions_delete", "delete response deleted is false")
	}
	if string(resp.Object) != "chat.completion.deleted" {
		return fail("chat_completions_delete", fmt.Sprintf("delete object is %q, want chat.completion.deleted", resp.Object))
	}
	deleted = true

	_, getErr := client.Chat.Completions.Get(ctx, created.ID)
	if getErr == nil {
		return fail("chat_completions_delete", "get after delete succeeded; completion still exists")
	}
	var apiErr *openai.Error
	if !errors.As(getErr, &apiErr) {
		return fmt.Errorf("get after delete failed: %w", getErr)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		return fail("chat_completions_delete", fmt.Sprintf("get after delete returned status %d, want 404", apiErr.StatusCode))
	}
	return nil
}