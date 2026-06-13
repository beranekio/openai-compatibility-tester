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
	if string(resp.Object) != "list" {
		return fail("embeddings", fmt.Sprintf("response object is %q, want list", resp.Object))
	}
	if resp.Model == "" {
		return fail("embeddings", "response missing model")
	}
	if !resp.JSON.Usage.Valid() {
		return fail("embeddings", "response missing usage")
	}
	if len(resp.Data) != 1 {
		return fail("embeddings", fmt.Sprintf("response has %d embeddings, want 1", len(resp.Data)))
	}
	for i, item := range resp.Data {
		if !item.JSON.Index.Valid() {
			return fail("embeddings", fmt.Sprintf("embedding %d missing index", i))
		}
		if item.Index != 0 {
			return fail("embeddings", fmt.Sprintf("embedding %d index is %d, want 0", i, item.Index))
		}
		if string(item.Object) != "embedding" {
			return fail("embeddings", fmt.Sprintf("embedding %d object is %q, want embedding", i, item.Object))
		}
		if len(item.Embedding) == 0 {
			return fail("embeddings", fmt.Sprintf("embedding %d vector is empty", i))
		}
	}
	return nil
}