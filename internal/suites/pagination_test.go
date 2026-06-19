package suites

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/pagination"
)

func TestValidateCursorListPageRejectsNilPage(t *testing.T) {
	err := validateCursorListPage[openai.FileObject]("files", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "list page is nil") {
		t.Fatalf("expected nil page validation error, got %v", err)
	}
}

func TestValidateCursorListPageRequiresDataField(t *testing.T) {
	page := parseCursorListPage[openai.FileObject](t, `{
		"object": "list",
		"has_more": false
	}`)

	err := validateCursorListPage("files", page, nil)
	if err == nil || !strings.Contains(err.Error(), "list missing data") {
		t.Fatalf("expected missing data validation error, got %v", err)
	}
}

func TestValidateCursorListPageRequiresHasMore(t *testing.T) {
	page := parseCursorListPage[openai.FileObject](t, `{
		"object": "list",
		"data": []
	}`)

	err := validateCursorListPage("files", page, nil)
	if err == nil || !strings.Contains(err.Error(), "list missing has_more") {
		t.Fatalf("expected missing has_more validation error, got %v", err)
	}
}

func TestValidateCursorListPageRequiresListObject(t *testing.T) {
	page := parseCursorListPage[openai.FileObject](t, `{
		"object": "file",
		"data": [],
		"has_more": false
	}`)

	err := validateCursorListPage("files", page, nil)
	if err == nil || !strings.Contains(err.Error(), `list object is "file"`) {
		t.Fatalf("expected wrong object validation error, got %v", err)
	}
}

func TestValidateCursorListPageAllowsEmptyData(t *testing.T) {
	page := parseCursorListPage[openai.FileObject](t, `{
		"object": "list",
		"data": [],
		"has_more": false
	}`)

	if err := validateCursorListPage("files", page, nil); err != nil {
		t.Fatalf("validateCursorListPage() error = %v", err)
	}
}

func TestValidateCursorListPageSkipsCursorFieldsWithoutIDGetter(t *testing.T) {
	page := parseCursorListPage[openai.FileObject](t, `{
		"object": "list",
		"data": [{
			"id": "file_mock",
			"object": "file",
			"bytes": 4,
			"created_at": 1700000000,
			"filename": "mock.txt",
			"purpose": "assistants"
		}],
		"has_more": false
	}`)

	if err := validateCursorListPage("files", page, nil); err != nil {
		t.Fatalf("validateCursorListPage() error = %v", err)
	}
}

func TestValidateCursorListPageRequiresFirstIDWithData(t *testing.T) {
	page := parseCursorListPage[openai.ChatCompletion](t, `{
		"object": "list",
		"data": [{
			"id": "chatcmpl_mock",
			"object": "chat.completion",
			"created": 1700000000,
			"model": "gpt-4o-mini",
			"choices": []
		}],
		"last_id": "chatcmpl_mock",
		"has_more": false
	}`)

	err := validateCursorListPage("chat_completions", page, func(c *openai.ChatCompletion) string { return c.ID })
	if err == nil || !strings.Contains(err.Error(), "list missing first_id") {
		t.Fatalf("expected missing first_id validation error, got %v", err)
	}
}

func TestValidateCursorListPageRequiresLastIDWithData(t *testing.T) {
	page := parseCursorListPage[openai.ChatCompletion](t, `{
		"object": "list",
		"data": [{
			"id": "chatcmpl_mock",
			"object": "chat.completion",
			"created": 1700000000,
			"model": "gpt-4o-mini",
			"choices": []
		}],
		"first_id": "chatcmpl_mock",
		"has_more": false
	}`)

	err := validateCursorListPage("chat_completions", page, func(c *openai.ChatCompletion) string { return c.ID })
	if err == nil || !strings.Contains(err.Error(), "list missing last_id") {
		t.Fatalf("expected missing last_id validation error, got %v", err)
	}
}

func TestValidateCursorListPageRejectsMismatchedLastID(t *testing.T) {
	page := parseCursorListPage[openai.ChatCompletion](t, `{
		"object": "list",
		"data": [{
			"id": "chatcmpl_mock",
			"object": "chat.completion",
			"created": 1700000000,
			"model": "gpt-4o-mini",
			"choices": []
		}],
		"first_id": "chatcmpl_mock",
		"last_id": "chatcmpl_other",
		"has_more": false
	}`)

	err := validateCursorListPage("chat_completions", page, func(c *openai.ChatCompletion) string { return c.ID })
	if err == nil || !strings.Contains(err.Error(), `list last_id is "chatcmpl_other"`) {
		t.Fatalf("expected wrong last_id validation error, got %v", err)
	}
}

func TestValidateCursorListPageAcceptsValidCursorFields(t *testing.T) {
	page := parseCursorListPage[openai.ChatCompletion](t, `{
		"object": "list",
		"data": [{
			"id": "chatcmpl_mock",
			"object": "chat.completion",
			"created": 1700000000,
			"model": "gpt-4o-mini",
			"choices": []
		}],
		"first_id": "chatcmpl_mock",
		"last_id": "chatcmpl_mock",
		"has_more": false
	}`)

	if err := validateCursorListPage("chat_completions", page, func(c *openai.ChatCompletion) string { return c.ID }); err != nil {
		t.Fatalf("validateCursorListPage() error = %v", err)
	}
}

func parseCursorListPage[T any](t *testing.T, raw string) *pagination.CursorPage[T] {
	t.Helper()

	var page pagination.CursorPage[T]
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return &page
}