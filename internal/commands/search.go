package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const kagiSearchURL = "https://kagi.com/api/v0/search"

type apiThumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
}

type searchAPIItem struct {
	T         int           `json:"t"`
	URL       string        `json:"url,omitempty"`
	Title     string        `json:"title,omitempty"`
	Snippet   string        `json:"snippet,omitempty"`
	Published string        `json:"published,omitempty"`
	Thumbnail *apiThumbnail `json:"thumbnail,omitempty"`
	List      []string      `json:"list,omitempty"`
}

type kagiSearchResponse struct {
	Meta api.Meta        `json:"meta"`
	Data []searchAPIItem `json:"data"`
}

type searchResult struct {
	Title        string        `json:"title"`
	Link         string        `json:"link"`
	Snippet      string        `json:"snippet"`
	Published    string        `json:"published,omitempty"`
	Thumbnail    *apiThumbnail `json:"thumbnail,omitempty"`
	Content      string        `json:"content,omitempty"`
	ContentError string        `json:"content_error,omitempty"`
}

type searchOutput struct {
	Query           string         `json:"query"`
	Meta            api.Meta       `json:"meta"`
	Results         []searchResult `json:"results"`
	RelatedSearches []string       `json:"related_searches,omitempty"`
}

type contentOutput struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

var (
	reComments = regexp.MustCompile(`(?is)<!--.*?-->`)
	reNoise    = regexp.MustCompile(`(?is)<(?:script|style|noscript|svg|iframe|nav|header|footer|aside)[^>]*>.*?</(?:script|style|noscript|svg|iframe|nav|header|footer|aside)>`)
	reBlocks   = regexp.MustCompile(`(?is)</?(p|div|section|article|main|h[1-6]|li|ul|ol|blockquote|pre|tr|table|hr|br)[^>]*>`)
	reTags     = regexp.MustCompile(`(?is)<[^>]+>`)
	reMultiNL  = regexp.MustCompile(`\n{3,}`)
	reTitle    = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
)

