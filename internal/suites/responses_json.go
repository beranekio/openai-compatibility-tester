package suites

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesJSON verifies structured JSON output on POST /v1/responses.
type ResponsesJSON struct{}

func (ResponsesJSON) Name() string { return "responses_json" }
func (ResponsesJSON) Description() string {
	return "Responses API structured JSON (POST /v1/responses, text.format json_schema)"
}

func (ResponsesJSON) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Reply with JSON containing an answer field"),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "answer",
					Strict: openai.Bool(true),
					Schema: answerJSONSchema(),
				},
			},
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses json request failed: %w", err)
	}
	if resp == nil {
		return fail("responses_json", "response is nil")
	}
	if err := validateResponseEnvelope("responses_json", resp); err != nil {
		return err
	}
	if resp.Status == responses.ResponseStatusCompleted {
		if isContentFilterIncompleteResponse(resp) {
			return nil
		}
		if responseOutputRefusal(resp) != "" {
			return nil
		}
		text := resp.OutputText()
		if text == "" {
			return fail("responses_json", "response produced no output text")
		}
		var parsed map[string]any
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			return fail("responses_json", fmt.Sprintf("output text is not valid JSON: %v", err))
		}
		if len(parsed) != 1 {
			return fail("responses_json", fmt.Sprintf("parsed JSON has %d top-level fields, want 1", len(parsed)))
		}
		answer, ok := parsed["answer"]
		if !ok {
			return fail("responses_json", `parsed JSON missing "answer" field`)
		}
		if _, ok := answer.(string); !ok {
			return fail("responses_json", `"answer" field is not a string`)
		}
		return nil
	}
	if isContentFilterIncompleteResponse(resp) {
		return nil
	}
	return fail("responses_json", fmt.Sprintf("response status is %q, want completed", resp.Status))
}

func answerJSONSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"answer": map[string]any{
				"type": "string",
			},
		},
		"required":             []string{"answer"},
		"additionalProperties": false,
	}
}
