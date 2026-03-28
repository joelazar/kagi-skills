package commands

import (
	"testing"
)

func TestResolveSessionCookie(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain token",
			input:    "abc123",
			expected: "abc123",
		},
		{
			name:     "token URL",
			input:    "https://kagi.com/search?token=xyz789",
			expected: "xyz789",
		},
		{
			name:     "token URL with other params",
			input:    "https://kagi.com/search?foo=bar&token=abc456&baz=qux",
			expected: "abc456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := resolveSessionCookie(tt.input)
			if cookie.Name != "kagi_session" {
				t.Errorf("expected cookie name 'kagi_session', got %q", cookie.Name)
			}
			if cookie.Value != tt.expected {
				t.Errorf("expected cookie value %q, got %q", tt.expected, cookie.Value)
			}
		})
	}
}
