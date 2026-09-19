package suites

import (
	"context"
	"net/http"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesGetStream verifies GET /v1/responses/{id} with stream=true.
type ResponsesGetStream struct{}

func (ResponsesGetStream) Name() string { return "responses_get_stream" }
func (ResponsesGetStream) Description() string {
	return "Responses API retrieve stream (POST /v1/responses with store, then GET /v1/responses/{id} stream=true)"
}

func (ResponsesGetStream) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	created, err := createStoredResponse(ctx, client, cfg)
	if err != nil {
		return err
	}
	defer deleteStoredResponseBestEffort(client, created.ID)

	var httpResp *http.Response
	stream := client.Responses.GetStreaming(ctx, created.ID, responses.ResponseGetParams{}, option.WithResponseInto(&httpResp))
	return consumeResponsesTextStream("responses_get_stream", stream, httpResp)
}
