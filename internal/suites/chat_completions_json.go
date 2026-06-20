package suites

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// ChatCompletionsJSON verifies POST /v1/chat/completions with response_format json_schema.
type ChatCompletionsJSON struct{}

func (ChatCompletionsJSON) Name() string { return "chat_completions_json" }
func (ChatCompletionsJSON) Description() string {
	return "Chat completion structured JSON (POST /v1/chat/completions, response_format json_schema)"
}

func (ChatCompletionsJSON) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Reply with JSON containing an answer field"),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "answer",
					Strict: openai.Bool(true),
					Schema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"answer": map[string]any{
								"type": "string",
							},
						},
						"required":             []string{"answer"},
						"additionalProperties": false,
					},
				},
			},
		},
		Store: openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("chat completion json request failed: %w", err)
	}
	if resp == nil {
		return fail("chat_completions_json", "response is nil")
	}
	if resp.ID == "" {
		return fail("chat_completions_json", "response missing id")
	}
	if len(resp.Choices) == 0 {
		return fail("chat_completions_json", "response missing choices")
	}

	choice := resp.Choices[0]
	if choice.FinishReason == "" {
		return fail("chat_completions_json", "choice missing finish_reason")
	}
	if string(choice.Message.Role) != "assistant" {
		return fail("chat_completions_json", fmt.Sprintf("choice message role is %q, want assistant", choice.Message.Role))
	}
	if !hasChatMessageOutput(choice.Message) && !isContentFilterFinishReason(choice.FinishReason) {
		return fail("chat_completions_json", "choice message has no content or refusal")
	}
	if isContentFilterFinishReason(choice.FinishReason) {
		return nil
	}
	if choice.Message.Refusal != "" {
		return nil
	}

	if err := validateStructuredAnswerJSON("chat_completions_json", choice.Message.Content); err != nil {
		return err
	}
	return nil
}

func validateStructuredAnswerJSON(suite string, content string) error {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return fail(suite, fmt.Sprintf("message content is not valid JSON: %v", err))
	}
	if len(parsed) != 1 {
		return fail(suite, fmt.Sprintf("parsed JSON has %d top-level fields, want 1", len(parsed)))
	}
	answer, ok := parsed["answer"]
	if !ok {
		return fail(suite, `parsed JSON missing "answer" field`)
	}
	if _, ok := answer.(string); !ok {
		return fail(suite, `"answer" field is not a string`)
	}
	return nil
}
