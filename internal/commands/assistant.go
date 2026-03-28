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

const (
	assistantPromptURL       = "https://kagi.com/assistant/prompt"
	assistantThreadListURL   = "https://kagi.com/assistant/thread_list"
	assistantThreadOpenURL   = "https://kagi.com/assistant/thread_open"
	assistantThreadDeleteURL = "https://kagi.com/assistant/thread_delete"
	assistantZeroBranchUUID  = "00000000-0000-4000-0000-000000000000"
	kagiStreamAccept         = "application/vnd.kagi.stream"
)

// --- Stream frame types ---

type assistantHello struct {
	Version string `json:"v"`
	Trace   string `json:"trace"`
}

type assistantThreadPayload struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Ack       string   `json:"ack,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	ExpiresAt string   `json:"expires_at,omitempty"`
	Saved     bool     `json:"saved"`
	Shared    bool     `json:"shared"`
	BranchID  string   `json:"branch_id,omitempty"`
	TagIDs    []string `json:"tag_ids,omitempty"`
}

type assistantMessagePayload struct {
	ID             string   `json:"id"`
	ThreadID       string   `json:"thread_id"`
	CreatedAt      string   `json:"created_at,omitempty"`
	BranchList     []string `json:"branch_list,omitempty"`
	State          string   `json:"state"`
	Prompt         string   `json:"prompt"`
	Reply          string   `json:"reply,omitempty"`
	Markdown       string   `json:"md,omitempty"`
	ReferencesMD   string   `json:"references_md,omitempty"`
	ReferencesHTML string   `json:"references_html,omitempty"`
	TraceID        string   `json:"trace_id,omitempty"`
}

type assistantThreadListPayload struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Saved     bool   `json:"saved"`
	Shared    bool   `json:"shared"`
}

// --- Response types ---

type assistantPromptResponse struct {
	Trace      string `json:"trace,omitempty"`
	ThreadID   string `json:"thread_id"`
	Title      string `json:"title,omitempty"`
	Output     string `json:"output"`
	References string `json:"references,omitempty"`
}

type assistantOutput struct {
	Query      string `json:"query"`
	ThreadID   string `json:"thread_id"`
	Title      string `json:"title,omitempty"`
	Output     string `json:"output"`
	References string `json:"references,omitempty"`
}

type threadListOutput struct {
	Threads []threadSummary `json:"threads"`
}

type threadSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at,omitempty"`
	Saved     bool   `json:"saved"`
}

type threadDetail struct {
	ID       string                `json:"id"`
	Title    string                `json:"title,omitempty"`
	Messages []threadDetailMessage `json:"messages"`
}

