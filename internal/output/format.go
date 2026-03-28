// Package output provides output formatting for the Kagi CLI.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Format represents an output format type.
type Format string

const (
	FormatJSON     Format = "json"
	FormatCompact  Format = "compact"
	FormatPretty   Format = "pretty"
	FormatMarkdown Format = "markdown"
	FormatCSV      Format = "csv"
)

// ValidFormats returns a list of all valid format strings.
func ValidFormats() []string {
	return []string{
		string(FormatJSON),
		string(FormatCompact),
		string(FormatPretty),
		string(FormatMarkdown),
		string(FormatCSV),
	}
}

// ParseFormat validates and returns a Format from a string.
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case FormatJSON, FormatCompact, FormatPretty, FormatMarkdown, FormatCSV:
		return Format(s), nil
	default:
		return "", fmt.Errorf("unknown format %q — valid: json, compact, pretty, markdown, csv", s)
	}
}

// WriteJSON writes v as indented JSON to stdout.
func WriteJSON(v any) error {
	return WriteJSONTo(os.Stdout, v)
}

// WriteJSONTo writes v as indented JSON to w.
func WriteJSONTo(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteCompact writes v as compact JSON to stdout.
func WriteCompact(v any) error {
	return WriteCompactTo(os.Stdout, v)
}

// WriteCompactTo writes v as compact JSON to w.
func WriteCompactTo(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}
