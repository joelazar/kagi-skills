package commands

import "testing"

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abcdefghijklmnop", "abcd…op"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "1234…89"},
	}

	for _, tt := range tests {
		got := maskKey(tt.input)
		if got != tt.want {
			t.Errorf("maskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
