package suites

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3/packages/pagination"
)

type cursorListEnvelope struct {
	Object  string `json:"object"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
}

// validateCursorListPage checks common fields on pagination.CursorPage list responses:
// non-nil page, valid data and has_more fields, object=="list", and optionally
// first_id/last_id cursor fields when data is non-empty and id is provided.
func validateCursorListPage[T any](suite string, page *pagination.CursorPage[T], id func(T) string) error {
	if page == nil {
		return fail(suite, "list page is nil")
	}
	if !page.JSON.Data.Valid() {
		return fail(suite, "list missing data")
	}
	if !page.JSON.HasMore.Valid() {
		return fail(suite, "list missing has_more")
	}
	var envelope cursorListEnvelope
	if err := json.Unmarshal([]byte(page.RawJSON()), &envelope); err != nil {
		return fail(suite, "list response is not valid JSON")
	}
	if envelope.Object != "list" {
		return fail(suite, fmt.Sprintf("list object is %q, want list", envelope.Object))
	}
	if len(page.Data) == 0 || id == nil {
		return nil
	}
	firstID := id(page.Data[0])
	lastID := id(page.Data[len(page.Data)-1])
	if envelope.FirstID == "" {
		return fail(suite, "list missing first_id")
	}
	if envelope.LastID == "" {
		return fail(suite, "list missing last_id")
	}
	if envelope.FirstID != firstID {
		return fail(suite, fmt.Sprintf("list first_id is %q, want %q", envelope.FirstID, firstID))
	}
	if envelope.LastID != lastID {
		return fail(suite, fmt.Sprintf("list last_id is %q, want %q", envelope.LastID, lastID))
	}
	return nil
}