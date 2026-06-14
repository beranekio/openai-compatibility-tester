package suites

import (
	"bytes"
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ImagesVariations verifies POST /v1/images/variations via client.Images.NewVariation.
type ImagesVariations struct{}

func (ImagesVariations) Name() string        { return "images_variations" }
func (ImagesVariations) Description() string { return "Image variations (POST /v1/images/variations)" }

func (ImagesVariations) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Images.NewVariation(ctx, openai.ImageNewVariationParams{
		Model:          openai.ImageModel(cfg.ImageModel),
		Image:          bytes.NewReader(smallPNGBytes()),
		N: openai.Int(1),
	})
	if err != nil {
		return fmt.Errorf("image variation request failed: %w", err)
	}
	return validateImagesResponse("images_variations", resp)
}