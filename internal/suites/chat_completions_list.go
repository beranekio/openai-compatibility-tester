package suites

import (
	"context"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsList verifies GET /v1/chat/completions lists a stored chat completion.
type ChatCompletionsList struct{}

func (ChatCompletionsList) Name() string { return "chat_completions_list" }
func (ChatCompletionsList) Description() string {
	return "Chat completion list (POST /v1/chat/completions with store, then GET /v1/chat/completions)"
}

func (ChatCompletionsList) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredChatCompletion(ctx, client, cfg)
	if err != nil {
		return err
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = client.Chat.Completions.Delete(cleanupCtx, created.ID)
	}()

	page, err := client.Chat.Completions.List(ctx, openai.ChatCompletionListParams{})
	if err != nil {
		return fmt.Errorf("chat completion list failed: %w", err)
	}
	if page == nil {
		return fail("chat_completions_list", "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail("chat_completions_list", "list missing has_more")
	}
	for _, item := range page.Data {
		if item.ID == created.ID {
			if err := validateChatCompletionEnvelope("chat_completions_list", &item); err != nil {
				return err
			}
			return nil
		}
	}
	return fail("chat_completions_list", fmt.Sprintf("list missing stored completion id %q", created.ID))
}