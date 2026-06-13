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
		Store: openai.Bool(false),
	})
	defer stream.Close()

	var hasOutput bool
	var completed bool
	var terminalFailure bool
	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "response.output_text.delta", "response.refusal.delta":
			if event.Delta != "" {
				hasOutput = true
			}
		case "response.completed":
			completed = true
		case "response.failed", "response.incomplete", "error":
			terminalFailure = true
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("responses stream failed: %w", err)
	}
	if terminalFailure {
		return fail("responses_stream", "stream ended with a failure event")
	}
	if !completed {
		return fail("responses_stream", "stream missing response.completed event")
	}
	if !hasOutput {
		return fail("responses_stream", "stream produced no output text or refusal")
	}
	return nil
}