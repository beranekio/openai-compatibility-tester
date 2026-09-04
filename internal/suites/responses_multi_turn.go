package suites

import (
	"context"
	"fmt"

	"github.com/beranekio/openai-compatibility-tester/internal/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// ResponsesMultiTurn verifies a follow-up POST /v1/responses using previous_response_id.
type ResponsesMultiTurn struct{}

func (ResponsesMultiTurn) Name() string { return "responses_multi_turn" }
func (ResponsesMultiTurn) Description() string {
	return "Responses API multi-turn (POST /v1/responses with previous_response_id)"
}

func (ResponsesMultiTurn) Run(ctx context.Context, client openai.Client, cfg *config.Config) error {
	first, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Remember the word: pong"),
		},
		Store: openai.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("responses multi-turn first request failed: %w", err)
	}
	if err := validateResponseEnvelope("responses_multi_turn", first); err != nil {
		return err
	}
	defer deleteStoredResponseBestEffort(client, first.ID)

	if string(first.Status) == "completed" {
		if err := validateCompletedResponseHasOutput("responses_multi_turn", first); err != nil {
			return err
		}
	} else if !isContentFilterIncompleteResponse(first) {
		return fail("responses_multi_turn", fmt.Sprintf("first response status is %q, want completed", first.Status))
	}

	second, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: cfg.ResponsesModel,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Reply with the word I asked you to remember."),
		},
		PreviousResponseID: openai.String(first.ID),
		Store:              openai.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("responses multi-turn follow-up failed: %w", err)
	}
	if err := validateResponseEnvelope("responses_multi_turn", second); err != nil {
		return err
	}
	if second.ID == first.ID {
		return fail("responses_multi_turn", "follow-up response id matches the first response")
	}
	if string(second.Status) == "completed" {
		return validateCompletedResponseHasOutput("responses_multi_turn", second)
	}
	if isContentFilterIncompleteResponse(second) {
		return nil
	}
	return fail("responses_multi_turn", fmt.Sprintf("follow-up status is %q, want completed", second.Status))
}
