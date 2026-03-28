package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/output"
	"github.com/spf13/cobra"
)

func newBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Show cached API balance",
		Long:  "Display the most recently cached API balance from any Kagi API call.",
		RunE: func(_ *cobra.Command, _ []string) error {
			cached, err := api.LoadBalanceCache()
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return errors.New("no cached API balance yet; run a Kagi API command first")
				}
				return err
			}

			format := getFormat()
			if format == output.FormatJSON {
				return output.WriteJSON(cached)
			}
			if format == output.FormatCompact {
				return output.WriteCompact(cached)
			}

			fmt.Printf("API Balance: $%.4f\n", cached.APIBalance)
			fmt.Printf("Updated: %s\n", cached.UpdatedAt)
			if cached.Source != "" {
				fmt.Printf("Source: %s\n", cached.Source)
			}
			return nil
		},
	}

	return cmd
}
