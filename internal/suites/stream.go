package suites

import (
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/openai/openai-go/v3"
)

func validateEventStreamContentType(suite string, resp *http.Response) error {
	if resp == nil {
		return fail(suite, "stream response is nil")
	}
	contentType := resp.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fail(suite, fmt.Sprintf("Content-Type %q is invalid: %v", contentType, err))
	}
	if mediaType != "text/event-stream" {
		return fail(suite, fmt.Sprintf("Content-Type is %q, want text/event-stream", strings.TrimSpace(contentType)))
	}
	return nil
}

func validateCompletionChunk(suite string, chunk openai.Completion) error {
	if !chunk.JSON.Created.Valid() {
		return fail(suite, "stream chunk missing created")
	}
	if chunk.Model == "" {
		return fail(suite, "stream chunk missing model")
	}
	if string(chunk.Object) != "text_completion" {
		return fail(suite, fmt.Sprintf("stream chunk object is %q, want text_completion", chunk.Object))
	}
	if len(chunk.Choices) == 0 {
		return fail(suite, "stream chunk missing choices")
	}
	choice := chunk.Choices[0]
	if !choice.JSON.Index.Valid() {
		return fail(suite, "stream chunk choice missing index")
	}
	return nil
}

func validateChatCompletionChunkEnvelope(suite string, chunk openai.ChatCompletionChunk) error {
	if !chunk.JSON.Created.Valid() {
		return fail(suite, "stream chunk missing created")
	}
	if chunk.Model == "" {
		return fail(suite, "stream chunk missing model")
	}
	if string(chunk.Object) != "chat.completion.chunk" {
		return fail(suite, fmt.Sprintf("stream chunk object is %q, want chat.completion.chunk", chunk.Object))
	}
	return nil
}

func validateChatCompletionChunkChoice(suite string, chunk openai.ChatCompletionChunk) error {
	if len(chunk.Choices) == 0 {
		return fail(suite, "stream chunk missing choices")
	}
	choice := chunk.Choices[0]
	if !choice.JSON.Index.Valid() {
		return fail(suite, "stream chunk choice missing index")
	}
	if !choice.JSON.Delta.Valid() {
		return fail(suite, "stream chunk choice missing delta")
	}
	return nil
}

func validateChatCompletionChunk(suite string, chunk openai.ChatCompletionChunk) error {
	if err := validateChatCompletionChunkEnvelope(suite, chunk); err != nil {
		return err
	}
	return validateChatCompletionChunkChoice(suite, chunk)
}

func validateChatCompletionStreamUsage(suite string, chunk openai.ChatCompletionChunk) error {
	if !chunk.JSON.Usage.Valid() {
		return fail(suite, "stream chunk missing usage")
	}
	usage := chunk.Usage
	if !usage.JSON.PromptTokens.Valid() {
		return fail(suite, "stream usage missing prompt_tokens")
	}
	if !usage.JSON.CompletionTokens.Valid() {
		return fail(suite, "stream usage missing completion_tokens")
	}
	if !usage.JSON.TotalTokens.Valid() {
		return fail(suite, "stream usage missing total_tokens")
	}
	if usage.TotalTokens <= 0 {
		return fail(suite, "stream usage total_tokens must be greater than zero")
	}
	if usage.TotalTokens != usage.PromptTokens+usage.CompletionTokens {
		return fail(suite, "stream usage total_tokens does not equal prompt_tokens + completion_tokens")
	}
	return nil
}