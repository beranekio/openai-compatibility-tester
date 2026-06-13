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

func validateChatCompletionChunk(suite string, chunk openai.ChatCompletionChunk) error {
	if !chunk.JSON.Created.Valid() {
		return fail(suite, "stream chunk missing created")
	}
	if chunk.Model == "" {
		return fail(suite, "stream chunk missing model")
	}
	if string(chunk.Object) != "chat.completion.chunk" {
		return fail(suite, fmt.Sprintf("stream chunk object is %q, want chat.completion.chunk", chunk.Object))
	}
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