package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joelazar/kagi/internal/api"
)

func TestFastGPTResponseParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := fastGPTResponse{
			Meta: api.Meta{ID: "test-id"},
			Data: fastGPTData{
				Output: "Test answer",
				Tokens: 42,
				References: []reference{
					{Title: "Ref 1", URL: "https://example.com", Snippet: "snippet"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL) //nolint:noctx // test code
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var fgResp fastGPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&fgResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if fgResp.Data.Output != "Test answer" {
		t.Errorf("expected 'Test answer', got %q", fgResp.Data.Output)
	}
	if fgResp.Data.Tokens != 42 {
		t.Errorf("expected 42 tokens, got %d", fgResp.Data.Tokens)
	}
	if len(fgResp.Data.References) != 1 {
		t.Errorf("expected 1 reference, got %d", len(fgResp.Data.References))
	}
}
