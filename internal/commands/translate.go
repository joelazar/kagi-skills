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

const (
	translateBootstrapURL = "https://translate.kagi.com/"
	translateDetectURL    = "https://translate.kagi.com/api/detect"
	translateTextURL      = "https://translate.kagi.com/api/translate"
)

type translateDetectResponse struct {
	ISO  string `json:"iso"`
	Name string `json:"name"`
}

type translateTextResponse struct {
	Translation string `json:"translation"`
}

type translateOutput struct {
	Text        string `json:"text"`
	Translation string `json:"translation"`
	SourceLang  string `json:"source_lang,omitempty"`
	TargetLang  string `json:"target_lang"`
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
		Short: "Translate text via Kagi Translate",
		Long: `Translate text using Kagi's translation service.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie).

Text can be provided as arguments or piped via stdin.`,
		Example: `  kagi translate --to de "Hello, world!"
  kagi translate --from en --to ja "Good morning"
  echo "Bonjour le monde" | kagi translate --to en
  kagi translate --to es --formality formal "How are you?"`,
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

			timeout := time.Duration(timeoutSec) * time.Second
			result, err := doTranslate(sessionToken, text, sourceLang, targetLang, formality, timeout)
			if err != nil {
				return err
			}

			return renderTranslateOutput(result)
		},
	}

	cmd.Flags().StringVar(&sourceLang, "from", "auto", "source language code (auto-detect if omitted)")
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
	if out.SourceLang != "" {
		fmt.Fprintf(output.Stderr(), "[detected: %s]\n", out.SourceLang)
	}

	return nil
}

func doTranslate(sessionToken, text, sourceLang, targetLang, formality string, timeout time.Duration) (translateOutput, error) {
	client := api.NewHTTPClient(timeout)

	// Step 1: Bootstrap translate session
	translateSession, err := bootstrapTranslateSession(client, sessionToken)
	if err != nil {
		return translateOutput{}, fmt.Errorf("translate bootstrap failed: %w", err)
	}

	cookieHeader := fmt.Sprintf("kagi_session=%s; translate_session=%s", sessionToken, translateSession)

	// Step 2: Detect source language if auto
	effectiveSource := strings.ToLower(sourceLang)
	detectedName := ""
	if effectiveSource == "" || effectiveSource == "auto" {
		detected, err := detectLanguage(client, cookieHeader, text)
		if err != nil {
			return translateOutput{}, fmt.Errorf("language detection failed: %w", err)
		}
		if detected.ISO != "" {
			effectiveSource = detected.ISO
			detectedName = detected.Name
		}
	}

	// Step 3: Translate
	translation, err := translateText(client, cookieHeader, translateSession, text, effectiveSource, targetLang, formality)
	if err != nil {
		return translateOutput{}, err
	}

	return translateOutput{
		Text:        text,
		Translation: translation,
		SourceLang:  detectedName,
		TargetLang:  targetLang,
	}, nil
}

// bootstrapTranslateSession visits translate.kagi.com to obtain a translate_session cookie.
func bootstrapTranslateSession(client *http.Client, sessionToken string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, translateBootstrapURL, nil)
	if err != nil {
		return "", err
	}

	req.AddCookie(&http.Cookie{Name: "kagi_session", Value: sessionToken})
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	// Don't follow redirects automatically — we need the Set-Cookie header
	noRedirectClient := *client
	noRedirectClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", errors.New("invalid or expired Kagi session token for Kagi Translate")
	}

	// Extract translate_session cookie from response
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "translate_session" {
			return cookie.Value, nil
		}
	}

	return "", errors.New("translate bootstrap did not return a translate_session cookie")
}

func detectLanguage(client *http.Client, cookieHeader, text string) (*translateDetectResponse, error) {
	payload := map[string]any{
		"text":                 text,
		"include_alternatives": true,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, translateDetectURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", cookieHeader)
	req.Header.Set("Content-Type", "application/json")
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
		return nil, fmt.Errorf("language detection HTTP %d: %s", resp.StatusCode, truncateString(string(body), 200))
	}

	var result translateDetectResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse detection response: %w", err)
	}

	return &result, nil
}

func translateText(client *http.Client, cookieHeader, translateSession, text, from, to, formality string) (string, error) {
	payload := map[string]any{
		"text":          text,
		"from":          from,
		"to":            to,
		"stream":        false,
		"session_token": translateSession,
	}
	if formality != "" {
		payload["formality"] = formality
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, translateTextURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Cookie", cookieHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, api.MaxResponseBody))
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("translation HTTP %d: %s", resp.StatusCode, truncateString(string(body), 200))
	}

	var result translateTextResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse translation response: %w", err)
	}

	if result.Translation == "" {
		return "", errors.New("empty translation response")
	}

	return result.Translation, nil
}
