package mockserver

import "net/http"

// mockImageB64JSON is an 8x8 PNG used for image API mock responses.
const mockImageB64JSON = "iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAYAAADED76LAAAAEklEQVR4nGP4n2L0Hx9mGBkKACBDpQFoN/xgAAAAAElFTkSuQmCC"

func handleImagesGenerations(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"created": 1700000000,
		"data": []map[string]any{
			{"b64_json": mockImageB64JSON},
		},
	})
}