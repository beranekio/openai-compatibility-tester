package suites

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

func parseAPIError(t *testing.T, status int, payload string) *openai.Error {
	t.Helper()

	var wrapper struct {
		Error openai.Error `json:"error"`
	}
	if err := json.Unmarshal([]byte(payload), &wrapper); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	wrapper.Error.StatusCode = status
	return &wrapper.Error
}

func TestValidateErrorResponseAPIError(t *testing.T) {
	valid := parseAPIError(t, http.StatusBadRequest, `{
		"error": {
			"message": "The model oct-invalid-model does not exist",
			"type": "invalid_request_error",
			"param": "model",
			"code": "model_not_found"
		}
	}`)
	if err := validateErrorResponseAPIError("error_responses", valid); err != nil {
		t.Fatalf("validateErrorResponseAPIError() error = %v", err)
	}

	tests := []struct {
		name    string
		apiErr  *openai.Error
		wantErr string
	}{
		{
			name: "422 with model code",
			apiErr: parseAPIError(t, http.StatusUnprocessableEntity, `{
				"error": {
					"message": "invalid request",
					"type": "invalid_request_error",
					"code": "invalid_model"
				}
			}`),
		},
		{
			name: "message mentions model",
			apiErr: parseAPIError(t, http.StatusBadRequest, `{
				"error": {
					"message": "Unknown model requested",
					"type": "invalid_request_error"
				}
			}`),
		},
		{
			name: "missing message",
			apiErr: &openai.Error{
				StatusCode: http.StatusBadRequest,
				Type:       "invalid_request_error",
			},
			wantErr: "error missing message",
		},
		{
			name: "401 auth failure",
			apiErr: parseAPIError(t, http.StatusUnauthorized, `{
				"error": {
					"message": "Incorrect API key provided",
					"type": "invalid_request_error",
					"code": "invalid_api_key"
				}
			}`),
			wantErr: "401",
		},
		{
			name: "429 rate limit",
			apiErr: parseAPIError(t, http.StatusTooManyRequests, `{
				"error": {
					"message": "Rate limit reached for model",
					"type": "rate_limit_error"
				}
			}`),
			wantErr: "429",
		},
		{
			name: "generic 400 without model evidence",
			apiErr: parseAPIError(t, http.StatusBadRequest, `{
				"error": {
					"message": "bad request",
					"type": "invalid_request_error"
				}
			}`),
			wantErr: "model-specific evidence",
		},
		{
			name: "5xx server error",
			apiErr: parseAPIError(t, http.StatusInternalServerError, `{
				"error": {
					"message": "The model oct-invalid-model does not exist",
					"type": "server_error"
				}
			}`),
			wantErr: "want 4xx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateErrorResponseAPIError("error_responses", tt.apiErr)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateErrorResponseAPIError() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateErrorResponseAPIError() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}