package commands

import (
	"os"
	"regexp"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
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

// renderMarkdownForTerminal renders markdown with glamour styling for terminal output.
// Falls back to plain markdown if glamour fails or stdout is not a terminal.
func renderMarkdownForTerminal(md string) string {
	if md == "" {
		return ""
	}

	if !term.IsTerminal(int(os.Stdout.Fd())) { //nolint:gosec
		return md
	}

	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 { //nolint:gosec
		width = w
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return md
	}

	rendered, err := r.Render(md)
	if err != nil {
		return md
	}

	return strings.TrimRight(rendered, "\n")
}
