package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

type batchResult struct {
	Query  string        `json:"query"`
	Index  int           `json:"index"`
	Output *searchOutput `json:"output,omitempty"`
	Error  string        `json:"error,omitempty"`
}

type batchOutput struct {
	Results []batchResult `json:"results"`
	Total   int           `json:"total"`
	Errors  int           `json:"errors"`
}

// @lat: [[cli#Search and retrieval]]
func newBatchCmd() *cobra.Command {
	var (
		concurrency int
		limit       int
		timeoutSec  int
		rateLimit   int
	)

	cmd := &cobra.Command{
		Use:   "batch [queries...]",
		Short: "Run parallel searches with concurrency control",
		Long: `Run multiple search queries in parallel with configurable concurrency and rate limiting.
Queries can be provided as arguments or piped via stdin (one per line).`,
		Example: `  kagi batch "query1" "query2" "query3"
  echo -e "query1\nquery2\nquery3" | kagi batch
  cat queries.txt | kagi batch --concurrency 3 --format json`,
		RunE: func(_ *cobra.Command, args []string) error {
			queries := args

			// Read from stdin if no args
			if len(queries) == 0 {
				stat, err := os.Stdin.Stat()
				if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
					scanner := bufio.NewScanner(os.Stdin)
					for scanner.Scan() {
						q := strings.TrimSpace(scanner.Text())
						if q != "" {
							queries = append(queries, q)
						}
					}
					if err := scanner.Err(); err != nil {
						return fmt.Errorf("reading stdin: %w", err)
					}
				}
			}

			if len(queries) == 0 {
				return errors.New("no queries provided; pass as arguments or pipe via stdin")
			}

			apiKey, err := api.ResolveAPIKey(cfg)
			if err != nil {
				return err
			}

			if concurrency < 1 {
				concurrency = 1
			}
			if concurrency > 10 {
				concurrency = 10
			}

			results := runBatchSearches(queries, apiKey, concurrency, limit, timeoutSec, rateLimit)
			return renderBatchOutput(results)
		},
	}

	cmd.Flags().IntVar(&concurrency, "concurrency", 3, "max parallel requests (1-10)")
	cmd.Flags().IntVarP(&limit, "num", "n", 5, "results per query")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 15, "HTTP timeout per request in seconds")
	cmd.Flags().IntVar(&rateLimit, "rate-limit", 0, "delay between requests in ms (0 = no delay)")

	return cmd
}

// @lat: [[overview#Capability Families#API-key commands]]
func runBatchSearches(queries []string, apiKey string, concurrency, limit, timeoutSec, rateLimitMs int) batchOutput {
	results := make([]batchResult, len(queries))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, query := range queries {
		wg.Add(1)
		go func(idx int, q string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if rateLimitMs > 0 && idx > 0 {
				time.Sleep(time.Duration(rateLimitMs) * time.Millisecond)
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := fetchSearch(client, apiKey, q, limit)

			result := batchResult{
				Query: q,
				Index: idx,
			}

			if err != nil {
				result.Error = err.Error()
			} else {
				out := &searchOutput{
					Query:   q,
					Meta:    resp.Meta,
					Results: make([]searchResult, 0, len(resp.Data)),
				}
				for _, item := range resp.Data {
					if item.T == 0 {
						out.Results = append(out.Results, searchResult{
							Title:     item.Title,
							Link:      item.URL,
							Snippet:   item.Snippet,
							Published: item.Published,
							Thumbnail: item.Thumbnail,
						})
					}
				}
				result.Output = out
				_ = api.SaveBalanceCache(resp.Meta, "kagi-batch")
			}

			results[idx] = result
		}(i, query)
	}

	wg.Wait()

	errorCount := 0
	for _, r := range results {
		if r.Error != "" {
			errorCount++
		}
	}

	return batchOutput{
		Results: results,
		Total:   len(queries),
		Errors:  errorCount,
	}
}

func renderBatchOutput(out batchOutput) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

	// Pretty/text — write NDJSON for easy piping
	enc := json.NewEncoder(io.Writer(os.Stdout))
	for _, r := range out.Results {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "[batch: %d queries, %d errors]\n", out.Total, out.Errors)
	return nil
}
