package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesWebSearch verifies hosted web search on POST /v1/responses.
type ResponsesWebSearch struct{}

func (ResponsesWebSearch) Name() string { return "responses_web_search" }
func (ResponsesWebSearch) Description() string {
	return "Responses API with web_search tool (POST /v1/responses)"
}

func (ResponsesWebSearch) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("What is a positive news story from today?"),
		},
		Tools: []responses.ToolUnionParam{
			responses.ToolParamOfWebSearch(responses.WebSearchToolTypeWebSearch),
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses web search request failed: %w", err)
	}
	return validateHostedToolResponse("responses_web_search", resp, "web_search_call")
}
