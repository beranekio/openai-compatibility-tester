package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesVision verifies multimodal POST /v1/responses with image input.
type ResponsesVision struct{}

func (ResponsesVision) Name() string { return "responses_vision" }
func (ResponsesVision) Description() string {
	return "Responses API with vision input (POST /v1/responses)"
}

func (ResponsesVision) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.VisionModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfMessage(
					responses.ResponseInputMessageContentListParam{
						responses.ResponseInputContentParamOfInputText("Describe this image in one short sentence."),
						responsesVisionImageContent(),
					},
					responses.EasyInputMessageRoleUser,
				),
			},
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses vision request failed: %w", err)
	}
	if err := validateResponseEnvelope("responses_vision", resp); err != nil {
		return err
	}
	if string(resp.Status) == "completed" {
		return validateCompletedResponseHasOutput("responses_vision", resp)
	}
	if isContentFilterIncompleteResponse(resp) {
		return nil
	}
	return fail("responses_vision", fmt.Sprintf("response status is %q, want completed", resp.Status))
}

func responsesVisionImageContent() responses.ResponseInputContentUnionParam {
	image := responses.ResponseInputContentParamOfInputImage(responses.ResponseInputImageDetailLow)
	image.OfInputImage.ImageURL = openai.String(smallPNGDataURL)
	return image
}
