package commands

import (
	"fmt"

	"github.com/joelazar/kagi/internal/version"
	"github.com/spf13/cobra"
)

// @lat: [[cli#Local operator workflows]]
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build information",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(version.Info())
			return nil
		},
	}
}
