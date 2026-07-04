package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ImagesGenerationsStream verifies streaming POST /v1/images/generations via
// client.Images.GenerateStreaming.
type ImagesGenerationsStream struct{}

func (ImagesGenerationsStream) Name() string { return "images_generations_stream" }
func (ImagesGenerationsStream) Description() string {
	return "Streaming image generation (POST /v1/images/generations, stream=true)"
}

func (ImagesGenerationsStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Images.GenerateStreaming(ctx, openai.ImageGenerateParams{
		Model:         openai.ImageModel(cfg.ImageModel),
		Prompt:        "A simple red circle on a white background.",
		PartialImages: openai.Int(1),
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("image generation stream failed: %w", err)
	}
	if err := validateEventStreamContentType("images_generations_stream", httpResp); err != nil {
		return err
	}

	return consumeImageStream("images_generations_stream", stream, "image_generation", func(e openai.ImageGenStreamEventUnion) imageStreamEventInfo {
		return imageStreamEventInfo{
			eventType:              e.Type,
			b64JSON:                e.B64JSON,
			createdAtValid:         e.JSON.CreatedAt.Valid(),
			outputFormatValid:      e.JSON.OutputFormat.Valid(),
			sizeValid:              e.JSON.Size.Valid(),
			partialImageIndexValid: e.JSON.PartialImageIndex.Valid(),
		}
	})
}
