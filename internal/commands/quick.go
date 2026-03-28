package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const quickAnswerURL = "https://kagi.com/api/v1/quick_answer"

type quickAnswerResponse struct {
	Answer     string           `json:"answer"`
	References []quickReference `json:"references,omitempty"`
	Cached     bool             `json:"cached,omitempty"`
}

type quickReference struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

type quickOutput struct {
	Query      string           `json:"query"`
	Answer     string           `json:"answer"`
	References []quickReference `json:"references,omitempty"`
}

func newQuickCmd() *cobra.Command {
	var timeoutSec int

	cmd := &cobra.Command{
		Use:   "quick <query>",
		Short: "Get a quick answer via Kagi subscriber session",
		Long: `Get an AI-generated quick answer using your Kagi subscriber session.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie or token URL).

This uses Kagi's subscriber Quick Answer feature, not the paid FastGPT API.`,
		Example: `  kagi quick "What is the population of Tokyo?"
  kagi quick "How to reverse a string in Go?" --format json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return errors.New("query is required")
			}

			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := fetchQuickAnswer(client, sessionToken, query)
			if err != nil {
				return err
			}

			out := quickOutput{
				Query:      query,
				Answer:     resp.Answer,
				References: resp.References,
			}

			return renderQuickOutput(out)
		},
	}

	cmd.Flags().IntVar(&timeoutSec, "timeout", 30, "HTTP timeout in seconds")

	return cmd
}

func renderQuickOutput(out quickOutput) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

	fmt.Println(out.Answer)

	if len(out.References) > 0 {
		fmt.Println()
		fmt.Println("--- References ---")
		for i, ref := range out.References {
			fmt.Printf("[%d] %s\n", i+1, ref.Title)
			fmt.Printf("    %s\n", ref.URL)
			if ref.Snippet != "" {
				fmt.Printf("    %s\n", ref.Snippet)
			}
		}
	}

	return nil
}

func fetchQuickAnswer(client *http.Client, sessionToken, query string) (*quickAnswerResponse, error) {
	params := url.Values{}
	params.Set("q", query)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, quickAnswerURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	// Session token can be a full URL or just the token value
	token := sessionToken
	if strings.Contains(token, "token=") {
		if u, err := url.Parse(token); err == nil {
			if t := u.Query().Get("token"); t != "" {
				token = t
			}
		}
	}

	req.AddCookie(&http.Cookie{
		Name:  "kagi_session",
		Value: token,
	})
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

	var out quickAnswerResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if out.Answer == "" {
		return nil, errors.New("empty response from Quick Answer")
	}

	return &out, nil
}

func truncateString(s string, maxLen int) string { //nolint:unparam // generic helper
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
