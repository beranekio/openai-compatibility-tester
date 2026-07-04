package mockserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// mockImageB64JSON is an 8x8 PNG used for image API mock responses.
const mockImageB64JSON = "iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAYAAADED76LAAAAEklEQVR4nGP4n2L0Hx9mGBkKACBDpQFoN/xgAAAAAElFTkSuQmCC"

func writeMockImagesResponse(w http.ResponseWriter) {
	writeJSON(w, map[string]any{
		"created": 1700000000,
		"data": []map[string]any{
			{"b64_json": mockImageB64JSON},
		},
	})
}

func handleImagesGenerations(w http.ResponseWriter, r *http.Request) {
	if isImagesGenerateStream(r) {
		writeImageStream(w, "image_generation")
		return
	}
	writeMockImagesResponse(w)
}

func handleImagesEdits(w http.ResponseWriter, r *http.Request) {
	if isImagesEditStream(r) {
		writeImageStream(w, "image_edit")
		return
	}
	writeMockImagesResponse(w)
}

func handleImagesVariations(w http.ResponseWriter, _ *http.Request) {
	writeMockImagesResponse(w)
}

// isImagesGenerateStream reports whether the JSON generations request asked
// for streaming (the SDK injects "stream": true via WithJSONSet).
func isImagesGenerateStream(r *http.Request) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}
	var req struct {
		Stream bool `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}
	return req.Stream
}

// isImagesEditStream reports whether the multipart edits request asked for
// streaming (the SDK adds a "stream" form field set to "true").
func isImagesEditStream(r *http.Request) bool {
	if err := r.ParseMultipartForm(1 << 20); err != nil || r.MultipartForm == nil {
		return false
	}
	defer r.MultipartForm.RemoveAll()
	return firstFormValue(r.MultipartForm.Value["stream"]) == "true"
}

// writeImageStream emits a partial-image event followed by a completed event
// for the given event prefix ("image_generation" or "image_edit").
func writeImageStream(w http.ResponseWriter, prefix string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(payload map[string]any) {
		data, _ := json.Marshal(payload)
		_, _ = w.Write(bytes.Join([][]byte{[]byte("data: "), data, []byte("\n\n")}, nil))
	}

	writeEvent(map[string]any{
		"type":                prefix + ".partial_image",
		"b64_json":            mockImageB64JSON,
		"partial_image_index": 1,
		"created_at":          1700000000,
		"output_format":       "png",
		"size":                "1024x1024",
	})
	writeEvent(map[string]any{
		"type":          prefix + ".completed",
		"b64_json":      mockImageB64JSON,
		"created_at":    1700000000,
		"output_format": "png",
		"size":          "1024x1024",
	})
}
