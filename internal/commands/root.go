// Package commands provides all CLI commands for the Kagi CLI.
package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/joelazar/kagi/internal/config"
	"github.com/joelazar/kagi/internal/output"
	"github.com/joelazar/kagi/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	formatFlag   string
	interactiveF bool
	noTUIFlag    bool
	cfg          *config.Config
)

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kagi",
		Short: "Kagi CLI — search, summarize, and more from the terminal",
		Long:  "A unified CLI for all Kagi APIs: search, FastGPT, summarizer, enrichment, and more.",
		Example: `  kagi search "golang generics"
  kagi fastgpt "What is the capital of France?"
  kagi summarize https://example.com/article
  kagi enrich web "independent blogs"
  kagi translate "Hello world" --target German
  kagi news --category tech
  kagi balance`,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			var err error
			cfg, err = config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
			}
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			// Launch TUI when run with no args (or --interactive), unless --no-tui or piped.
			if noTUIFlag || !isTerminal() {
				return errors.New("no command specified. Run 'kagi --help' for usage")
			}
			executor := tui.NewExecutor(cfg)
			return tui.Run(executor)
		},
	}

	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "json",
		fmt.Sprintf("output format (%s)", "json, compact, pretty, markdown, csv"))
	rootCmd.Flags().BoolVarP(&interactiveF, "interactive", "i", false, "launch interactive TUI mode")
	rootCmd.Flags().BoolVar(&noTUIFlag, "no-tui", false, "force non-interactive mode (useful in pipes)")

	rootCmd.AddCommand(
		newSearchCmd(),
		newFastGPTCmd(),
		newSummarizeCmd(),
		newEnrichCmd(),
		newBalanceCmd(),
		newSmallWebCmd(),
		newBatchCmd(),
		newQuickCmd(),
		newAssistantCmd(),
		newAskPageCmd(),
		newTranslateCmd(),
		newNewsCmd(),
		newVersionCmd(),
	)

	return rootCmd
}

// getFormat returns the parsed output format from the --format flag.
func getFormat() output.Format {
	f, err := output.ParseFormat(formatFlag)
	if err != nil {
		return output.FormatJSON
	}
	return f
}

// isTerminal returns true if stdout is a terminal (not piped).
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) //nolint:gosec
}
