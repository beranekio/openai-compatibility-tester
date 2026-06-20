package suites

import (
	"strings"
	"testing"
)

func TestValidateWeatherToolArguments(t *testing.T) {
	if err := validateWeatherToolArguments("responses_tools", `{"location":"San Francisco, CA"}`); err != nil {
		t.Fatalf("validateWeatherToolArguments() error = %v", err)
	}
	tests := []struct {
		name    string
		args    string
		wantErr string
	}{
		{name: "empty", args: "", wantErr: "tool call missing arguments"},
		{name: "invalid json", args: "not-json", wantErr: "tool call arguments are not valid JSON"},
		{name: "missing location", args: "{}", wantErr: `tool call arguments missing required "location"`},
		{name: "empty location", args: `{"location":""}`, wantErr: `"location" field is empty`},
		{name: "non-string location", args: `{"location":1}`, wantErr: `"location" field is not a string`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWeatherToolArguments("responses_tools", tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateWeatherToolArguments() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}