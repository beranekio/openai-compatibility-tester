package testutil

import (
	"encoding/base64"
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
		err := ValidateBinaryHTTPResponse(nil, 1)
		if err == nil {
			t.Fatal("expected error for nil response")
		}
	})

	t.Run("nil body", func(t *testing.T) {
		err := ValidateBinaryHTTPResponse(&http.Response{}, 1)
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
			err := ValidateBinaryHTTPResponse(testHTTPResponse(tt.contentType, tt.body), 1)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateBinaryHTTPResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBase64Data(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{name: "valid base64", data: "YQ==", wantErr: false},
		{name: "empty data", data: "", wantErr: true},
		{name: "invalid base64", data: "not-base64!!!", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBase64Data(tt.data, 1)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateBase64Data() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWAVBytes(t *testing.T) {
	if err := ValidateWAVBytes(SmallWAVBytes()); err != nil {
		t.Fatalf("ValidateWAVBytes() on embedded fixture error = %v", err)
	}
	if err := ValidateWAVBytes([]byte("not wav")); err == nil {
		t.Fatal("expected error for invalid WAV bytes")
	}
}

func TestValidateBase64WAVData(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString(SmallWAVBytes())
	if err := ValidateBase64WAVData(valid, 12); err != nil {
		t.Fatalf("ValidateBase64WAVData() on embedded fixture error = %v", err)
	}
	if err := ValidateBase64WAVData("not-base64!!!", 1); err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestEmbeddedFixtures(t *testing.T) {
	if len(SmallPNGBytes()) == 0 {
		t.Fatal("embedded PNG fixture is empty")
	}
	if len(SmallWAVBytes()) < 12 {
		t.Fatal("embedded WAV fixture is too short")
	}
	if string(SmallWAVBytes()[0:4]) != "RIFF" {
		t.Fatal("embedded WAV fixture missing RIFF header")
	}
}
