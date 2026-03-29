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
	Hint        string
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
			Description: "Search the web with optional per-result content fetching",
			Hint:        "Interactive parity: query, result count, and optional content extraction are exposed. Balance, format, and timeout flags stay CLI-only.",
			Fields: []InputField{
				{Key: "query", Label: "Search query", Placeholder: "e.g. golang generics", Required: true},
				{Key: "limit", Label: "Number of results", Placeholder: "10"},
				{Key: "include_content", Label: "Fetch page content? (yes/no)", Placeholder: "no"},
			},
		},
		{
			Name:        "search content",
			Description: "Extract readable page content from a URL",
			Hint:        "TUI-native version of `kagi search content <url>`.",
			Fields:      []InputField{{Key: "url", Label: "URL", Placeholder: "https://example.com/article", Required: true}},
		},
		{
			Name:        "fastgpt",
			Description: "Get an AI-synthesized answer with references",
			Hint:        "Interactive parity intentionally keeps the core question flow. Cache, balance, format, and timeout flags remain CLI-oriented.",
			Fields:      []InputField{{Key: "query", Label: "Question", Placeholder: "e.g. What is the capital of France?", Required: true}},
		},
		{
			Name:        "summarize",
			Description: "Summarize a URL or pasted text",
			Hint:        "Interactive parity exposes the input mode, engine, summary type, and output language. Raw format/cache/timeout flags stay CLI-only.",
			Fields: []InputField{
				{Key: "input", Label: "URL or text", Placeholder: "https://example.com/article or pasted text", Required: true, MaxLength: 8000},
				{Key: "input_mode", Label: "Input mode (url/text)", Placeholder: "url"},
				{Key: "engine", Label: "Engine (cecil/agnes/daphne/muriel)", Placeholder: "cecil"},
				{Key: "summary_type", Label: "Summary type (summary/takeaway)", Placeholder: "summary"},
				{Key: "lang", Label: "Output language code", Placeholder: "e.g. EN, DE, FR"},
			},
		},
		{
			Name:        "enrich",
			Description: "Search Kagi's independent-web and news indexes",
			Hint:        "Interactive parity exposes index choice (`web` vs `news`) and result count.",
			Fields: []InputField{
				{Key: "index", Label: "Index (web/news)", Placeholder: "web"},
				{Key: "query", Label: "Search query", Placeholder: "e.g. independent blogs about Go", Required: true},
				{Key: "limit", Label: "Maximum results", Placeholder: "10"},
			},
		},
		{
			Name:        "quick",
			Description: "Use your subscriber session for a quick AI answer",
			Hint:        "Fast, single-question subscriber workflow. Format and timeout flags remain CLI-only.",
			Fields:      []InputField{{Key: "query", Label: "Question", Placeholder: "e.g. weather in Berlin", Required: true}},
		},
		{
			Name:        "translate",
			Description: "Translate text between languages",
			Hint:        "Interactive parity exposes target language, optional source language, and formality.",
			Fields: []InputField{
				{Key: "text", Label: "Text to translate", Placeholder: "Hello, world!", Required: true, MaxLength: 8000},
				{Key: "target_lang", Label: "Target language code", Placeholder: "e.g. DE, JA, EN", Required: true},
				{Key: "source_lang", Label: "Source language code", Placeholder: "auto"},
				{Key: "formality", Label: "Formality (formal/informal)", Placeholder: "optional"},
			},
		},
		{
			Name:        "news",
			Description: "Browse Kagi's curated news feed",
			Hint:        "Interactive parity exposes category, language, and result count.",
			Fields: []InputField{
				{Key: "category", Label: "Category", Placeholder: "e.g. technology, science, world"},
				{Key: "limit", Label: "Maximum items", Placeholder: "20"},
				{Key: "lang", Label: "Language code", Placeholder: "en"},
			},
		},
		{
			Name:        "smallweb",
			Description: "Browse recent posts from Kagi's Small Web feed",
			Hint:        "Interactive parity exposes the fetch limit; output format and timeout remain CLI-only.",
			Fields:      []InputField{{Key: "limit", Label: "Maximum entries", Placeholder: "20"}},
		},
		{
			Name:        "askpage",
			Description: "Ask a question about a specific URL",
			Hint:        "Interactive parity focuses on the page URL and question prompt.",
			Fields: []InputField{
				{Key: "url", Label: "URL to ask about", Placeholder: "https://example.com/article", Required: true},
				{Key: "query", Label: "Question about the page", Placeholder: "What is the main argument?", Required: true, MaxLength: 4000},
			},
		},
		{
			Name:        "assistant",
			Description: "Chat with Kagi Assistant",
			Hint:        "Interactive parity exposes the main prompt plus an optional thread ID for continuing a conversation.",
			Fields: []InputField{
				{Key: "query", Label: "Prompt", Placeholder: "Explain quantum computing", Required: true, MaxLength: 4000},
				{Key: "thread_id", Label: "Existing thread ID (optional)", Placeholder: "continue a previous thread"},
			},
		},
		{
			Name:        "assistant threads",
			Description: "Browse assistant conversation threads",
			Hint:        "TUI-native thread workflow: list recent threads, then open a thread's full conversation in the detail view.",
			Fields:      []InputField{{Key: "limit", Label: "Maximum threads to load", Placeholder: "10"}},
		},
		{
			Name:        "assistant delete thread",
			Description: "Delete an assistant thread by ID",
			Hint:        "Destructive action kept explicit in the TUI. Use `assistant threads` first to inspect thread IDs.",
			Fields:      []InputField{{Key: "thread_id", Label: "Thread ID", Placeholder: "paste a thread ID", Required: true}},
		},
		{
			Name:        "balance",
			Description: "Show your cached API balance",
			Hint:        "The TUI uses the same cached balance file as `kagi balance`.",
			Fields:      nil,
		},
		{
			Name:        "auth",
			Description: "Check API-key health and subscriber-token presence",
			Hint:        "Interactive story for `kagi auth`: validate the API key, then report whether a session token is configured.",
			Fields:      nil,
		},
		{
			Name:        "config",
			Description: "Inspect the config path and update saved defaults",
			Hint:        "Leave fields blank to keep existing values. Clearing values remains a file-editing task outside the TUI.",
			Fields: []InputField{
				{Key: "api_key", Label: "API key (optional update)", Placeholder: "leave blank to preserve current value", Sensitive: true},
				{Key: "session_token", Label: "Session token (optional update)", Placeholder: "leave blank to preserve current value", Sensitive: true},
				{Key: "default_format", Label: "Default format", Placeholder: "json / compact / pretty / markdown / csv"},
				{Key: "search_region", Label: "Default search region", Placeholder: "e.g. us-en, de-de"},
			},
		},
		{
			Name:        "completion",
			Description: "Show shell-completion setup commands",
			Hint:        "The TUI explains how to generate completions, while the actual completion scripts remain a CLI export flow.",
			Fields:      nil,
		},
		{
			Name:        "version",
			Description: "Show version and build metadata",
			Hint:        "Interactive story for `kagi version`.",
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
