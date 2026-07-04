package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// ChatCompletionsUpdate verifies POST /v1/chat/completions/{id} (update) via
// client.Chat.Completions.Update. It creates a stored completion, patches its
// metadata, and validates the patched field is reflected on the returned object.
// The SDK's ChatCompletion response struct does not expose metadata, so the
// patched value is read from the response's raw JSON.
type ChatCompletionsUpdate struct{}

func (ChatCompletionsUpdate) Name() string { return "chat_completions_update" }
func (ChatCompletionsUpdate) Description() string {
	return "Chat completion update (POST /v1/chat/completions with store, then POST /v1/chat/completions/{id})"
}

func (ChatCompletionsUpdate) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredChatCompletion(ctx, client, cfg)
	if err != nil {
		return err
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = client.Chat.Completions.Delete(cleanupCtx, created.ID)
	}()

	const metadataKey = "compat_test"
	const metadataValue = "updated"
	updated, err := client.Chat.Completions.Update(ctx, created.ID, openai.ChatCompletionUpdateParams{
		Metadata: shared.Metadata{
			metadataKey: metadataValue,
		},
	})
	if err != nil {
		return fmt.Errorf("chat completion update failed: %w", err)
	}
	if err := validateChatCompletionEnvelope("chat_completions_update", updated); err != nil {
		return err
	}
	if updated.ID != created.ID {
		return fail("chat_completions_update", fmt.Sprintf("update id is %q, want %q", updated.ID, created.ID))
	}
	if err := validateChatCompletionMetadata("chat_completions_update", updated.RawJSON(), metadataKey, metadataValue); err != nil {
		return err
	}

	// Re-fetch to confirm the patch persisted.
	refetched, err := client.Chat.Completions.Get(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("chat completion get after update failed: %w", err)
	}
	if err := validateChatCompletionEnvelope("chat_completions_update", refetched); err != nil {
		return err
	}
	if err := validateChatCompletionMetadata("chat_completions_update", refetched.RawJSON(), metadataKey, metadataValue); err != nil {
		return err
	}
	return nil
}

func validateChatCompletionMetadata(suite, rawJSON, key, want string) error {
	var envelope struct {
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &envelope); err != nil {
		return fail(suite, "update response is not valid JSON")
	}
	if envelope.Metadata == nil {
		return fail(suite, "update response missing metadata")
	}
	if got, ok := envelope.Metadata[key]; !ok {
		return fail(suite, fmt.Sprintf("update response metadata missing key %q", key))
	} else if got != want {
		return fail(suite, fmt.Sprintf("update metadata[%q] is %q, want %q", key, got, want))
	}
	return nil
}

