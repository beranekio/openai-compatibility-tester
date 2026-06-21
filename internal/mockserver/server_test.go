package mockserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBrokenServerReturnsJSONErrorWithContentType(t *testing.T) {
	server := BrokenServer()
	t.Cleanup(server.Close)

	resp, err := http.Post(server.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"gpt-4o-mini"}`))
	if err != nil {
		t.Fatalf("http.Post() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status code = %d, want 400", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
}

// Handler() must register the same routes as New(); verify a representative
// endpoint (chat completions) responds through a plain net/http server.
func TestHandlerServesChatCompletions(t *testing.T) {
	ts := httptest.NewServer(Handler())
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"gpt-4o-mini"}`))
	if err != nil {
		t.Fatalf("http.Post() error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want 200, body = %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "chat.completion") {
		t.Fatalf("response body = %s, want a chat.completion object", body)
	}
}