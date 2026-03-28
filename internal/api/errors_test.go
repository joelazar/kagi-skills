package api_test

import (
	"net/http"
	"testing"

	"github.com/joelazar/kagi/internal/api"
)

func TestParseAPIError_StructuredError(t *testing.T) {
	body := []byte(`{"error":[{"code":401,"msg":"Unauthorized: invalid API key"}]}`)
	err := api.ParseAPIError(http.StatusUnauthorized, body)

	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", err.StatusCode)
	}
	if err.Code != http.StatusUnauthorized {
		t.Errorf("expected code 401, got %d", err.Code)
	}
	if err.Message != "Unauthorized: invalid API key" {
		t.Errorf("unexpected message: %s", err.Message)
	}
	expected := "HTTP 401: Unauthorized: invalid API key"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestParseAPIError_RawBody(t *testing.T) {
	body := []byte(`Internal Server Error`)
	err := api.ParseAPIError(http.StatusInternalServerError, body)

	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", err.StatusCode)
	}
	if err.Message != "" {
		t.Errorf("expected empty message, got %q", err.Message)
	}
	if err.RawBody != "Internal Server Error" {
		t.Errorf("unexpected raw body: %s", err.RawBody)
	}
}

func TestParseAPIError_TruncatesLongBody(t *testing.T) {
	body := make([]byte, 600)
	for i := range body {
		body[i] = 'x'
	}
	err := api.ParseAPIError(http.StatusInternalServerError, body)

	if len(err.RawBody) != 503 { // 500 + "..."
		t.Errorf("expected truncated body length 503, got %d", len(err.RawBody))
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      api.APIError
		expected string
	}{
		{
			name:     "with message",
			err:      api.APIError{StatusCode: http.StatusForbidden, Message: "Forbidden"},
			expected: "HTTP 403: Forbidden",
		},
		{
			name:     "with raw body",
			err:      api.APIError{StatusCode: http.StatusInternalServerError, RawBody: "error text"},
			expected: "HTTP 500: error text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
