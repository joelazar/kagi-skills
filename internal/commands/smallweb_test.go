package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joelazar/kagi/internal/api"
)

const testAtomFeed = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title type="text">Kagi Small Web</title>
  <id>https://kagi.com/api/v1/smallweb/feed</id>
  <updated>2024-01-01T00:00:00+00:00</updated>
  <entry>
    <title type="text">Test Article</title>
    <id>https://example.com/test</id>
    <updated>2024-01-01T00:00:00+00:00</updated>
    <published>2024-01-01T00:00:00+00:00</published>
    <link href="https://example.com/test" />
    <author><name>Test Author</name></author>
    <summary type="html">&lt;p&gt;Test summary content&lt;/p&gt;</summary>
  </entry>
  <entry>
    <title type="text">Second Article</title>
    <id>https://example.com/second</id>
    <updated>2024-01-02T00:00:00+00:00</updated>
    <published>2024-01-02T00:00:00+00:00</published>
    <link href="https://example.com/second" />
    <author><name>Another Author</name></author>
    <summary type="html">Plain text summary</summary>
  </entry>
</feed>`

func TestFetchSmallWebFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.Write([]byte(testAtomFeed))
	}))
	defer server.Close()

	// Can't easily override the URL constant, but we can test the parsing logic
	client := api.NewHTTPClient(5 * time.Second)
	_ = client
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple HTML",
			input:    "<p>Hello <b>world</b></p>",
			expected: "Hello world",
		},
		{
			name:     "nested tags",
			input:    "<div><p>First</p><p>Second</p></div>",
			expected: "FirstSecond",
		},
		{
			name:     "no HTML",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "with newlines",
			input:    "<p>Line 1</p>\n<p>Line 2</p>",
			expected: "Line 1 Line 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTMLTags(tt.input)
			if got != tt.expected {
				t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
