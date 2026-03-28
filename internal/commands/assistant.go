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

const assistantURL = "https://kagi.com/api/v1/assistant"

type assistantRequest struct {
	Query    string `json:"query"`
	ThreadID string `json:"thread_id,omitempty"`
}

type assistantResponse struct {
	Output     string               `json:"output"`
	ThreadID   string               `json:"thread_id,omitempty"`
	References []assistantReference `json:"references,omitempty"`
}

type assistantReference struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type assistantOutput struct {
	Query      string               `json:"query"`
	Output     string               `json:"output"`
	ThreadID   string               `json:"thread_id,omitempty"`
	References []assistantReference `json:"references,omitempty"`
}

type threadListOutput struct {
	Threads []threadSummary `json:"threads"`
}

type threadSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func newAssistantCmd() *cobra.Command {
	var (
		threadID   string
		timeoutSec int
	)

	cmd := &cobra.Command{
		Use:   "assistant <query>",
		Short: "Chat with Kagi Assistant",
		Long: `Ask questions and get AI-powered answers from Kagi Assistant.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie or token URL).

Supports conversation threads via --thread flag.`,
		Example: `  kagi assistant "Explain quantum computing"
  kagi assistant "Tell me more" --thread <thread-id>
  kagi assistant "Compare Go and Rust" --format json`,
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
			resp, err := callAssistant(client, sessionToken, query, threadID)
			if err != nil {
				return err
			}

			out := assistantOutput{
				Query:      query,
				Output:     resp.Output,
				ThreadID:   resp.ThreadID,
				References: resp.References,
			}

			return renderAssistantOutput(out)
		},
	}

	cmd.Flags().StringVar(&threadID, "thread", "", "continue an existing conversation thread")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 60, "HTTP timeout in seconds")

	cmd.AddCommand(newAssistantThreadCmd())

	return cmd
}

func newAssistantThreadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thread",
		Short: "Manage assistant conversation threads",
	}

	cmd.AddCommand(
		newAssistantThreadListCmd(),
		newAssistantThreadGetCmd(),
		newAssistantThreadDeleteCmd(),
	)

	return cmd
}

func newAssistantThreadListCmd() *cobra.Command {
	var timeoutSec int

	return &cobra.Command{
		Use:   "list",
		Short: "List conversation threads",
		RunE: func(_ *cobra.Command, _ []string) error {
			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			threads, err := listAssistantThreads(client, sessionToken)
			if err != nil {
				return err
			}

			format := getFormat()
			if format == output.FormatJSON {
				return output.WriteJSON(threadListOutput{Threads: threads})
			}
			if format == output.FormatCompact {
				return output.WriteCompact(threadListOutput{Threads: threads})
			}

			if len(threads) == 0 {
				fmt.Println("No threads found.")
				return nil
			}

			for _, t := range threads {
				fmt.Printf("ID:    %s\n", t.ID)
				fmt.Printf("Title: %s\n", t.Title)
				if t.UpdatedAt != "" {
					fmt.Printf("Date:  %s\n", t.UpdatedAt)
				}
				fmt.Println()
			}
			return nil
		},
	}
}

func newAssistantThreadGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <thread-id>",
		Short: "Get a conversation thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(30 * time.Second)
			thread, err := getAssistantThread(client, sessionToken, args[0])
			if err != nil {
				return err
			}

			format := getFormat()
			if format == output.FormatJSON || format == output.FormatCompact {
				if format == output.FormatCompact {
					return output.WriteCompact(thread)
				}
				return output.WriteJSON(thread)
			}

			fmt.Printf("Thread: %s\n\n", args[0])
			for _, msg := range thread.Messages {
				fmt.Printf("[%s] %s\n\n", msg.Role, msg.Content)
			}
			return nil
		},
	}
}

func newAssistantThreadDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <thread-id>",
		Short: "Delete a conversation thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			client := api.NewHTTPClient(15 * time.Second)
			if err := deleteAssistantThread(client, sessionToken, args[0]); err != nil {
				return err
			}

			fmt.Printf("Thread %s deleted.\n", args[0])
			return nil
		},
	}
}

func renderAssistantOutput(out assistantOutput) error {
	format := getFormat()

	if format == output.FormatJSON {
		return output.WriteJSON(out)
	}
	if format == output.FormatCompact {
		return output.WriteCompact(out)
	}

	fmt.Println(out.Output)

	if len(out.References) > 0 {
		fmt.Println()
		fmt.Println("--- References ---")
		for i, ref := range out.References {
			fmt.Printf("[%d] %s\n    %s\n", i+1, ref.Title, ref.URL)
		}
	}

	if out.ThreadID != "" {
		fmt.Fprintf(output.Stderr(), "[thread: %s]\n", out.ThreadID)
	}

	return nil
}

func resolveSessionCookie(sessionToken string) *http.Cookie {
	token := sessionToken
	if strings.Contains(token, "token=") {
		if u, err := url.Parse(token); err == nil {
			if t := u.Query().Get("token"); t != "" {
				token = t
			}
		}
	}
	return &http.Cookie{
		Name:  "kagi_session",
		Value: token,
	}
}

func callAssistant(client *http.Client, sessionToken, query, threadID string) (*assistantResponse, error) {
	reqBody := assistantRequest{
		Query:    query,
		ThreadID: threadID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, assistantURL, bytes.NewReader(bodyBytes))
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

	var out assistantResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &out, nil
}

type threadDetail struct {
	ID       string          `json:"id"`
	Messages []threadMessage `json:"messages"`
}

type threadMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func listAssistantThreads(client *http.Client, sessionToken string) ([]threadSummary, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, assistantURL+"/threads", nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(resolveSessionCookie(sessionToken))
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

	var threads []threadSummary
	if err := json.Unmarshal(body, &threads); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return threads, nil
}

func getAssistantThread(client *http.Client, sessionToken, threadID string) (*threadDetail, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, assistantURL+"/threads/"+threadID, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(resolveSessionCookie(sessionToken))
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

	var thread threadDetail
	if err := json.Unmarshal(body, &thread); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &thread, nil
}

func deleteAssistantThread(client *http.Client, sessionToken, threadID string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, assistantURL+"/threads/"+threadID, nil)
	if err != nil {
		return err
	}

	req.AddCookie(resolveSessionCookie(sessionToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, api.MaxResponseBody))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateString(string(body), 200))
	}

	return nil
}
