package suites

import (
	"encoding/base64"
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
	if !strings.HasPrefix(mediaType, "audio/") &&
		mediaType != "application/octet-stream" &&
		mediaType != "application/zip" &&
		mediaType != "application/binary" {
		return fail(suite, fmt.Sprintf("Content-Type is %q, want audio/*, application/octet-stream, application/zip, or application/binary", mediaType))
	}
	return nil
}

func validateBase64Data(suite string, data string, minBytes int) error {
	if strings.TrimSpace(data) == "" {
		return fail(suite, "audio data is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fail(suite, fmt.Sprintf("audio data is not valid base64: %v", err))
	}
	if len(decoded) < minBytes {
		return fail(suite, fmt.Sprintf("audio data has %d bytes after decode, want at least %d", len(decoded), minBytes))
	}
	return nil
}

func validateWAVBytes(suite string, data []byte) error {
	if len(data) < 12 {
		return fail(suite, "audio data is too short to be WAV")
	}
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return fail(suite, "audio data is not a WAV file")
	}
	return nil
}

func validateBase64WAVData(suite string, data string, minBytes int) error {
	if strings.TrimSpace(data) == "" {
		return fail(suite, "audio data is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fail(suite, fmt.Sprintf("audio data is not valid base64: %v", err))
	}
	if len(decoded) < minBytes {
		return fail(suite, fmt.Sprintf("audio data has %d bytes after decode, want at least %d", len(decoded), minBytes))
	}
	return validateWAVBytes(suite, decoded)
}