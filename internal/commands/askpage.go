package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

type askPageOutput struct {
	URL        string `json:"url"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	ThreadID   string `json:"thread_id,omitempty"`
	References string `json:"references,omitempty"`
}

func newAskPageCmd() *cobra.Command {
	var timeoutSec int

	cmd := &cobra.Command{
		Use:   "askpage <url> <question>",
		Short: "Ask a question about a web page",
		Long: `Ask a question about a specific web page using Kagi Assistant.
Requires KAGI_SESSION_TOKEN (your Kagi session cookie).

The question is answered using the content of the provided URL as context.`,
		Example: `  kagi askpage https://go.dev/doc/effective_go "What are the naming conventions?"
  kagi askpage https://example.com/article "Summarize the key points" --format json`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			pageURL := strings.TrimSpace(args[0])
			question := strings.TrimSpace(strings.Join(args[1:], " "))

			if pageURL == "" {
				return errors.New("URL is required")
			}
			if question == "" {
				return errors.New("question is required")
			}

			sessionToken, err := api.ResolveSessionToken(cfg)
			if err != nil {
				return err
			}

			// Build the prompt: URL + newline + question (same as Rust CLI)
			prompt := pageURL + "\n" + question

			client := api.NewHTTPClient(time.Duration(timeoutSec) * time.Second)
			resp, err := callAssistantPrompt(client, sessionToken, prompt, "")
			if err != nil {
				return err
			}

			out := askPageOutput{
				URL:        pageURL,
				Question:   question,
				Answer:     resp.Output,
				ThreadID:   resp.ThreadID,
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

	text := htmlToMarkdown(out.Answer)
	fmt.Println(text)

	if out.References != "" {
		refs := htmlToMarkdown(out.References)
		if refs != "" {
			fmt.Println()
			fmt.Println("--- References ---")
			fmt.Println(refs)
		}
	}

	if out.ThreadID != "" {
		fmt.Fprintf(output.Stderr(), "[thread: %s]\n", out.ThreadID)
	}

	return nil
}
