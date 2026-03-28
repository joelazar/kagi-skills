// Package main is the entry point for the kagi CLI.
package main

import (
	"fmt"
	"os"

	"github.com/joelazar/kagi/internal/commands"
)

func main() {
	if err := commands.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
