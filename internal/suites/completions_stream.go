package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// CompletionsStream verifies streaming legacy completions.
type CompletionsStream struct{}

func (CompletionsStream) Name() string { return "completions_stream" }
func (CompletionsStream) Description() string {
	return "Streaming text completion (POST /v1/completions, stream=true)"
}

func (CompletionsStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Completions.NewStreaming(ctx, openai.CompletionNewParams{
		Model: openai.CompletionNewParamsModel(cfg.CompletionModel),
		Prompt: openai.CompletionNewParamsPromptUnion{
			OfString: openai.String("Say hello"),
		},
		MaxTokens: openai.Int(16),
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("completion stream failed: %w", err)
	}
	if err := validateEventStreamContentType("completions_stream", httpResp); err != nil {
		return err
	}

	chunks := 0
	var hasOutput bool
	var finished bool
	var finishReason string
	for stream.Next() {
		chunk := stream.Current()
		chunks++
		if chunk.ID == "" {
			return fail("completions_stream", "stream chunk missing id")
		}
		if err := validateCompletionChunk("completions_stream", chunk); err != nil {
			return err
		}
		choice := chunk.Choices[0]
		if string(choice.FinishReason) != "" {
			finished = true
			finishReason = string(choice.FinishReason)
		}
		if choice.Text != "" {
			hasOutput = true
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("completion stream failed: %w", err)
	}
	if chunks == 0 {
		return fail("completions_stream", "stream returned no chunks")
	}
	if !finished {
		return fail("completions_stream", "stream missing terminal finish_reason")
	}
	if !hasOutput && !isContentFilterFinishReason(finishReason) {
		return fail("completions_stream", "stream produced no text")
	}
	return nil
}