type threadDetailMessage struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	Prompt    string `json:"prompt,omitempty"`
	Reply     string `json:"reply,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
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
Requires KAGI_SESSION_TOKEN (your Kagi session cookie).

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
			resp, err := callAssistantPrompt(client, sessionToken, query, threadID)
			if err != nil {
				return err
			}

			out := assistantOutput{
				Query:      query,
				ThreadID:   resp.ThreadID,
				Title:      resp.Title,
				Output:     resp.Output,
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

	cmd := &cobra.Command{
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
				if t.CreatedAt != "" {
					fmt.Printf("Date:  %s\n", t.CreatedAt)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&timeoutSec, "timeout", 30, "HTTP timeout in seconds")

	return cmd
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

			fmt.Printf("Thread: %s\n", thread.ID)
			if thread.Title != "" {
				fmt.Printf("Title:  %s\n", thread.Title)
			}
			fmt.Println()
			for _, msg := range thread.Messages {
				if msg.Prompt != "" {
					fmt.Printf("[user] %s\n\n", msg.Prompt)
				}
				if msg.Reply != "" {
					fmt.Printf("[assistant] %s\n\n", msg.Reply)
				}
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

			client := api.NewHTTPClient(30 * time.Second)

			// First, fetch the thread to get its details (needed for delete payload)
			thread, err := getAssistantThread(client, sessionToken, args[0])
			if err != nil {
				return fmt.Errorf("failed to fetch thread for deletion: %w", err)
			}

			if err := deleteAssistantThread(client, sessionToken, thread); err != nil {
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

	if out.References != "" {
		fmt.Println()
		fmt.Println("--- References ---")
		fmt.Println(out.References)
	}

	if out.ThreadID != "" {
		fmt.Fprintf(output.Stderr(), "[thread: %s]\n", out.ThreadID)
	}

	return nil
}

// --- Session cookie ---

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

// --- Kagi stream protocol ---

// executeKagiStream sends a POST request using Kagi's streaming protocol and returns the raw body.
func executeKagiStream(client *http.Client, url, sessionToken string, payload any) (string, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.AddCookie(resolveSessionCookie(sessionToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", kagiStreamAccept)
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

	bodyStr := string(body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", errors.New("invalid or expired Kagi session token")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateString(bodyStr, 200))
	}

	// Check for HTML response (indicates auth redirect)
	if strings.HasPrefix(strings.TrimSpace(bodyStr), "<!") || strings.HasPrefix(strings.TrimSpace(bodyStr), "<html") {
		return "", errors.New("invalid or expired Kagi session token")
	}

	return bodyStr, nil
}

// parseStreamFrames splits a Kagi stream body into tag:payload pairs.
func parseStreamFrames(body string) map[string][]string {
	frames := make(map[string][]string)
	for frame := range strings.SplitSeq(body, "\x00\n") {
		frame = strings.TrimSpace(frame)
		if frame == "" {
			continue
		}
		tag, payload, ok := strings.Cut(frame, ":")
		if !ok {
			continue
		}
		frames[tag] = append(frames[tag], payload)
	}
	return frames
}

// --- Assistant API calls ---

func callAssistantPrompt(client *http.Client, sessionToken, query, threadID string) (*assistantPromptResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("assistant query cannot be empty")
	}

	var tidValue any
	if tid := strings.TrimSpace(threadID); tid != "" {
		tidValue = tid
	}

	payload := map[string]any{
		"focus": map[string]any{
			"thread_id":  tidValue,
			"branch_id":  assistantZeroBranchUUID,
			"prompt":     query,
			"message_id": nil,
		},
	}

	body, err := executeKagiStream(client, assistantPromptURL, sessionToken, payload)
	if err != nil {
		return nil, err
	}

	return parseAssistantPromptStream(body)
}

func parseAssistantPromptStream(body string) (*assistantPromptResponse, error) {
	frames := parseStreamFrames(body)

	resp := &assistantPromptResponse{}

	// Parse hello frame
	if hellos, ok := frames["hi"]; ok && len(hellos) > 0 {
		var hello assistantHello
		if err := json.Unmarshal([]byte(hellos[0]), &hello); err == nil {
			resp.Trace = hello.Trace
		}
	}

	// Parse thread frame
	if threads, ok := frames["thread.json"]; ok && len(threads) > 0 {
		var thread assistantThreadPayload
		if err := json.Unmarshal([]byte(threads[0]), &thread); err != nil {
			return nil, fmt.Errorf("failed to parse assistant thread frame: %w", err)
		}
		resp.ThreadID = thread.ID
		resp.Title = thread.Title
	} else {
		return nil, errors.New("assistant response did not include a thread.json frame")
	}

	// Parse message frame
	if err := parseAssistantMessage(frames, resp); err != nil {
		return nil, err
	}

	// Check for limit notice
	if notices, ok := frames["limit_notice.html"]; ok && len(notices) > 0 {
		return nil, fmt.Errorf("kagi assistant rate limited: %s", notices[0])
	}

	// Check for unauthorized
	if _, ok := frames["unauthorized"]; ok {
		return nil, errors.New("invalid or expired Kagi session token")
	}

	return resp, nil
}

func parseAssistantMessage(frames map[string][]string, resp *assistantPromptResponse) error {
	messages, ok := frames["new_message.json"]
	if !ok || len(messages) == 0 {
		return errors.New("assistant response did not include a new_message.json frame")
	}

	var msg assistantMessagePayload
	if err := json.Unmarshal([]byte(messages[0]), &msg); err != nil {
		return fmt.Errorf("failed to parse assistant message frame: %w", err)
	}

	if msg.State == "error" {
		errText := msg.Markdown
		if errText == "" {
			errText = msg.Reply
		}
		if errText == "" {
			errText = "Kagi Assistant returned an error"
		}
		return errors.New(errText)
	}

	resp.Output = msg.Markdown
	if resp.Output == "" {
		resp.Output = msg.Reply
	}

	// If both md and reply are empty, accumulate from tokens.json frames
	if resp.Output == "" {
		resp.Output = accumulateTokens(frames)
	}

	resp.References = msg.ReferencesMD

	return nil
}

// accumulateTokens extracts accumulated HTML content from tokens.json frames.
// The last tokens.json frame contains the full accumulated HTML.
func accumulateTokens(frames map[string][]string) string {
	tokenFrames, ok := frames["tokens.json"]
	if !ok || len(tokenFrames) == 0 {
		return ""
	}

	// The last tokens.json frame should have the full accumulated content
	last := tokenFrames[len(tokenFrames)-1]

	var token struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(last), &token); err != nil {
		return ""
	}

	return token.Text
}

func listAssistantThreads(client *http.Client, sessionToken string) ([]threadSummary, error) {
	payload := map[string]any{"limit": 100}

	body, err := executeKagiStream(client, assistantThreadListURL, sessionToken, payload)
	if err != nil {
		return nil, err
	}

	return parseAssistantThreadListStream(body)
}

func parseAssistantThreadListStream(body string) ([]threadSummary, error) {
	frames := parseStreamFrames(body)

	threadFrames, ok := frames["thread.json"]
	if !ok {
		return []threadSummary{}, nil
	}

	threads := make([]threadSummary, 0, len(threadFrames))
	for _, raw := range threadFrames {
		var t assistantThreadListPayload
		if err := json.Unmarshal([]byte(raw), &t); err != nil {
			continue
		}
		threads = append(threads, threadSummary{
			ID:        t.ID,
			Title:     t.Title,
			CreatedAt: t.CreatedAt,
			Saved:     t.Saved,
		})
	}

	return threads, nil
}

func getAssistantThread(client *http.Client, sessionToken, threadID string) (*threadDetail, error) {
	tid := strings.TrimSpace(threadID)
	if tid == "" {
		return nil, errors.New("thread ID is required")
	}

	payload := map[string]any{
		"focus": map[string]any{
			"thread_id": tid,
			"branch_id": assistantZeroBranchUUID,
		},
	}

	body, err := executeKagiStream(client, assistantThreadOpenURL, sessionToken, payload)
	if err != nil {
		return nil, err
	}

	return parseAssistantThreadOpenStream(body, tid)
}

func parseAssistantThreadOpenStream(body, threadID string) (*threadDetail, error) {
	frames := parseStreamFrames(body)

	detail := &threadDetail{ID: threadID}

	// Parse thread info
	if threads, ok := frames["thread.json"]; ok && len(threads) > 0 {
		var t assistantThreadPayload
		if err := json.Unmarshal([]byte(threads[0]), &t); err == nil {
			detail.Title = t.Title
			detail.ID = t.ID
		}
	}

	// Parse messages
	if msgFrames, ok := frames["message.json"]; ok {
		for _, raw := range msgFrames {
			var msg assistantMessagePayload
			if err := json.Unmarshal([]byte(raw), &msg); err != nil {
				continue
			}
			reply := msg.Markdown
			if reply == "" {
				reply = msg.Reply
			}
			detail.Messages = append(detail.Messages, threadDetailMessage{
				ID:        msg.ID,
				State:     msg.State,
				Prompt:    msg.Prompt,
				Reply:     reply,
				CreatedAt: msg.CreatedAt,
			})
		}
	}

	return detail, nil
}

func deleteAssistantThread(client *http.Client, sessionToken string, thread *threadDetail) error {
	payload := map[string]any{
		"threads": []map[string]any{
			{
				"id":    thread.ID,
				"title": thread.Title,
			},
		},
	}

	_, err := executeKagiStream(client, assistantThreadDeleteURL, sessionToken, payload)
	return err
}
