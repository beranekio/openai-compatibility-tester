package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// Embeddings verifies POST /v1/embeddings via client.Embeddings.New.
type Embeddings struct{}

func (Embeddings) Name() string        { return "embeddings" }
func (Embeddings) Description() string { return "Embeddings (POST /v1/embeddings)" }

func (Embeddings) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(cfg.EmbeddingModel),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String("compatibility test"),
		},
	})
	if err != nil {
		return fmt.Errorf("embedding request failed: %w", err)
	}
	if resp == nil {
		return fail("embeddings", "response is nil")
	}
	if len(resp.Data) == 0 {
		return fail("embeddings", "response missing data")
	}
	if len(resp.Data[0].Embedding) == 0 {
		return fail("embeddings", "embedding vector is empty")
	}
	return nil
}