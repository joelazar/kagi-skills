package commands

import (
	"bytes"
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

const quickAnswerURL = "https://kagi.com/mother/context"

type quickMessagePayload struct {
	ID                string   `json:"id"`
	ThreadID          string   `json:"thread_id"`
	CreatedAt         string   `json:"created_at,omitempty"`
	State             string   `json:"state"`
	Prompt            string   `json:"prompt"`
	Reply             string   `json:"reply,omitempty"`
	Markdown          string   `json:"md,omitempty"`
	ReferencesMD      string   `json:"references_md,omitempty"`
	FollowupQuestions []string `json:"followup_questions,omitempty"`
}

type quickOutput struct {
	Query      string `json:"query"`
	Answer     string `json:"answer"`
	References string `json:"references,omitempty"`
}

func newQuickCmd() *cobra.Command {
	var timeoutSec int

	cmd := &cobra.Command{
		Use:   "quick <query>",
		Short: "Get a quick answer via Kagi subscriber session",
		Long: `Get an AI-generated quick answer using your Kagi subscriber session.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie).

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

			return renderQuickOutput(resp)
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

	if out.References != "" {
		fmt.Println()
		fmt.Println("--- References ---")
		fmt.Println(out.References)
	}

	return nil
}

func fetchQuickAnswer(client *http.Client, sessionToken, query string) (quickOutput, error) {
	params := url.Values{}
	params.Set("q", query)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, quickAnswerURL+"?"+params.Encode(), bytes.NewReader([]byte{}))
	if err != nil {
		return quickOutput{}, err
	}

	req.AddCookie(resolveSessionCookie(sessionToken))
	req.Header.Set("Accept", kagiStreamAccept)
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Cache-Control", "no-store")
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return quickOutput{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, api.MaxResponseBody))
	if err != nil {
		return quickOutput{}, err
	}

	bodyStr := string(body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return quickOutput{}, errors.New("invalid or expired Kagi session token")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return quickOutput{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateString(bodyStr, 200))
	}

	// Check for HTML response (auth redirect)
	trimmed := strings.TrimSpace(bodyStr)
	if strings.HasPrefix(trimmed, "<!") || strings.HasPrefix(trimmed, "<html") {
		return quickOutput{}, errors.New("invalid or expired Kagi session token")
	}

	return parseQuickAnswerStream(bodyStr, query)
}

func parseQuickAnswerStream(body, query string) (quickOutput, error) {
	frames := parseStreamFrames(body)

	// Check for errors first
	if _, ok := frames["unauthorized"]; ok {
		return quickOutput{}, errors.New("invalid or expired Kagi session token")
	}
	if notices, ok := frames["limit_notice.html"]; ok && len(notices) > 0 {
		return quickOutput{}, fmt.Errorf("kagi quick answer unavailable: %s", notices[0])
	}

	// Parse message
	msgFrames, ok := frames["new_message.json"]
	if !ok || len(msgFrames) == 0 {
		return quickOutput{}, errors.New("quick answer response did not include a new_message.json frame")
	}

	var msg quickMessagePayload
	if err := json.Unmarshal([]byte(msgFrames[0]), &msg); err != nil {
		return quickOutput{}, fmt.Errorf("failed to parse quick answer message: %w", err)
	}

	if msg.State == "error" {
		errText := msg.Markdown
		if errText == "" {
			errText = msg.Reply
		}
		if errText == "" {
			errText = "Kagi Quick Answer returned an error"
		}
		return quickOutput{}, errors.New(errText)
	}

	answer := msg.Markdown
	if answer == "" {
		answer = msg.Reply
	}
	// Fall back to accumulated tokens.json content
	if answer == "" {
		answer = accumulateTokens(frames)
	}

	return quickOutput{
		Query:      query,
		Answer:     answer,
		References: msg.ReferencesMD,
	}, nil
}
