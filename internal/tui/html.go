package tui

import (
	"regexp"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

var (
	reSupTag     = regexp.MustCompile(`<sup[^>]*>.*?</sup>`)
	reDetailsTag = regexp.MustCompile(`(?s)<details>.*?</details>`)
)

// htmlToMarkdown converts HTML-heavy API output into cleaner markdown for TUI detail rendering.
func htmlToMarkdown(html string) string {
	if html == "" {
		return ""
	}

	s := reDetailsTag.ReplaceAllString(html, "")
	s = reSupTag.ReplaceAllString(s, "")

	md, err := htmltomarkdown.ConvertString(s)
	if err != nil {
		return strings.TrimSpace(s)
	}

	return strings.TrimSpace(md)
}
