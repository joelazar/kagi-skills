package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTranslateResponseParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := translateTextResponse{
			Translation: "Hallo Welt!",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL) //nolint:noctx // test code
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var tResp translateTextResponse
	if err := json.NewDecoder(resp.Body).Decode(&tResp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if tResp.Translation != "Hallo Welt!" {
		t.Errorf("expected 'Hallo Welt!', got %q", tResp.Translation)
	}
}
