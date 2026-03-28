// Package api provides shared HTTP client and error handling for Kagi APIs.
package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError represents an error returned by the Kagi API.
type APIError struct {
	StatusCode int
	Code       int
	Message    string
	RawBody    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.RawBody)
}

// ParseAPIError attempts to parse a Kagi API error response body.
// Returns an *APIError with structured fields if possible, or raw body as fallback.
func ParseAPIError(statusCode int, body []byte) *APIError {
	var errResp struct {
		Error []struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"error"`
	}

	if json.Unmarshal(body, &errResp) == nil && len(errResp.Error) > 0 {
		return &APIError{
			StatusCode: statusCode,
			Code:       errResp.Error[0].Code,
			Message:    errResp.Error[0].Msg,
		}
	}

	raw := strings.TrimSpace(string(body))
	if len(raw) > 500 {
		raw = raw[:500] + "..."
	}

	return &APIError{
		StatusCode: statusCode,
		RawBody:    raw,
	}
}
