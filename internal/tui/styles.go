// Package tui provides an interactive terminal UI for the Kagi CLI.
package tui

import "github.com/charmbracelet/lipgloss"

// Theme colors — uses adaptive colors that work on both light and dark terminals.
var (
	// ColorPrimary is the main accent color.
	ColorPrimary = lipgloss.AdaptiveColor{Light: "#5A4FCF", Dark: "#8B7FFF"}
	// ColorSecondary is used for URLs and success states.
	ColorSecondary = lipgloss.AdaptiveColor{Light: "#2B8A6F", Dark: "#5FD7A7"}
	// ColorMuted is for dimmed/help text.
	ColorMuted = lipgloss.AdaptiveColor{Light: "#888888", Dark: "#666666"}
	// ColorError is for error messages.
	ColorError = lipgloss.AdaptiveColor{Light: "#CC3333", Dark: "#FF6666"}
	// ColorHighlight is for prominent text.
	ColorHighlight = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFFFF"}
	// ColorBorder is for container borders.
	ColorBorder = lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#444444"}
	// ColorBg is the status bar background.
	ColorBg = lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#1E1E2E"}
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
