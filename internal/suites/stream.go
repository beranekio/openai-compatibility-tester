package suites

import (
	"fmt"
	"mime"
	"net/http"
	"strings"
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