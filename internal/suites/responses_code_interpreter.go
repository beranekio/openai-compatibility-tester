package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesCodeInterpreter verifies hosted code interpreter on POST /v1/responses.
type ResponsesCodeInterpreter struct{}

func (ResponsesCodeInterpreter) Name() string { return "responses_code_interpreter" }
func (ResponsesCodeInterpreter) Description() string {
	return "Responses API with code_interpreter tool (POST /v1/responses)"
}

func (ResponsesCodeInterpreter) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("What is 1+1?"),
		},
		Tools: []responses.ToolUnionParam{
			responses.ToolParamOfCodeInterpreter(responses.ToolCodeInterpreterContainerCodeInterpreterContainerAutoParam{}),
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses code interpreter request failed: %w", err)
	}
	return validateHostedToolResponse("responses_code_interpreter", resp, "code_interpreter_call")
}
