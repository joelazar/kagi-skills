package commands

import (
	"regexp"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

var (
	reSupTag     = regexp.MustCompile(`<sup[^>]*>.*?</sup>`)
	reDetailsTag = regexp.MustCompile(`(?s)<details>.*?</details>`)
)

// htmlToMarkdown converts HTML content to clean markdown for terminal display.
func htmlToMarkdown(html string) string {
	if html == "" {
		return ""
	}

	// Remove <details> blocks (search context, not useful in terminal output)
	s := reDetailsTag.ReplaceAllString(html, "")

	// Remove superscript citation references (e.g., <sup><a href="...">1</a></sup>)
	s = reSupTag.ReplaceAllString(s, "")

	md, err := htmltomarkdown.ConvertString(s)
	if err != nil {
		// Fall back to basic tag stripping on error
		return strings.TrimSpace(s)
	}

	return strings.TrimSpace(md)
}
