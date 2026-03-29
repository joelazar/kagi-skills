// Package tui provides an interactive terminal UI for the Kagi CLI.
package tui

import (
	"github.com/charmbracelet/lipgloss"
	kagistyle "github.com/joelazar/kagi/internal/style"
)

// Theme colors — uses a Kagi-inspired adaptive palette that works on both light and dark terminals.
var (
	// ColorPrimary is the main accent color.
	ColorPrimary = lipgloss.AdaptiveColor{Light: kagistyle.PurpleLight, Dark: kagistyle.PurpleDark}
	// ColorSecondary is used for URLs and success states.
	ColorSecondary = lipgloss.AdaptiveColor{Light: kagistyle.TealLight, Dark: kagistyle.TealDark}
	// ColorMuted is for dimmed/help text.
	ColorMuted = lipgloss.AdaptiveColor{Light: kagistyle.MutedLight, Dark: kagistyle.MutedDark}
	// ColorError is for error messages.
	ColorError = lipgloss.AdaptiveColor{Light: kagistyle.ErrorLight, Dark: kagistyle.ErrorDark}
	// ColorHighlight is for prominent text.
	ColorHighlight = lipgloss.AdaptiveColor{Light: kagistyle.InkLight, Dark: kagistyle.InkDark}
	// ColorBorder is for container borders.
	ColorBorder = lipgloss.AdaptiveColor{Light: kagistyle.BorderLight, Dark: kagistyle.BorderDark}
	// ColorBg is the status bar background.
	ColorBg = lipgloss.AdaptiveColor{Light: kagistyle.PanelLight, Dark: kagistyle.PanelDark}
)

// Shared styles used across TUI components.
var (
	// TitleStyle is the title bar at the top of the app.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight).
			Background(ColorPrimary).
			Padding(0, 1)

	// SubtitleStyle is for breadcrumbs.
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	// SelectedItemStyle is for active/selected list items.
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	// NormalItemStyle is for normal list items.
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight)

	// DimStyle is for dimmed/description text.
	DimStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// URLStyle is for URL rendering.
	URLStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Underline(true)

	// ErrorStyle is for error messages.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// StatusBarStyle is for the status bar at the bottom.
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorBg).
			Padding(0, 1)

	// HelpStyle is for help text in the status bar.
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// FooterStatusStyle is for the status text rendered beside the adaptive help view.
	FooterStatusStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// BorderedStyle is for bordered containers in detail views.
	BorderedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// SpinnerStyle is for spinner/loading text.
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)
)
