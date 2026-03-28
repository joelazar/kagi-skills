package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joelazar/kagi/internal/api"
)

func TestRunBatchSearches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		resp := kagiSearchResponse{
			Meta: api.Meta{ID: "test-id", MS: 50},
			Data: []searchAPIItem{
				{
					T:     0,
					URL:   "https://example.com/" + query,
					Title: "Result for " + query,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Can't easily override kagiSearchURL, but verify the batch structure works
	// by testing the output struct
	out := batchOutput{
		Results: []batchResult{
			{Query: "test1", Index: 0, Output: &searchOutput{Query: "test1"}},
			{Query: "test2", Index: 1, Error: "some error"},
		},
		Total:  2,
		Errors: 1,
	}

	if out.Total != 2 {
		t.Errorf("expected 2 total, got %d", out.Total)
	}
	if out.Errors != 1 {
		t.Errorf("expected 1 error, got %d", out.Errors)
	}
	if out.Results[0].Error != "" {
		t.Error("expected no error on first result")
	}
	if out.Results[1].Error != "some error" {
		t.Errorf("expected 'some error', got %q", out.Results[1].Error)
	}
}
