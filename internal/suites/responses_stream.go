package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesStream verifies streaming responses API.
type ResponsesStream struct{}

func (ResponsesStream) Name() string { return "responses_stream" }
func (ResponsesStream) Description() string {
	return "Streaming responses API (POST /v1/responses, stream=true)"
}

func (ResponsesStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	stream := client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Model: cfg.Model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Count from one to three."),
		},
		MaxOutputTokens: openai.Int(32),
	})

	events := 0
	var text string
	for stream.Next() {
		event := stream.Current()
		events++
		text += event.Delta
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("responses stream failed: %w", err)
	}
	if events == 0 {
		return fail("responses_stream", "stream returned no events")
	}
	if text == "" {
		return fail("responses_stream", "stream produced no text content")
	}
	return nil
}