// Package version provides build-time version information.
//
// Version, Commit, and Date are set at build time via -ldflags:
//
//	-X github.com/joelazar/kagi/internal/version.Version=...
//	-X github.com/joelazar/kagi/internal/version.Commit=...
//	-X github.com/joelazar/kagi/internal/version.Date=...
package version

import "runtime"

var (
	// Version is the semantic version (e.g. "1.2.0").
	Version = "dev"
	// Commit is the git commit hash.
	Commit = "unknown"
	// Date is the build date (ISO 8601).
	Date = "unknown"
)

// Info returns a formatted multi-line version string.
func Info() string {
	return "kagi " + Version + "\n" +
		"commit: " + Commit + "\n" +
		"built:  " + Date + "\n" +
		"go:     " + runtime.Version()
}
