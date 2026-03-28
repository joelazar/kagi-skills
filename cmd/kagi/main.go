// Package main is the entry point for the kagi CLI.
package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/joelazar/kagi/internal/commands"
	"github.com/joelazar/kagi/internal/version"
)

func main() {
	rootCmd := commands.NewRootCmd()

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version.Version),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}
