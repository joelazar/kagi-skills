package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const newsURL = "https://kagi.com/api/v1/news"

type newsItem struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Source    string `json:"source,omitempty"`
	Published string `json:"published,omitempty"`
	Snippet   string `json:"snippet,omitempty"`
	Category  string `json:"category,omitempty"`
}

type newsResponse struct {
	Items []newsItem `json:"items"`
}

type newsOutput struct {
	Category string     `json:"category,omitempty"`
	Items    []newsItem `json:"items"`
	Count    int        `json:"count"`
}

func newNewsCmd() *cobra.Command {
	var (
		category   string
		limit      int
		timeoutSec int
	)

	cmd := &cobra.Command{
		Use:   "news",
		Short: "Browse Kagi news feed",
		Long: `Browse news from Kagi's curated news feed with category filtering.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie or token URL).

Categories: world, business, technology, science, health, sports, entertainment`,
		Example: `  kagi news
  kagi news --category technology
  kagi news --category science -n 5
  kagi news --format json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := fetchNews(client, sessionToken, category, limit)
			if err != nil {
				return err
			}

			out := newsOutput{
				Category: category,
				Items:    resp.Items,
				Count:    len(resp.Items),
			}

			return renderNewsOutput(out)
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "news category: world, business, technology, science, health, sports, entertainment")
	cmd.Flags().IntVarP(&limit, "num", "n", 20, "number of items")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 15, "HTTP timeout in seconds")

	return cmd
}

func renderNewsOutput(out newsOutput) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

	if len(out.Items) == 0 {
		fmt.Println("No news items found.")
		return nil
	}

	for i, item := range out.Items {
		fmt.Printf("--- %d ---\n", i+1)
		fmt.Printf("Title:  %s\n", item.Title)
		fmt.Printf("URL:    %s\n", item.URL)
		if item.Source != "" {
			fmt.Printf("Source: %s\n", item.Source)
		}
		if item.Published != "" {
			fmt.Printf("Date:   %s\n", item.Published)
		}
		if item.Category != "" {
			fmt.Printf("Cat:    %s\n", item.Category)
		}
		if item.Snippet != "" {
			snippet := item.Snippet
			runes := []rune(snippet)
			if len(runes) > 200 {
				snippet = string(runes[:200]) + "..."
			}
			fmt.Printf("        %s\n", snippet)
		}
		fmt.Println()
	}

	return nil
}

func fetchNews(client *http.Client, sessionToken, category string, limit int) (*newsResponse, error) {
	params := url.Values{}
	if category != "" {
		params.Set("category", strings.ToLower(category))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	reqURL := newsURL
	if len(params) > 0 {
		reqURL = newsURL + "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(resolveSessionCookie(sessionToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, api.MaxResponseBody))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateString(string(body), 200))
	}

	var out newsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &out, nil
}
