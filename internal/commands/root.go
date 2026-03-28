// Package commands provides all CLI commands for the Kagi CLI.
package commands

import (
	"fmt"
	"os"

	"github.com/joelazar/kagi/internal/config"
	"github.com/joelazar/kagi/internal/output"
	"github.com/joelazar/kagi/internal/version"
	"github.com/spf13/cobra"
)

var (
	formatFlag string
	cfg        *config.Config
)

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "kagi",
		Short:   "Kagi CLI — search, summarize, and more from the terminal",
		Long:    "A unified CLI for all Kagi APIs: search, FastGPT, summarizer, enrichment, and more.",
		Version: version.Version,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			var err error
			cfg, err = config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "json",
		fmt.Sprintf("output format (%s)", "json, compact, pretty, markdown, csv"))

	rootCmd.AddCommand(
		newSearchCmd(),
		newFastGPTCmd(),
		newSummarizeCmd(),
		newEnrichCmd(),
		newBalanceCmd(),
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
