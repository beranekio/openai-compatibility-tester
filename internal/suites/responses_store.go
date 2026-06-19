package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

const storedResponseInput = "Reply with exactly the word: pong"

func createStoredResponse(ctx context.Context, client openai.Client, cfg *config.Config) (*responses.Response, error) {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(storedResponseInput),
		},
		Store: openai.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("create stored response failed: %w", err)
	}
	if resp == nil {
		return nil, fail("responses", "create stored response returned nil")
	}
	if resp.ID == "" {
		return nil, fail("responses", "create stored response missing id")
	}
	return resp, nil
}
