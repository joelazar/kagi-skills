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

const translateURL = "https://kagi.com/api/v1/translate"

type translateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang,omitempty"`
	TargetLang string `json:"target_lang"`
	Formality  string `json:"formality,omitempty"`
}

type translateResponse struct {
	Translation  string   `json:"translation"`
	SourceLang   string   `json:"source_lang,omitempty"`
	TargetLang   string   `json:"target_lang,omitempty"`
	Alternatives []string `json:"alternatives,omitempty"`
	Suggestions  []string `json:"suggestions,omitempty"`
}

type translateOutput struct {
	Text         string   `json:"text"`
	Translation  string   `json:"translation"`
	SourceLang   string   `json:"source_lang,omitempty"`
	TargetLang   string   `json:"target_lang"`
	Alternatives []string `json:"alternatives,omitempty"`
	Suggestions  []string `json:"suggestions,omitempty"`
}

func newTranslateCmd() *cobra.Command {
	var (
		sourceLang string
		targetLang string
		formality  string
		timeoutSec int
	)

	cmd := &cobra.Command{
		Use:   "translate <text>",
		Short: "Translate text via Kagi",
		Long: `Translate text using Kagi's translation service.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie or token URL).

Text can be provided as arguments or piped via stdin.`,
		Example: `  kagi translate --to DE "Hello, world!"
  kagi translate --from EN --to JA "Good morning"
  echo "Bonjour le monde" | kagi translate --to EN
  kagi translate --to ES --formality formal "How are you?"`,
		RunE: func(_ *cobra.Command, args []string) error {
			text := strings.TrimSpace(strings.Join(args, " "))

			// Read from stdin if no args
			if text == "" {
				stat, err := os.Stdin.Stat()
				if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
					stdinBytes, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))
					if err != nil {
						return fmt.Errorf("reading stdin: %w", err)
					}
					text = strings.TrimSpace(string(stdinBytes))
				}
			}

			if text == "" {
				return errors.New("text to translate is required")
			}
			if targetLang == "" {
				return errors.New("--to (target language) is required")
			}

			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := callTranslate(client, sessionToken, text, sourceLang, targetLang, formality)
			if err != nil {
				return err
			}

			out := translateOutput{
				Text:         text,
				Translation:  resp.Translation,
				SourceLang:   resp.SourceLang,
				TargetLang:   resp.TargetLang,
				Alternatives: resp.Alternatives,
				Suggestions:  resp.Suggestions,
			}

			return renderTranslateOutput(out)
		},
	}

	cmd.Flags().StringVar(&sourceLang, "from", "", "source language code (auto-detect if omitted)")
	cmd.Flags().StringVar(&targetLang, "to", "", "target language code (required)")
	cmd.Flags().StringVar(&formality, "formality", "", "formality level: formal, informal")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 30, "HTTP timeout in seconds")

	_ = cmd.MarkFlagRequired("to")

	return cmd
}

func renderTranslateOutput(out translateOutput) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

	fmt.Println(out.Translation)

	if len(out.Alternatives) > 0 {
		fmt.Println()
		fmt.Println("--- Alternatives ---")
		for _, alt := range out.Alternatives {
			fmt.Printf("- %s\n", alt)
		}
	}

	if len(out.Suggestions) > 0 {
		fmt.Println()
		fmt.Println("--- Suggestions ---")
		for _, sug := range out.Suggestions {
			fmt.Printf("- %s\n", sug)
		}
	}

	return nil
}

func callTranslate(client *http.Client, sessionToken, text, sourceLang, targetLang, formality string) (*translateResponse, error) {
	reqBody := translateRequest{
		Text:       text,
		SourceLang: strings.ToUpper(sourceLang),
		TargetLang: strings.ToUpper(targetLang),
		Formality:  formality,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, translateURL, bytes.NewReader(bodyBytes))
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

	var out translateResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if out.Translation == "" {
		return nil, errors.New("empty response from translation service")
	}

	return &out, nil
}
