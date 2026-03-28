package commands

import (
	"strings"
	"testing"
)

func TestHtmlToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "empty",
			input:    "",
			contains: nil,
		},
		{
			name:     "plain text",
			input:    "hello world",
			contains: []string{"hello world"},
		},
		{
			name:     "paragraph",
			input:    "<p>Hello</p><p>World</p>",
			contains: []string{"Hello", "World"},
		},
		{
			name:     "heading",
			input:    "<h3>Title</h3><p>Content</p>",
			contains: []string{"### Title", "Content"},
		},
		{
			name:     "list",
			input:    "<ul><li>one</li><li>two</li></ul>",
			contains: []string{"one", "two"},
		},
		{
			name:     "strips details blocks",
			input:    "<details><summary>Search</summary>context stuff</details><p>Answer here</p>",
			contains: []string{"Answer here"},
			excludes: []string{"context stuff"},
		},
		{
			name:     "strips sup citations",
			input:    `<p>Fact<sup><a href="http://example.com">1</a></sup></p>`,
			contains: []string{"Fact"},
			excludes: []string{"<sup>"},
		},
		{
			name:     "link",
			input:    `<a href="https://example.com">Example</a>`,
			contains: []string{"Example", "https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := htmlToMarkdown(tt.input)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("htmlToMarkdown() missing %q in output:\n%s", want, got)
				}
			}
			for _, bad := range tt.excludes {
				if strings.Contains(got, bad) {
					t.Errorf("htmlToMarkdown() should not contain %q in output:\n%s", bad, got)
				}
			}
		})
	}
}
