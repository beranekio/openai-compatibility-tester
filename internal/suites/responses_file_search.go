package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesFileSearch verifies hosted file search on POST /v1/responses.
type ResponsesFileSearch struct{}

func (ResponsesFileSearch) Name() string { return "responses_file_search" }
func (ResponsesFileSearch) Description() string {
	return "Responses API with file_search tool (POST /v1/responses)"
}

func (ResponsesFileSearch) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	store, err := createVectorStoreForSuite(ctx, client, "responses_file_search", "compatibility-test-responses-file-search")
	if err != nil {
		return err
	}
	var fileID string
	defer func() {
		cleanupVectorStoreArtifacts(client, store.ID, fileID)
	}()

	uploaded, err := uploadVectorStoreSourceFile(ctx, client, "responses_file_search")
	if err != nil {
		return err
	}
	fileID = uploaded.ID

	attached, err := client.VectorStores.Files.New(ctx, store.ID, openai.VectorStoreFileNewParams{
		FileID: uploaded.ID,
	})
	if err != nil {
		return fmt.Errorf("responses file search attach failed: %w", err)
	}
	if err := validateVectorStoreFileObject("responses_file_search", attached, store.ID); err != nil {
		return err
	}

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Search the files for the compatibility test phrase."),
		},
		Tools: []responses.ToolUnionParam{
			responses.ToolParamOfFileSearch([]string{store.ID}),
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses file search request failed: %w", err)
	}
	return validateHostedToolResponse("responses_file_search", resp, "file_search_call")
}
