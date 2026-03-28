package commands

import (
	"testing"
)

func TestCleanLine(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello   world  ", "hello world"},
		{"single", "single"},
		{"  ", ""},
		{"", ""},
		{"  multiple   spaces   here  ", "multiple spaces here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanLine(tt.input)
			if got != tt.expected {
				t.Errorf("cleanLine(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
		{"hello", -1, ""},
		{"héllo", 3, "hél"},
	}

	for _, tt := range tests {
		got := truncateRunes(tt.input, tt.limit)
		if got != tt.expected {
			t.Errorf("truncateRunes(%q, %d) = %q, want %q", tt.input, tt.limit, got, tt.expected)
		}
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "basic title",
			html:     `<html><head><title>Test Page</title></head></html>`,
			expected: "Test Page",
		},
		{
			name:     "no title",
			html:     `<html><head></head></html>`,
			expected: "",
		},
		{
			name:     "title with whitespace",
			html:     `<title>  Hello   World  </title>`,
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTitle(tt.html)
			if got != tt.expected {
				t.Errorf("extractTitle() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExtractReadableText(t *testing.T) {
	html := `<html>
<head><script>console.log("test")</script></head>
<body>
<nav>Navigation stuff</nav>
<p>Hello world</p>
<p>Second paragraph</p>
</body>
</html>`

	result := extractReadableText(html)
	if result == "" {
		t.Error("expected non-empty result")
	}
	if !containsSubstring(result, "Hello world") {
		t.Error("expected to find 'Hello world' in result")
	}
	if containsSubstring(result, "console.log") {
		t.Error("expected script content to be stripped")
	}
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
