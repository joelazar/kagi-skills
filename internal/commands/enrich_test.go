package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joelazar/kagi/internal/api"
)

func TestFetchEnrich(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot test-key" {
			t.Errorf("expected 'Bot test-key', got %q", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("q") != "golang" {
			t.Errorf("expected query 'golang', got %q", r.URL.Query().Get("q"))
		}

		snippet := "A Go resource"
		resp := enrichResponse{
			Meta: api.Meta{ID: "test-id", MS: 50},
			Data: []enrichAPIItem{
				{
					T:       0,
					Rank:    1,
					URL:     "https://example.com",
					Title:   "Go Resource",
					Snippet: &snippet,
				},
				{
					T:     0,
					Rank:  2,
					URL:   "https://example2.com",
					Title: "Another Go Resource",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewHTTPClient(5 * time.Second)
	resp, err := fetchEnrich(client, "test-key", server.URL, "golang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "Go Resource" {
		t.Errorf("expected 'Go Resource', got %q", resp.Data[0].Title)
	}
}
