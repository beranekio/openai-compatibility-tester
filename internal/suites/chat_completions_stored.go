package suites

import (
	"context"
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
	return validateCursorListPage(suite, page, func(c *openai.ChatCompletion) string { return c.ID })
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