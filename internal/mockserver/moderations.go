package mockserver

import "net/http"

func handleModerations(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"id":    "modr-mock",
		"model": "omni-moderation-latest",
		"results": []map[string]any{
			{
				"flagged": false,
				"categories": map[string]bool{
					"harassment":            false,
					"harassment/threatening": false,
					"hate":                  false,
					"hate/threatening":      false,
					"illicit":               false,
					"illicit/violent":       false,
					"self-harm":             false,
					"self-harm/instructions": false,
					"self-harm/intent":      false,
					"sexual":                false,
					"sexual/minors":         false,
					"violence":              false,
					"violence/graphic":      false,
				},
				"category_scores": map[string]float64{
					"harassment":            0.0001,
					"harassment/threatening": 0.0001,
					"hate":                  0.0001,
					"hate/threatening":      0.0001,
					"illicit":               0.0001,
					"illicit/violent":       0.0001,
					"self-harm":             0.0001,
					"self-harm/instructions": 0.0001,
					"self-harm/intent":      0.0001,
					"sexual":                0.0001,
					"violence":              0.0001,
					"violence/graphic":      0.0001,
					"sexual/minors":         0.0001,
				},
				"category_applied_input_types": map[string][]string{
					"harassment":            {"text"},
					"harassment/threatening": {"text"},
					"hate":                  {"text"},
					"hate/threatening":      {"text"},
					"illicit":               {"text"},
					"illicit/violent":       {"text"},
					"self-harm":             {"text"},
					"self-harm/instructions": {"text"},
					"self-harm/intent":      {"text"},
					"sexual":                {"text"},
					"sexual/minors":         {"text"},
					"violence":              {"text"},
					"violence/graphic":      {"text"},
				},
			},
		},
	})
}