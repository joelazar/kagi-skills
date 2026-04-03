package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const (
	enrichWebURL  = "https://kagi.com/api/v0/enrich/web"
	enrichNewsURL = "https://kagi.com/api/v0/enrich/news"
)

type enrichAPIItem struct {
	T         int     `json:"t"`
	Rank      int     `json:"rank,omitempty"`
	URL       string  `json:"url,omitempty"`
	Title     string  `json:"title,omitempty"`
	Snippet   *string `json:"snippet"`
	Published string  `json:"published,omitempty"`
}

type enrichResponse struct {
	Meta api.Meta        `json:"meta"`
	Data []enrichAPIItem `json:"data"`
}

type enrichResult struct {
	Rank      int    `json:"rank"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Snippet   string `json:"snippet,omitempty"`
	Published string `json:"published,omitempty"`
}

type enrichOutput struct {
	Query   string         `json:"query"`
	Index   string         `json:"index"`
	Meta    api.Meta       `json:"meta"`
	Results []enrichResult `json:"results"`
}

// @lat: [[cli#Discovery feeds and enrichment]]
func newEnrichCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enrich",
		Short: "Search Kagi's enrichment indexes",
		Long:  "Search Kagi's Teclis (web) and TinyGem (news) indexes for independent web content and non-mainstream news.",
	}

	cmd.AddCommand(newEnrichSubCmd("web", enrichWebURL, "Teclis — non-commercial, independent web content"))
	cmd.AddCommand(newEnrichSubCmd("news", enrichNewsURL, "TinyGem — non-mainstream news & discussions"))

	return cmd
}

// @lat: [[cli#Discovery feeds and enrichment]]
func newEnrichSubCmd(index, endpoint, description string) *cobra.Command {
	var (
		limit       int
		showBalance bool
		timeoutSec  int
	)

	cmd := &cobra.Command{
		Use:   index + " <query>",
		Short: description,
		Example: fmt.Sprintf(`  kagi enrich %s "golang"
  kagi enrich %s "rust programming" -n 5 --format json`, index, index),
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))

			apiKey, err := api.ResolveAPIKey(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := fetchEnrich(client, apiKey, endpoint, query)
			if err != nil {
				return err
			}
			_ = api.SaveBalanceCache(resp.Meta, "kagi-enrich")

			results := make([]enrichResult, 0, len(resp.Data))
			for _, item := range resp.Data {
				if item.T != 0 {
					continue
				}
				r := enrichResult{
					Rank:      item.Rank,
					Title:     html.UnescapeString(item.Title),
					URL:       item.URL,
					Published: item.Published,
				}
				if item.Snippet != nil {
					r.Snippet = html.UnescapeString(*item.Snippet)
				}
				results = append(results, r)
			}

			sort.Slice(results, func(i, j int) bool {
				return results[i].Rank < results[j].Rank
			})

			if limit > 0 && len(results) > limit {
				results = results[:limit]
			}

			out := enrichOutput{
				Query:   query,
				Index:   index,
				Meta:    resp.Meta,
				Results: results,
			}

			return renderEnrichOutput(out, showBalance)
		},
	}

	cmd.Flags().IntVarP(&limit, "num", "n", 0, "maximum number of results (0 = all)")
	cmd.Flags().BoolVar(&showBalance, "show-balance", false, "print API balance to stderr")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 15, "HTTP timeout in seconds")

	return cmd
}

func renderEnrichOutput(out enrichOutput, showBalance bool) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

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
		fmt.Printf("URL:   %s\n", r.URL)
		if r.Published != "" {
			fmt.Printf("Date:  %s\n", r.Published)
		}
		if r.Snippet != "" {
			fmt.Printf("       %s\n", r.Snippet)
		}
		fmt.Println()
	}

	if showBalance && out.Meta.APIBalance != nil {
		fmt.Fprintf(os.Stderr, "[API Balance: $%.4f | results: %d]\n", *out.Meta.APIBalance, len(out.Results))
	}

	return nil
}

// @lat: [[overview#Capability Families#API-key commands]]
func fetchEnrich(client *http.Client, apiKey, endpoint, query string) (*enrichResponse, error) {
	params := url.Values{}
	params.Set("q", query)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var out enrichResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &out, nil
}
