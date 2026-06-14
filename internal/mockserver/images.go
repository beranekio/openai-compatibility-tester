package mockserver

import "net/http"

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

func handleImagesGenerations(w http.ResponseWriter, _ *http.Request) {
	writeMockImagesResponse(w)
}

func handleImagesEdits(w http.ResponseWriter, _ *http.Request) {
	writeMockImagesResponse(w)
}

func handleImagesVariations(w http.ResponseWriter, _ *http.Request) {
	writeMockImagesResponse(w)
}