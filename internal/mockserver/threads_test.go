package mockserver

import "testing"

func TestThreadMessageTextFromContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content any
		want    string
	}{
		{name: "plain string", content: "pong", want: "pong"},
		{name: "flat value map", content: map[string]any{"value": "pong"}, want: "pong"},
		{
			name: "text part map",
			content: map[string]any{
				"type": "text",
				"text": map[string]any{"value": "pong"},
			},
			want: "pong",
		},
		{
			name: "content parts array",
			content: []any{
				map[string]any{
					"type": "text",
					"text": map[string]any{"value": "Reply with exactly the word: pong."},
				},
			},
			want: "Reply with exactly the word: pong.",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := threadMessageTextFromContent(tc.content); got != tc.want {
				t.Fatalf("threadMessageTextFromContent() = %q, want %q", got, tc.want)
			}
		})
	}
}