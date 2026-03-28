package commands

import (
	"bytes"
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

const askPageURL = "https://kagi.com/api/v1/ask_page"

type askPageRequest struct {
	URL      string `json:"url"`
	Question string `json:"question"`
}

type askPageResponse struct {
	Answer     string               `json:"answer"`
	References []assistantReference `json:"references,omitempty"`
}

type askPageOutput struct {
	URL        string               `json:"url"`
	Question   string               `json:"question"`
	Answer     string               `json:"answer"`
	References []assistantReference `json:"references,omitempty"`
}

func newAskPageCmd() *cobra.Command {
	var timeoutSec int

	cmd := &cobra.Command{
		Use:   "askpage <url> <question>",
		Short: "Ask a question about a specific URL",
		Long: `Ask a question about the content of a specific URL using Kagi's AI.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie or token URL).`,
		Example: `  kagi askpage https://golang.org/doc/go1.22 "What are the new features?"
  kagi askpage https://arxiv.org/abs/1706.03762 "Summarize the main contribution" --format json`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			targetURL := strings.TrimSpace(args[0])
			question := strings.TrimSpace(strings.Join(args[1:], " "))

			if targetURL == "" {
				return errors.New("URL is required")
			}
			if question == "" {
				return errors.New("question is required")
			}

			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := callAskPage(client, sessionToken, targetURL, question)
			if err != nil {
				return err
			}

			out := askPageOutput{
				URL:        targetURL,
				Question:   question,
				Answer:     resp.Answer,
				References: resp.References,
			}

			return renderAskPageOutput(out)
		},
	}

	cmd.Flags().IntVar(&timeoutSec, "timeout", 60, "HTTP timeout in seconds")

	return cmd
}

func renderAskPageOutput(out askPageOutput) error {
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
			fmt.Printf("[%d] %s\n    %s\n", i+1, ref.Title, ref.URL)
		}
	}

	return nil
}

func callAskPage(client *http.Client, sessionToken, targetURL, question string) (*askPageResponse, error) {
	reqBody := askPageRequest{
		URL:      targetURL,
		Question: question,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, askPageURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.AddCookie(resolveSessionCookie(sessionToken))
	req.Header.Set("Content-Type", "application/json")
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

	var out askPageResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if out.Answer == "" {
		return nil, errors.New("empty response from Ask Page")
	}

	return &out, nil
}
