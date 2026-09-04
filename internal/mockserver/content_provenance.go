package mockserver

import "net/http"

func handleContentProvenanceChecks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"object":     "content_provenance_check",
		"created_at": 1700000000,
		"results": []map[string]any{
			{
				"type":             "c2pa",
				"outcome":          "not_detected",
				"validation_state": "not_present",
				"issuer":           "",
				"model":            "",
				"generated_at":     "",
			},
			{
				"type":         "synthid",
				"outcome":      "not_detected",
				"model":        nil,
				"generated_at": nil,
			},
		},
	})
}
