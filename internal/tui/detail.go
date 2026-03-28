package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// DetailModel displays a detailed view of a result with markdown rendering.
type DetailModel struct {
	viewport viewport.Model
	title    string
	url      string
	ready    bool
	width    int
	height   int
}

// NewDetailModel creates a detail view for a result item.
func NewDetailModel(item ResultItem, width, height int) DetailModel {
	vp := viewport.New(width, height-4) // Reserve space for header/footer.

	content := renderMarkdown(item.Detail, width-4)
	vp.SetContent(content)

	return DetailModel{
		viewport: vp,
		title:    item.Title,
		url:      item.URL,
		ready:    true,
		width:    width,
		height:   height,
	}
}

// Update handles viewport messages.
func (m *DetailModel) Update(msg tea.Msg) (*DetailModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the detail view.
func (m *DetailModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	header := TitleStyle.Render("  " + m.title)
	if m.url != "" {
		header += "\n" + URLStyle.Render(m.url)
	}

	footer := m.footerView()

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

// SetSize updates the viewport dimensions.
func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height - 4
}

func (m *DetailModel) footerView() string {
	percent := fmt.Sprintf(" %3.f%% ", m.viewport.ScrollPercent()*100)
	info := lipgloss.NewStyle().Foreground(ColorMuted).Render(percent)
	help := DimStyle.Render("esc: back  •  o: open  •  y: copy URL  •  ↑↓/jk: scroll")

	return StatusBarStyle.Width(m.width).Render(help + "  " + info)
}

// renderMarkdown renders markdown content using glamour.
func renderMarkdown(content string, width int) string {
	if content == "" {
		return DimStyle.Render("No content available.")
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}

	rendered, err := r.Render(content)
	if err != nil {
		return content
	}

	return rendered
}
