package suites

import (
	"context"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ChatCompletionsGet verifies GET /v1/chat/completions/{id} after creating a stored chat completion.
type ChatCompletionsGet struct{}

func (ChatCompletionsGet) Name() string { return "chat_completions_get" }
func (ChatCompletionsGet) Description() string {
	return "Chat completion get (POST /v1/chat/completions with store, then GET /v1/chat/completions/{id})"
}

func (ChatCompletionsGet) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredChatCompletion(ctx, client, cfg)
	if err != nil {
		return err
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = client.Chat.Completions.Delete(cleanupCtx, created.ID)
	}()

	got, err := client.Chat.Completions.Get(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("chat completion get failed: %w", err)
	}
	if err := validateChatCompletionEnvelope("chat_completions_get", got); err != nil {
		return err
	}
	if got.ID != created.ID {
		return fail("chat_completions_get", fmt.Sprintf("get id is %q, want %q", got.ID, created.ID))
	}
	if len(got.Choices) == 0 {
		return fail("chat_completions_get", "get response missing choices")
	}
	return validateChatCompletionChoice("chat_completions_get", got.Choices[0])
}
