// Package main is the entry point for the kagi CLI.
package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/joelazar/kagi/internal/commands"
	kagistyle "github.com/joelazar/kagi/internal/style"
	"github.com/joelazar/kagi/internal/version"
)

// @lat: [[architecture#Process entry and command assembly]]
func main() {
	rootCmd := commands.NewRootCmd()

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version.Version),
		fang.WithColorSchemeFunc(kagistyle.FangColorScheme),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
