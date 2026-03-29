package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command represents a selectable command in the menu.
type Command struct {
	Name        string
	Description string
	// Fields describes the input fields required for this command.
	Fields []InputField
}

// FilterValue implements list.Item.
func (c Command) FilterValue() string { return c.Name }

// commandDelegate renders menu items.
type commandDelegate struct{}

func (d commandDelegate) Height() int                             { return 2 }
func (d commandDelegate) Spacing() int                            { return 0 }
func (d commandDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d commandDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	cmd, ok := listItem.(Command)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	var name, desc string
	if isSelected {
		name = SelectedItemStyle.Render("▸ " + cmd.Name)
		desc = lipgloss.NewStyle().Foreground(ColorSecondary).Render("  " + cmd.Description)
	} else {
		name = NormalItemStyle.Render("  " + cmd.Name)
		desc = DimStyle.Render("  " + cmd.Description)
	}

	fmt.Fprintf(w, "%s\n%s", name, desc)
}

// MenuCommands returns all available commands for the TUI menu.
func MenuCommands() []Command {
	return []Command{
		{
			Name:        "search",
			Description: "Search the web via Kagi Search API",
			Fields: []InputField{
				{Key: "query", Label: "Search query", Placeholder: "e.g. golang generics", Required: true},
				{Key: "limit", Label: "Number of results", Placeholder: "10", Required: false},
			},
		},
		{
			Name:        "fastgpt",
			Description: "Get AI-synthesized answers via FastGPT",
			Fields: []InputField{
				{Key: "query", Label: "Question", Placeholder: "e.g. What is the capital of France?", Required: true},
			},
		},
		{
			Name:        "summarize",
			Description: "Summarize a URL or text",
			Fields: []InputField{
				{Key: "url", Label: "URL to summarize", Placeholder: "https://example.com/article", Required: true},
				{Key: "engine", Label: "Engine (cecil/agnes/daphne/muriel)", Placeholder: "cecil", Required: false},
			},
		},
		{
			Name:        "enrich",
			Description: "Search Kagi's non-commercial web & news indexes",
			Fields: []InputField{
				{Key: "query", Label: "Search query", Placeholder: "e.g. independent blogs about Go", Required: true},
			},
		},
		{
			Name:        "quick",
			Description: "Quick Answer via subscriber session",
			Fields: []InputField{
				{Key: "query", Label: "Question", Placeholder: "e.g. weather in Berlin", Required: true},
			},
		},
		{
			Name:        "translate",
			Description: "Translate text between languages",
			Fields: []InputField{
				{Key: "text", Label: "Text to translate", Placeholder: "Hello, world!", Required: true},
				{Key: "target", Label: "Target language", Placeholder: "e.g. German, French", Required: true},
			},
		},
		{
			Name:        "news",
			Description: "Browse news by category",
			Fields: []InputField{
				{Key: "category", Label: "Category (world/business/tech/science/sports/...)", Placeholder: "world", Required: false},
			},
		},
		{
			Name:        "smallweb",
			Description: "Browse the Small Web feed",
			Fields:      nil,
		},
		{
			Name:        "askpage",
			Description: "Ask questions about a URL",
			Fields: []InputField{
				{Key: "url", Label: "URL to ask about", Placeholder: "https://example.com/article", Required: true},
				{Key: "query", Label: "Question about the page", Placeholder: "What is the main argument?", Required: true},
			},
		},
		{
			Name:        "balance",
			Description: "Check your API balance",
			Fields:      nil,
		},
	}
}

// newMenuList creates a list.Model configured as the command menu.
func newMenuList(width, height int) list.Model {
	commands := MenuCommands()
	items := make([]list.Item, len(commands))
	for i, cmd := range commands {
		items[i] = cmd
	}

	l := list.New(items, commandDelegate{}, width, height)
	l.Title = "Kagi CLI"
	l.Styles.Title = TitleStyle
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return l
}
