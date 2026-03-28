package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const fastGPTURL = "https://kagi.com/api/v0/fastgpt"

type fastGPTRequest struct {
	Query     string `json:"query"`
	Cache     bool   `json:"cache"`
	WebSearch bool   `json:"web_search"`
}

type reference struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

type fastGPTData struct {
	Output     string      `json:"output"`
	Tokens     int         `json:"tokens"`
	References []reference `json:"references"`
}

type fastGPTResponse struct {
	Meta api.Meta    `json:"meta"`
	Data fastGPTData `json:"data"`
}

type fastGPTOutputJSON struct {
	Query      string      `json:"query"`
	Output     string      `json:"output"`
	Tokens     int         `json:"tokens"`
	References []reference `json:"references,omitempty"`
	Meta       api.Meta    `json:"meta"`
}

func newFastGPTCmd() *cobra.Command {
	var (
		noRefs      bool
		noCache     bool
		showBalance bool
		timeoutSec  int
	)

	cmd := &cobra.Command{
		Use:   "fastgpt <query>",
		Short: "Get AI-synthesized answers via Kagi FastGPT",
		Long:  "Ask a question and get an AI-synthesized answer backed by live web search results.",
		Example: `  kagi fastgpt "What is the capital of France?"
  kagi fastgpt "How does Go garbage collection work?" --format json
  kagi fastgpt "Latest Go release" --no-cache`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return errors.New("query is required")
			}

			apiKey, err := api.ResolveAPIKey(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := callFastGPT(client, apiKey, query, !noCache)
			if err != nil {
				return err
			}
			_ = api.SaveBalanceCache(resp.Meta, "kagi-fastgpt")

			format := getFormat()
			if format == output.FormatJSON || format == output.FormatCompact {
				out := fastGPTOutputJSON{
					Query:      query,
					Output:     resp.Data.Output,
					Tokens:     resp.Data.Tokens,
					References: resp.Data.References,
					Meta:       resp.Meta,
				}
				if noRefs {
					out.References = nil
				}
				if format == output.FormatCompact {
					return output.WriteCompact(out)
				}
				return output.WriteJSON(out)
			}

			// Pretty/text output
			fmt.Println(resp.Data.Output)

			if !noRefs && len(resp.Data.References) > 0 {
				fmt.Println()
				fmt.Println("--- References ---")
				for i, ref := range resp.Data.References {
					fmt.Printf("[%d] %s\n", i+1, ref.Title)
					fmt.Printf("    %s\n", ref.URL)
					if ref.Snippet != "" {
						fmt.Printf("    %s\n", ref.Snippet)
					}
				}
			}

			if showBalance && resp.Meta.APIBalance != nil {
				fmt.Fprintf(os.Stderr, "[API Balance: $%.4f | tokens: %d]\n", *resp.Meta.APIBalance, resp.Data.Tokens)
			} else {
				fmt.Fprintf(os.Stderr, "[tokens: %d]\n", resp.Data.Tokens)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noRefs, "no-refs", false, "suppress references in text output")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "bypass cached responses")
	cmd.Flags().BoolVar(&showBalance, "show-balance", false, "print API balance to stderr")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 30, "HTTP timeout in seconds")

	return cmd
}

func callFastGPT(client *http.Client, apiKey, query string, cache bool) (*fastGPTResponse, error) {
	reqBody := fastGPTRequest{
		Query:     query,
		Cache:     cache,
		WebSearch: true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fastGPTURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		return nil, err
	}

	var out fastGPTResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if out.Data.Output == "" {
		return nil, errors.New("empty response from FastGPT API")
	}

	return &out, nil
}
