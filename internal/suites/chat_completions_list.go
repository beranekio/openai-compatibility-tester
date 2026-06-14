package suites

import (
	"context"
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

	item, err := findStoredChatCompletionInList(ctx, client, "chat_completions_list", created.ID)
	if err != nil {
		return err
	}
	if err := validateChatCompletionEnvelope("chat_completions_list", item); err != nil {
		return err
	}
	if len(item.Choices) == 0 {
		return fail("chat_completions_list", "listed completion missing choices")
	}
	return validateChatCompletionChoice("chat_completions_list", item.Choices[0])
}