package suites

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

func validateBinaryHTTPResponse(suite string, resp *http.Response, minBytes int) error {
	if resp == nil {
		return fail(suite, "response is nil")
	}
	if resp.Body == nil {
		return fail(suite, "response body is nil")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: read response body: %w", suite, err)
	}
	if len(body) < minBytes {
		return fail(suite, fmt.Sprintf("response body has %d bytes, want at least %d", len(body), minBytes))
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		return fail(suite, "response missing Content-Type")
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fail(suite, fmt.Sprintf("Content-Type %q is invalid: %v", contentType, err))
	}
	if !strings.HasPrefix(mediaType, "audio/") && mediaType != "application/octet-stream" {
		return fail(suite, fmt.Sprintf("Content-Type is %q, want audio/* or application/octet-stream", mediaType))
	}
	return nil
}