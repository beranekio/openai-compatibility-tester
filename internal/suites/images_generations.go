package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
)

// ImagesGenerations verifies POST /v1/images/generations via client.Images.Generate.
type ImagesGenerations struct{}

func (ImagesGenerations) Name() string        { return "images_generations" }
func (ImagesGenerations) Description() string { return "Image generation (POST /v1/images/generations)" }

func (ImagesGenerations) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Images.Generate(ctx, openai.ImageGenerateParams{
		Model: openai.ImageModel(cfg.ImageModel),
		Prompt: "A simple red circle on a white background.",
		N:      openai.Int(1),
		ResponseFormat: openai.ImageGenerateParamsResponseFormatB64JSON,
	})
	if err != nil {
		return fmt.Errorf("image generation request failed: %w", err)
	}
	return validateImagesResponse("images_generations", resp)
}