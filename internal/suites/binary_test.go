package suites

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func testHTTPResponse(contentType, body string) *http.Response {
	return &http.Response{
		Header: http.Header{"Content-Type": {contentType}},
		Body:   io.NopCloser(strings.NewReader(body)),
	}
}

func TestValidateBinaryHTTPResponse(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		err := validateBinaryHTTPResponse("audio_speech", nil, 1)
		if err == nil {
			t.Fatal("expected error for nil response")
		}
	})

	t.Run("nil body", func(t *testing.T) {
		err := validateBinaryHTTPResponse("audio_speech", &http.Response{}, 1)
		if err == nil {
			t.Fatal("expected error for nil body")
		}
	})

	tests := []struct {
		name        string
		contentType string
		body        string
		wantErr     bool
	}{
		{name: "audio mpeg", contentType: "audio/mpeg", body: "x", wantErr: false},
		{name: "octet stream", contentType: "application/octet-stream", body: "x", wantErr: false},
		{name: "json error", contentType: "application/json", body: `{"error":"oops"}`, wantErr: true},
		{name: "html error page", contentType: "text/html", body: "<html>error</html>", wantErr: true},
		{name: "plain text", contentType: "text/plain", body: "not audio", wantErr: true},
		{name: "missing content type", contentType: "", body: "x", wantErr: true},
		{name: "empty body", contentType: "audio/mpeg", body: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBinaryHTTPResponse("audio_speech", testHTTPResponse(tt.contentType, tt.body), 1)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateBinaryHTTPResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}