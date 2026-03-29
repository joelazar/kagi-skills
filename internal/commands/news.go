package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const (
	newsLatestURL     = "https://news.kagi.com/api/batches/latest"
	newsCategoriesURL = "https://news.kagi.com/api/categories/metadata"
	newsBatchURL      = "https://news.kagi.com/api/batches"
)

type newsLatestBatch struct {
	ID string `json:"id"`
}

type newsBatchCategoriesResponse struct {
	Categories []newsBatchCategory `json:"categories"`
}

type newsBatchCategory struct {
	ID           string `json:"id"`
	CategoryID   string `json:"categoryId"`
	CategoryName string `json:"categoryName"`
}

type newsStory struct {
	Title     string `json:"title"`
	URL       string `json:"link,omitempty"`
	Domain    string `json:"domain,omitempty"`
	Published string `json:"date,omitempty"`
	Category  string `json:"category,omitempty"`
}

type newsCluster struct {
	Title        string      `json:"title"`
	Category     string      `json:"category"`
	ShortSummary string      `json:"short_summary"`
	Articles     []newsStory `json:"articles"`
}

type newsStoriesResponse struct {
	Stories []newsCluster `json:"stories"`
}

type newsOutputItem struct {
	Title    string `json:"title"`
	Summary  string `json:"summary,omitempty"`
	Category string `json:"category,omitempty"`
	URL      string `json:"url,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type newsOutput struct {
	Category string           `json:"category,omitempty"`
	Items    []newsOutputItem `json:"items"`
	Count    int              `json:"count"`
}

func newNewsCmd() *cobra.Command {
	var (
		category   string
		limit      int
		lang       string
		timeoutSec int
	)

	cmd := &cobra.Command{
		Use:   "news",
		Short: "Browse Kagi's curated news feed",
		Long: `Browse news from Kagi's curated news feed, with optional category filtering.

This uses the public news.kagi.com API and does not require authentication.

Common categories: world, business, technology, science, health, sports, entertainment`,
		Example: `  kagi news
  kagi news --category technology
  kagi news --category science -n 5
  kagi news --format json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			items, err := fetchNews(client, category, limit, lang)
			if err != nil {
				return err
			}

			out := newsOutput{
				Category: category,
				Items:    items,
				Count:    len(items),
			}

			return renderNewsOutput(out)
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "news category name or slug (for example: technology, science, world)")
	cmd.Flags().IntVarP(&limit, "num", "n", 20, "maximum number of items")
	cmd.Flags().StringVar(&lang, "lang", "en", "language code (e.g., en, de, fr)")
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
		if item.URL != "" {
			fmt.Printf("URL:    %s\n", item.URL)
		}
		if item.Domain != "" {
			fmt.Printf("Source: %s\n", item.Domain)
		}
		if item.Category != "" {
			fmt.Printf("Cat:    %s\n", item.Category)
		}
		if item.Summary != "" {
			summary := item.Summary
			runes := []rune(summary)
			if len(runes) > 300 {
				summary = string(runes[:300]) + "..."
			}
			fmt.Printf("        %s\n", summary)
		}
		fmt.Println()
	}

	return nil
}

func fetchNews(client *http.Client, category string, limit int, lang string) ([]newsOutputItem, error) {
	if lang == "" {
		lang = "en"
	}

	// Step 1: Get latest batch
	var latest newsLatestBatch
	if err := fetchJSON(client, newsLatestURL+"?lang="+lang, &latest); err != nil {
		return nil, fmt.Errorf("fetching latest news batch: %w", err)
	}

	// Step 2: Get batch categories
	var batchCats newsBatchCategoriesResponse
	batchCatsURL := fmt.Sprintf("%s/%s/categories?lang=%s", newsBatchURL, latest.ID, lang)
	if err := fetchJSON(client, batchCatsURL, &batchCats); err != nil {
		return nil, fmt.Errorf("fetching batch categories: %w", err)
	}

	// Step 3: Resolve category
	targetCat := resolveNewsCategory(batchCats.Categories, category)
	if targetCat == "" && category != "" {
		available := make([]string, 0, len(batchCats.Categories))
		for _, c := range batchCats.Categories {
			available = append(available, c.CategoryName)
		}
		return nil, fmt.Errorf("unknown category %q; available: %s", category, strings.Join(available, ", "))
	}

	// Step 4: Determine which category to fetch
	fetchCatID := targetCat
	if fetchCatID == "" && len(batchCats.Categories) > 0 {
		// Default to first category
		fetchCatID = batchCats.Categories[0].ID
	}
	if fetchCatID == "" {
		return nil, errors.New("no news categories available")
	}

	// Step 5: Fetch stories (response is an array of clusters)
	storiesURL := fmt.Sprintf("%s/%s/categories/%s/stories?lang=%s", newsBatchURL, latest.ID, fetchCatID, lang)
	var clusters newsStoriesResponse
	if err := fetchJSON(client, storiesURL, &clusters); err != nil {
		return nil, fmt.Errorf("fetching news stories: %w", err)
	}

	// Convert clusters to output items
	var items []newsOutputItem
	for _, cluster := range clusters.Stories {
		item := newsOutputItem{
			Title:    cluster.Title,
			Summary:  cluster.ShortSummary,
			Category: cluster.Category,
		}
		// Use first article's URL if available
		if len(cluster.Articles) > 0 {
			item.URL = cluster.Articles[0].URL
			item.Domain = cluster.Articles[0].Domain
		}
		items = append(items, item)
	}

	// Apply limit
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

func resolveNewsCategory(categories []newsBatchCategory, requested string) string {
	if requested == "" {
		return ""
	}
	lower := strings.ToLower(requested)
	for _, c := range categories {
		if strings.EqualFold(c.CategoryName, requested) || strings.EqualFold(c.CategoryID, requested) ||
			strings.Contains(strings.ToLower(c.CategoryName), lower) {
			return c.ID
		}
	}
	return ""
}

func fetchJSON(client *http.Client, url string, target any) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, api.MaxResponseBody))
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateString(string(body), 200))
	}

	return json.Unmarshal(body, target)
}
