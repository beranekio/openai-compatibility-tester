package suites

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

const storedChatCompletionInput = "Reply with exactly the word: pong"

func createStoredChatCompletion(ctx context.Context, client openai.Client, cfg *config.Config) (*openai.ChatCompletion, error) {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(storedChatCompletionInput),
		},
		Store: openai.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("create stored chat completion failed: %w", err)
	}
	if resp == nil {
		return nil, fail("chat_completions", "create stored chat completion returned nil")
	}
	if resp.ID == "" {
		return nil, fail("chat_completions", "create stored chat completion missing id")
	}
	return resp, nil
}

func validateChatCompletionListPage(suite string, page *pagination.CursorPage[openai.ChatCompletion]) error {
	if page == nil {
		return fail(suite, "list page is nil")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "list missing has_more")
	}
	var envelope struct {
		Object  string `json:"object"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
	}
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	if len(page.Data) == 0 {
		return nil
	}
	if envelope.FirstID == "" {
		return fail(suite, "list missing first_id")
	}
	if envelope.LastID == "" {
		return fail(suite, "list missing last_id")
	}
	if envelope.FirstID != page.Data[0].ID {
		return fail(suite, fmt.Sprintf("list first_id is %q, want %q", envelope.FirstID, page.Data[0].ID))
	}
	if envelope.LastID != page.Data[len(page.Data)-1].ID {
		return fail(suite, fmt.Sprintf("list last_id is %q, want %q", envelope.LastID, page.Data[len(page.Data)-1].ID))
	}
	return nil
}

func findStoredChatCompletionInList(ctx context.Context, client openai.Client, suite, wantID string) (*openai.ChatCompletion, error) {
	page, err := client.Chat.Completions.List(ctx, openai.ChatCompletionListParams{
		Order: openai.ChatCompletionListParamsOrderDesc,
	})
	if err != nil {
		return nil, fmt.Errorf("chat completion list failed: %w", err)
	}
	for {
		if err := validateChatCompletionListPage(suite, page); err != nil {
			return nil, err
		}
		for i := range page.Data {
			if page.Data[i].ID == wantID {
				item := page.Data[i]
				return &item, nil
			}
		}
		if !page.HasMore {
			break
		}
		page, err = page.GetNextPage()
		if err != nil {
			return nil, fmt.Errorf("chat completion list next page failed: %w", err)
		}
		if page == nil {
			break
		}
	}
	return nil, fail(suite, fmt.Sprintf("list missing stored completion id %q", wantID))
}
