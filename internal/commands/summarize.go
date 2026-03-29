package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

const summarizerURL = "https://kagi.com/api/v0/summarize"

type summarizeRequest struct {
	URL            string `json:"url,omitempty"`
	Text           string `json:"text,omitempty"`
	Engine         string `json:"engine,omitempty"`
	SummaryType    string `json:"summary_type,omitempty"`
	TargetLanguage string `json:"target_language,omitempty"`
	Cache          *bool  `json:"cache,omitempty"`
}

type summarizeData struct {
	Output string `json:"output"`
	Tokens int    `json:"tokens"`
}

type summarizeResponse struct {
	Meta api.Meta      `json:"meta"`
	Data summarizeData `json:"data"`
}

type summarizeOutputJSON struct {
	Input  string   `json:"input"`
	Output string   `json:"output"`
	Tokens int      `json:"tokens"`
	Engine string   `json:"engine,omitempty"`
	Type   string   `json:"type,omitempty"`
	Meta   api.Meta `json:"meta"`
}

var (
	validEngines = map[string]bool{
		"cecil":  true,
		"agnes":  true,
		"daphne": true,
		"muriel": true,
	}
	validSummaryTypes = map[string]bool{
		"summary":  true,
		"takeaway": true,
	}
)

func newSummarizeCmd() *cobra.Command {
	var (
		inputText   string
		engine      string
		summType    string
		targetLang  string
		noCache     bool
		showBalance bool
		timeoutSec  int
	)

	cmd := &cobra.Command{
		Use:   "summarize [url]",
		Short: "Summarize a URL or text with Kagi Summarizer",
		Long: `Summarize any URL or text using Kagi's Universal Summarizer API.
Supports multiple engines and summary types. Text can be provided with --text or via stdin.`,
		Example: `  kagi summarize https://en.wikipedia.org/wiki/Go_(programming_language)
  kagi summarize https://arxiv.org/abs/1706.03762 --engine muriel --type takeaway
  kagi summarize --text "Long article text here..." --format json
  echo "text to summarize" | kagi summarize`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var inputURL string
			if len(args) == 1 {
				inputURL = strings.TrimSpace(args[0])
			}

			// Check stdin if no URL and no --text
			if inputURL == "" && inputText == "" {
				stat, err := os.Stdin.Stat()
				if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
					stdinBytes, err := io.ReadAll(io.LimitReader(os.Stdin, 4<<20))
					if err != nil {
						return fmt.Errorf("reading stdin: %w", err)
					}
					inputText = strings.TrimSpace(string(stdinBytes))
				}
			}

			if inputURL == "" && inputText == "" {
				return errors.New("a URL or text input is required")
			}
			if inputURL != "" && inputText != "" {
				return errors.New("--text and a URL are mutually exclusive")
			}

			if engine != "" && !validEngines[engine] {
				return fmt.Errorf("unknown engine %q — valid: cecil, agnes, daphne, muriel", engine)
			}
			if summType != "" && !validSummaryTypes[summType] {
				return fmt.Errorf("unknown type %q — valid: summary, takeaway", summType)
			}

			apiKey, err := api.ResolveAPIKey(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := callSummarizer(client, apiKey, inputURL, inputText, engine, summType, targetLang, !noCache)
			if err != nil {
				return err
			}
			_ = api.SaveBalanceCache(resp.Meta, "kagi-summarizer")

			inputLabel := inputURL
			if inputLabel == "" {
				runes := []rune(inputText)
				if len(runes) > 80 {
					inputLabel = string(runes[:80]) + "..."
				} else {
					inputLabel = inputText
				}
			}

			format := getFormat()
			if format == output.FormatJSON || format == output.FormatCompact {
				out := summarizeOutputJSON{
					Input:  inputLabel,
					Output: resp.Data.Output,
					Tokens: resp.Data.Tokens,
					Engine: engine,
					Type:   summType,
					Meta:   resp.Meta,
				}
				if format == output.FormatCompact {
					return output.WriteCompact(out)
				}
				return output.WriteJSON(out)
			}

			// Pretty/text output
			fmt.Println(resp.Data.Output)

			if showBalance && resp.Meta.APIBalance != nil {
				fmt.Fprintf(os.Stderr, "[API Balance: $%.4f | tokens: %d]\n", *resp.Meta.APIBalance, resp.Data.Tokens)
			} else {
				fmt.Fprintf(os.Stderr, "[tokens: %d]\n", resp.Data.Tokens)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&inputText, "text", "", "summarize raw text instead of a URL")
	cmd.Flags().StringVar(&engine, "engine", "", "summarizer engine: cecil (default), agnes, daphne, muriel")
	cmd.Flags().StringVar(&summType, "type", "", "summary type: summary (default), takeaway")
	cmd.Flags().StringVar(&targetLang, "lang", "", "output language code (for example: EN, DE, FR, JA)")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "bypass cached responses")
	cmd.Flags().BoolVar(&showBalance, "show-balance", false, "print API balance to stderr")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 120, "HTTP timeout in seconds")

	return cmd
}

func callSummarizer(
	client *http.Client,
	apiKey, inputURL, inputText, engine, summType, targetLang string,
	cache bool,
) (*summarizeResponse, error) {
	reqBody := summarizeRequest{
		URL:            inputURL,
		Text:           inputText,
		Engine:         engine,
		SummaryType:    summType,
		TargetLanguage: targetLang,
	}
	if !cache {
		reqBody.Cache = new(bool)
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, summarizerURL, bytes.NewReader(bodyBytes))
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

	var out summarizeResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if out.Data.Output == "" {
		return nil, errors.New("empty response from Summarizer API")
	}

	return &out, nil
}
