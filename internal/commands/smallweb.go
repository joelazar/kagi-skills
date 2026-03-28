package commands

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const smallWebFeedURL = "https://kagi.com/api/v1/smallweb/feed/"

// Atom feed structures
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	ID        string     `xml:"id"`
	Updated   string     `xml:"updated"`
	Published string     `xml:"published"`
	Link      atomLink   `xml:"link"`
	Author    atomAuthor `xml:"author"`
	Summary   string     `xml:"summary"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type smallWebItem struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Author    string `json:"author,omitempty"`
	Published string `json:"published,omitempty"`
	Summary   string `json:"summary,omitempty"`
}

type smallWebOutput struct {
	Items []smallWebItem `json:"items"`
	Count int            `json:"count"`
}

func newSmallWebCmd() *cobra.Command {
	var (
		limit      int
		timeoutSec int
	)

	cmd := &cobra.Command{
		Use:   "smallweb",
		Short: "Browse Kagi's Small Web feed",
		Long: `Browse recent content from the "small web" — non-commercial content crafted by
individuals to express themselves or share knowledge without seeking financial gain.

This is a free API endpoint — no API key or subscription required.`,
		Example: `  kagi smallweb
  kagi smallweb -n 5
  kagi smallweb --format json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)

			feed, err := fetchSmallWebFeed(client, limit)
			if err != nil {
				return err
			}

			items := make([]smallWebItem, 0, len(feed.Entries))
			for _, entry := range feed.Entries {
				item := smallWebItem{
					Title:     entry.Title,
					URL:       entry.Link.Href,
					Author:    entry.Author.Name,
					Published: entry.Published,
				}
				// Strip HTML from summary for text output
				if entry.Summary != "" {
					item.Summary = stripHTMLTags(entry.Summary)
				}
				items = append(items, item)
			}

			out := smallWebOutput{
				Items: items,
				Count: len(items),
			}

			return renderSmallWebOutput(out)
		},
	}

	cmd.Flags().IntVarP(&limit, "num", "n", 20, "number of entries to fetch")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 15, "HTTP timeout in seconds")

	return cmd
}

func renderSmallWebOutput(out smallWebOutput) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

	if len(out.Items) == 0 {
		fmt.Println("No entries found.")
		return nil
	}

	for i, item := range out.Items {
		fmt.Printf("--- %d ---\n", i+1)
		fmt.Printf("Title:  %s\n", item.Title)
		fmt.Printf("URL:    %s\n", item.URL)
		if item.Author != "" && item.Author != "Unknown author" {
			fmt.Printf("Author: %s\n", item.Author)
		}
		if item.Published != "" {
			fmt.Printf("Date:   %s\n", item.Published)
		}
		if item.Summary != "" {
			summary := item.Summary
			runes := []rune(summary)
			if len(runes) > 200 {
				summary = string(runes[:200]) + "..."
			}
			fmt.Printf("        %s\n", summary)
		}
		fmt.Println()
	}

	return nil
}

func fetchSmallWebFeed(client *http.Client, limit int) (*atomFeed, error) {
	feedURL := smallWebFeedURL
	if limit > 0 {
		feedURL = fmt.Sprintf("%s?limit=%d", smallWebFeedURL, limit)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/atom+xml, application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d fetching Small Web feed", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, api.MaxResponseBody))
	if err != nil {
		return nil, err
	}

	var feed atomFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse Atom feed: %w", err)
	}

	return &feed, nil
}

// stripHTMLTags is a simple HTML tag stripper for feed summaries.
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	// Clean up whitespace
	text := result.String()
	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, " ")
}
