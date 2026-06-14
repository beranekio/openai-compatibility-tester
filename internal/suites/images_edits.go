package suites

import (
	"bytes"
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ImagesEdits verifies POST /v1/images/edits via client.Images.Edit.
type ImagesEdits struct{}

func (ImagesEdits) Name() string        { return "images_edits" }
func (ImagesEdits) Description() string { return "Image edits (POST /v1/images/edits)" }

func (ImagesEdits) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Images.Edit(ctx, openai.ImageEditParams{
		Model:  openai.ImageModel(cfg.ImageModel),
		Prompt: "Add a blue border.",
		Image: openai.ImageEditParamsImageUnion{
			OfFile: bytes.NewReader(smallPNGBytes()),
		},
		N: openai.Int(1),
	})
	if err != nil {
		return fmt.Errorf("image edit request failed: %w", err)
	}
	return validateImagesResponse("images_edits", resp)
}