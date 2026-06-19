package testutil

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

// ValidateBinaryHTTPResponse checks that resp contains a non-empty binary audio payload.
func ValidateBinaryHTTPResponse(resp *http.Response, minBytes int) error {
	if resp == nil {
		return fmt.Errorf("response is nil")
	}
	if resp.Body == nil {
		return fmt.Errorf("response body is nil")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if len(body) < minBytes {
		return fmt.Errorf("response body has %d bytes, want at least %d", len(body), minBytes)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		return fmt.Errorf("response missing Content-Type")
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("Content-Type %q is invalid: %w", contentType, err)
	}
	if !strings.HasPrefix(mediaType, "audio/") && mediaType != "application/octet-stream" {
		return fmt.Errorf("Content-Type is %q, want audio/* or application/octet-stream", mediaType)
	}
	return nil
}

// ValidateBase64Data checks that data is non-empty valid base64 with at least minBytes decoded.
func ValidateBase64Data(data string, minBytes int) error {
	if strings.TrimSpace(data) == "" {
		return fmt.Errorf("data is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("data is not valid base64: %w", err)
	}
	if len(decoded) < minBytes {
		return fmt.Errorf("data has %d bytes after decode, want at least %d", len(decoded), minBytes)
	}
	return nil
}

// ValidateWAVBytes checks that data looks like a minimal WAV file.
func ValidateWAVBytes(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("data is too short to be WAV")
	}
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return fmt.Errorf("data is not a WAV file")
	}
	return nil
}

// ValidateBase64WAVData checks that data is valid base64-encoded WAV with at least minBytes decoded.
func ValidateBase64WAVData(data string, minBytes int) error {
	if strings.TrimSpace(data) == "" {
		return fmt.Errorf("data is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("data is not valid base64: %w", err)
	}
	if len(decoded) < minBytes {
		return fmt.Errorf("data has %d bytes after decode, want at least %d", len(decoded), minBytes)
	}
	return ValidateWAVBytes(decoded)
}
