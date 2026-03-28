// Package data provides embedded data files for the Kagi CLI.
package data

import _ "embed"

// NewsFilterPresets contains the embedded news filter preset definitions.
//
//go:embed news-filter-presets.json
var NewsFilterPresets []byte
