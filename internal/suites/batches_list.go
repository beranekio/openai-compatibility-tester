package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// BatchesList verifies GET /v1/batches via client.Batches.List.
type BatchesList struct{}

func (BatchesList) Name() string { return "batches_list" }
func (BatchesList) Description() string {
	return "Batches API list (POST /v1/batches, then GET /v1/batches)"
}

func (BatchesList) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var batchID string
	var fileID string
	defer func() {
		cleanupBatchArtifacts(client, batchID, fileID)
	}()

	uploaded, err := uploadBatchInputFile(ctx, client, cfg)
	if err != nil {
		return err
	}
	fileID = uploaded.ID

	created, err := client.Batches.New(ctx, openai.BatchNewParams{
		CompletionWindow: openai.BatchNewParamsCompletionWindow24h,
		Endpoint:         openai.BatchNewParamsEndpointV1ChatCompletions,
		InputFileID:      uploaded.ID,
	})
	if err != nil {
		return fmt.Errorf("batch create failed: %w", err)
	}
	if created == nil || created.ID == "" {
		return fail("batches_list", "batch create missing id")
	}
	batchID = created.ID

	found, err := findBatchInList(ctx, client, "batches_list", created.ID)
	if err != nil {
		return err
	}
	if err := validateBatchEnvelope("batches_list", found); err != nil {
		return err
	}
	if found.ID != created.ID {
		return fail("batches_list", fmt.Sprintf("listed batch id is %q, want %q", found.ID, created.ID))
	}
	return nil
}

func findBatchInList(ctx context.Context, client openai.Client, suite, wantID string) (*openai.Batch, error) {
	page, err := client.Batches.List(ctx, openai.BatchListParams{
		Limit: openai.Int(100),
	})
	if err != nil {
		return nil, fmt.Errorf("batch list failed: %w", err)
	}
	for {
		if err := validateCursorListPage(suite, page, func(b *openai.Batch) string { return b.ID }); err != nil {
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
			return nil, fmt.Errorf("batch list next page failed: %w", err)
		}
		if page == nil {
			break
		}
	}
	return nil, fail(suite, fmt.Sprintf("list missing created batch id %q", wantID))
}
