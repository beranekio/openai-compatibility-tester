package suites

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"
	"github.com/beranekio/openai-compatibility-tester/internal/testutil"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ImagesEditsStream verifies streaming POST /v1/images/edits via
// client.Images.EditStreaming.
type ImagesEditsStream struct{}

func (ImagesEditsStream) Name() string { return "images_edits_stream" }
func (ImagesEditsStream) Description() string {
	return "Streaming image edit (POST /v1/images/edits, stream=true)"
}

func (ImagesEditsStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	var httpResp *http.Response
	stream := client.Images.EditStreaming(ctx, openai.ImageEditParams{
		Model:         openai.ImageModel(cfg.ImageModel),
		Prompt:        "Add a blue border.",
		Image:         openai.ImageEditParamsImageUnion{OfFile: testutil.SmallPNGReader()},
		PartialImages: openai.Int(1),
	}, option.WithResponseInto(&httpResp))
	defer stream.Close()

	if err := stream.Err(); err != nil {
		return fmt.Errorf("image edit stream failed: %w", err)
	}
	if err := validateEventStreamContentType("images_edits_stream", httpResp); err != nil {
		return err
	}

	return consumeImageStream("images_edits_stream", stream, "image_edit", func(e openai.ImageEditStreamEventUnion) (string, string) {
		return e.Type, e.B64JSON
	})
}