// @lat: [[cli#Search and retrieval]]
func newSearchCmd() *cobra.Command {
	var (
		limit           int
		fetchContent    bool
		showBalance     bool
		timeoutSec      int
		maxContentChars int
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the web with Kagi Search",
		Long:  "Search the web using Kagi's Search API, with optional readable page-content extraction.",
		Example: `  kagi search "golang generics"
  kagi search "rust async" -n 5 --content
  kagi search "latest news" --format json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return errors.New("query is required")
			}

			apiKey, err := api.ResolveAPIKey(cfg)
			if err != nil {
				return err
			}

			if limit < 1 {
				limit = 1
			}
			if limit > 100 {
				limit = 100
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := fetchSearch(client, apiKey, query, limit)
			if err != nil {
				return err
			}
			_ = api.SaveBalanceCache(resp.Meta, "kagi-search")

			out := searchOutput{
				Query:   query,
				Meta:    resp.Meta,
				Results: make([]searchResult, 0, len(resp.Data)),
			}

			for _, item := range resp.Data {
				switch item.T {
				case 0:
					out.Results = append(out.Results, searchResult{
						Title:     item.Title,
						Link:      item.URL,
						Snippet:   item.Snippet,
						Published: item.Published,
						Thumbnail: item.Thumbnail,
					})
				case 1:
					out.RelatedSearches = append(out.RelatedSearches, item.List...)
				}
			}

			if fetchContent {
				contentClient := api.NewSafeContentClient(time.Duration(timeoutSec) * time.Second)
				for i := range out.Results {
					title, content, fetchErr := fetchPageContent(contentClient, out.Results[i].Link, maxContentChars)
					if out.Results[i].Title == "" && title != "" {
						out.Results[i].Title = title
					}
					if fetchErr != nil {
						out.Results[i].ContentError = fetchErr.Error()
						continue
					}
					out.Results[i].Content = content
				}
			}

			return renderSearchOutput(out, showBalance)
		},
	}

	cmd.Flags().IntVarP(&limit, "num", "n", 10, "number of results (max 100)")
	cmd.Flags().BoolVar(&fetchContent, "content", false, "fetch readable content for each result")
	cmd.Flags().BoolVar(&showBalance, "show-balance", false, "print API balance to stderr")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 15, "HTTP timeout in seconds")
	cmd.Flags().IntVar(&maxContentChars, "max-content-chars", 5000, "max characters of fetched content per result")

	cmd.AddCommand(newContentCmd())

	return cmd
}

// @lat: [[cli#Search and retrieval]]
func newContentCmd() *cobra.Command {
	var (
		timeoutSec int
		maxChars   int
	)

	cmd := &cobra.Command{
		Use:   "content <url>",
		Short: "Extract readable page content from a URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			targetURL := strings.TrimSpace(args[0])
			parsedURL, err := api.ValidateRemoteFetchURL(targetURL)
			if err != nil {
				return err
			}

			client := api.NewSafeContentClient(time.Duration(timeoutSec) * time.Second)
			title, content, fetchErr := fetchPageContent(client, parsedURL.String(), maxChars)

			format := getFormat()
			if format == output.FormatJSON || format == output.FormatCompact {
				out := contentOutput{
					URL:     targetURL,
					Title:   title,
					Content: content,
				}
				if fetchErr != nil {
					out.Error = fetchErr.Error()
				}
				if format == output.FormatCompact {
					return output.WriteCompact(out)
				}
				return output.WriteJSON(out)
			}

			if fetchErr != nil {
				return fetchErr
			}
			if title != "" {
				fmt.Printf("# %s\n\n", title)
			}
			fmt.Println(content)
			return nil
		},
	}

	cmd.Flags().IntVar(&timeoutSec, "timeout", 20, "HTTP timeout in seconds")
	cmd.Flags().IntVar(&maxChars, "max-chars", 20000, "max characters to output")

	return cmd
}

func renderSearchOutput(out searchOutput, showBalance bool) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

	// Pretty/text output
	if len(out.Results) == 0 {
		fmt.Fprintln(os.Stderr, "No results found.")
		if showBalance && out.Meta.APIBalance != nil {
			fmt.Fprintf(os.Stderr, "[API Balance: $%.4f]\n", *out.Meta.APIBalance)
		}
		return nil
	}

	for i, r := range out.Results {
		fmt.Printf("--- Result %d ---\n", i+1)
		fmt.Printf("Title: %s\n", r.Title)
		fmt.Printf("Link: %s\n", r.Link)
		if r.Published != "" {
			fmt.Printf("Published: %s\n", r.Published)
		}
		fmt.Printf("Snippet: %s\n", r.Snippet)
		if r.Content != "" {
			fmt.Printf("Content:\n%s\n", r.Content)
		} else if r.ContentError != "" {
			fmt.Printf("Content: (Error: %s)\n", r.ContentError)
		}
		fmt.Println()
	}

	printRelatedSearches(out.RelatedSearches)

	if showBalance && out.Meta.APIBalance != nil {
		fmt.Fprintf(os.Stderr, "[API Balance: $%.4f]\n", *out.Meta.APIBalance)
	}

	return nil
}

func printRelatedSearches(terms []string) {
	if len(terms) == 0 {
		return
	}
	sorted := append([]string(nil), terms...)
	sort.Strings(sorted)
	fmt.Println("--- Related Searches ---")
	for _, term := range sorted {
		fmt.Printf("- %s\n", term)
	}
}

// @lat: [[overview#Capability Families#API-key commands]]
func fetchSearch(client *http.Client, apiKey, query string, limit int) (*kagiSearchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, kagiSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("User-Agent", api.DefaultUserAgent)
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var out kagiSearchResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// @lat: [[architecture#HTTP clients and safety boundaries]]
func fetchPageContent(client *http.Client, targetURL string, maxChars int) (title string, content string, err error) {
	parsedURL, err := api.ValidateRemoteFetchURL(targetURL)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", api.DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, api.MaxContentBody))
	if err != nil {
		return "", "", err
	}

	htmlDoc := string(body)
	title, content = tryReadability(htmlDoc, parsedURL.String())
	if title == "" {
		title = extractTitle(htmlDoc)
	}
	if content == "" {
		content = extractReadableText(htmlDoc)
	}

	if strings.TrimSpace(content) == "" {
		return title, "", errors.New("could not extract readable content")
	}

	if maxChars > 0 {
		content = truncateRunes(content, maxChars)
	}
	return title, content, nil
}

func tryReadability(htmlDoc, targetURL string) (title, content string) {
	pageURL, err := url.Parse(targetURL)
	if err != nil {
		return
	}
	article, err := readability.FromReader(strings.NewReader(htmlDoc), pageURL)
	if err != nil {
		return
	}
	if t := cleanLine(article.Title()); t != "" {
		title = t
	}
	var sb strings.Builder
	if err := article.RenderText(&sb); err != nil {
		return
	}
	content = strings.TrimSpace(sb.String())
	return
}

func extractTitle(htmlDoc string) string {
	matches := reTitle.FindStringSubmatch(htmlDoc)
	if len(matches) < 2 {
		return ""
	}
	return cleanLine(matches[1])
}

func extractReadableText(htmlDoc string) string {
	s := reComments.ReplaceAllString(htmlDoc, " ")
	s = reNoise.ReplaceAllString(s, "\n")
	s = reBlocks.ReplaceAllString(s, "\n")
	s = reTags.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\r", "")

	lines := strings.Split(s, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = cleanLine(line)
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}

	if len(cleaned) == 0 {
		return ""
	}

	joined := strings.Join(cleaned, "\n\n")
	joined = reMultiNL.ReplaceAllString(joined, "\n\n")
	return strings.TrimSpace(joined)
}

func cleanLine(s string) string {
	fields := strings.Fields(strings.TrimSpace(s))
	return strings.Join(fields, " ")
}

func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit])
}

func truncateString(s string, maxLen int) string { //nolint:unparam // generic helper
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
