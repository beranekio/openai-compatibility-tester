package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// EmbeddingsBatch verifies POST /v1/embeddings with array input via client.Embeddings.New.
type EmbeddingsBatch struct{}

func (EmbeddingsBatch) Name() string        { return "embeddings_batch" }
func (EmbeddingsBatch) Description() string { return "Embeddings batch input (POST /v1/embeddings)" }

func (EmbeddingsBatch) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(cfg.EmbeddingModel),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{"compatibility test", "batch input"},
		},
	})
	if err != nil {
		return fmt.Errorf("embedding batch request failed: %w", err)
	}
	if resp == nil {
		return fail("embeddings_batch", "response is nil")
	}
	if string(resp.Object) != "list" {
		return fail("embeddings_batch", fmt.Sprintf("response object is %q, want list", resp.Object))
	}
	if resp.Model == "" {
		return fail("embeddings_batch", "response missing model")
	}
	if !resp.JSON.Usage.Valid() {
		return fail("embeddings_batch", "response missing usage")
	}
	if len(resp.Data) != 2 {
		return fail("embeddings_batch", fmt.Sprintf("response has %d embeddings, want 2", len(resp.Data)))
	}
	for i, item := range resp.Data {
		if !item.JSON.Index.Valid() {
			return fail("embeddings_batch", fmt.Sprintf("embedding %d missing index", i))
		}
		if item.Index != int64(i) {
			return fail("embeddings_batch", fmt.Sprintf("embedding %d index is %d, want %d", i, item.Index, i))
		}
		if string(item.Object) != "embedding" {
			return fail("embeddings_batch", fmt.Sprintf("embedding %d object is %q, want embedding", i, item.Object))
		}
		if len(item.Embedding) == 0 {
			return fail("embeddings_batch", fmt.Sprintf("embedding %d vector is empty", i))
		}
	}
	return nil
}