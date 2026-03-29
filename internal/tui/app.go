package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppState represents the current state of the TUI.
type AppState int

const (
	// StateMenu is the command picker.
	StateMenu AppState = iota
	// StateInput is the input form for a selected command.
	StateInput
	// StateLoading indicates command execution in progress.
	StateLoading
	// StateResults shows the result list.
	StateResults
	// StateDetail shows the detail view of a result.
	StateDetail
	// StateError shows an error message.
	StateError
)

// CommandExecutor is called to execute a command with the given inputs.
// It returns result items or an error.
type CommandExecutor func(command string, inputs map[string]string) ([]ResultItem, error)

// commandResultMsg is sent when command execution completes.
type commandResultMsg struct {
	items []ResultItem
	err   error
}

// App is the main Bubble Tea model.
// Bubble Tea requires value receivers for Init/Update/View, but internal methods use pointer receivers.
//
//nolint:recvcheck
type App struct {
	state    AppState
	keys     KeyMap
	menu     list.Model
	input    InputModel
	results  list.Model
	detail   DetailModel
	spinner  spinner.Model
	help     help.Model
	executor CommandExecutor
	compact  bool

	selectedCommand Command
	errorMsg        string
	statusMsg       string
	width           int
	height          int
}

// RunOptions controls how the TUI is presented.
type RunOptions struct {
	// AltScreen enables Bubble Tea's alternate screen buffer.
	AltScreen bool
	// Compact enables a narrower, shorter layout that feels less full-screen.
	Compact bool
}

const (
	compactMaxWidth        = 88
	compactMaxListHeight   = 14
	compactMaxDetailHeight = 18
)

// NewApp creates a new TUI application.
func NewApp(executor CommandExecutor) App {
	return NewAppWithOptions(executor, RunOptions{})
}

// NewAppWithOptions creates a new TUI application with explicit run options.
func NewAppWithOptions(executor CommandExecutor, opts RunOptions) App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	keys := DefaultKeyMap()
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(ColorPrimary)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ColorMuted)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ColorMuted)
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(ColorPrimary)
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(ColorMuted)
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(ColorMuted)
	h.Styles.Ellipsis = lipgloss.NewStyle().Foreground(ColorMuted)

	return App{
		state:    StateMenu,
		keys:     keys,
		spinner:  s,
		help:     h,
		executor: executor,
		compact:  opts.Compact,
		width:    80,
		height:   24,
	}
}

// Init initializes the app.
func (a App) Init() tea.Cmd {
	return nil
}

// Update handles all messages.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.handleResize()
		return a, nil

	case commandResultMsg:
		a.handleCommandResult(msg)
		return a, nil

	case spinner.TickMsg:
		if a.state == StateLoading {
			var cmd tea.Cmd
			a.spinner, cmd = a.spinner.Update(msg)
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		// Global quit.
		if key.Matches(msg, a.keys.Quit) && a.state != StateInput {
			return a, tea.Quit
		}
		return a.handleKey(msg)
	}

	return a.updateCurrent(msg)
}

// View renders the current state.
func (a App) View() string {
	var content string

	switch a.state {
	case StateMenu:
		content = a.menu.View()
	case StateInput:
		content = a.input.View()
	case StateLoading:
		content = a.loadingView()
	case StateResults:
		content = a.results.View()
	case StateDetail:
		content = a.detail.View()
	case StateError:
		content = a.errorView()
	}

	if footer := a.footerView(); footer != "" {
		content += "\n" + footer
	}

	return content
}

// State returns the current app state (for testing).
func (a App) State() AppState {
	return a.state
}

// Run starts the TUI application with the default presentation settings.
func Run(executor CommandExecutor) error {
	return RunWithOptions(executor, RunOptions{})
}

// RunWithOptions starts the TUI application with explicit presentation options.
func RunWithOptions(executor CommandExecutor, opts RunOptions) error {
	app := NewAppWithOptions(executor, opts)
	app.menu = newMenuList(app.layoutWidth(), app.listHeight())

	programOpts := []tea.ProgramOption{}
	if opts.AltScreen {
		programOpts = append(programOpts, tea.WithAltScreen())
	}

	p := tea.NewProgram(app, programOpts...)
	_, err := p.Run()
	return err
}

func (a *App) handleResize() {
	layoutWidth := a.layoutWidth()
	listHeight := a.listHeight()
	detailHeight := a.detailHeight()

	switch a.state {
	case StateMenu:
		a.menu.SetSize(layoutWidth, listHeight)
	case StateInput:
		a.input.SetWidth(layoutWidth)
	case StateResults:
		a.results.SetSize(layoutWidth, listHeight)
	case StateDetail:
		a.detail.SetSize(layoutWidth, detailHeight)
	case StateLoading, StateError:
		// No resize handling needed.
	}
}

func (a App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.state {
	case StateMenu:
		return a.handleMenuKey(msg)
	case StateInput:
		return a.handleInputKey(msg)
	case StateResults:
		return a.handleResultsKey(msg)
	case StateDetail:
		return a.handleDetailKey(msg)
	case StateError:
		return a.handleErrorKey(msg)
	case StateLoading:
		return a, nil
	}
	return a, nil
}

func (a App) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, a.keys.Enter) {
		item, ok := a.menu.SelectedItem().(Command)
		if !ok {
			return a, nil
		}
		a.selectedCommand = item
		a.statusMsg = ""
		if len(item.Fields) == 0 {
			// No input needed, execute directly.
			return a.executeCommand(item.Name, nil)
		}
		a.state = StateInput
		a.input = NewInputModel(item.Name, item.Hint, item.Fields, a.layoutWidth())
		return a, nil
	}

	var cmd tea.Cmd
	a.menu, cmd = a.menu.Update(msg)
	return a, cmd
}

func (a App) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Back):
		a.statusMsg = ""
		a.state = StateMenu
		return a, nil

	case key.Matches(msg, a.keys.Enter):
		if a.input.Validate() {
			return a.executeCommand(a.selectedCommand.Name, a.input.Values())
		}
		return a, nil

	default:
		var cmd tea.Cmd
		_, cmd = a.input.Update(msg)
		return a, cmd
	}
}

func (a App) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Back):
		a.statusMsg = ""
		a.goBackToPrevious()
		return a, nil

	case key.Matches(msg, a.keys.Enter), key.Matches(msg, a.keys.Right):
		item, ok := a.results.SelectedItem().(ResultItem)
		if !ok {
			return a, nil
		}
		a.statusMsg = ""
		a.state = StateDetail
		a.detail = NewDetailModel(item, a.layoutWidth(), a.detailHeight())
		return a, nil

	case key.Matches(msg, a.keys.Open):
		item, ok := a.results.SelectedItem().(ResultItem)
		if ok && item.URL != "" {
			openBrowser(item.URL)
			a.statusMsg = "Opened in browser"
		}
		return a, nil

	case key.Matches(msg, a.keys.Yank):
		item, ok := a.results.SelectedItem().(ResultItem)
		if ok && item.URL != "" {
			if err := clipboard.WriteAll(item.URL); err == nil {
				a.statusMsg = "URL copied!"
			} else {
				a.statusMsg = "Copy failed"
			}
		}
		return a, nil

	default:
		var cmd tea.Cmd
		a.results, cmd = a.results.Update(msg)
		return a, cmd
	}
}

func (a App) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Back), key.Matches(msg, a.keys.Left):
		a.statusMsg = ""
		a.state = StateResults
		return a, nil

	case key.Matches(msg, a.keys.Open):
		if a.detail.url != "" {
			openBrowser(a.detail.url)
			a.statusMsg = "Opened in browser"
		}
		return a, nil

	case key.Matches(msg, a.keys.Yank):
		if a.detail.url != "" {
			if err := clipboard.WriteAll(a.detail.url); err == nil {
				a.statusMsg = "URL copied!"
			} else {
				a.statusMsg = "Copy failed"
			}
		}
		return a, nil

	default:
		var cmd tea.Cmd
		_, cmd = a.detail.Update(msg)
		return a, cmd
	}
}

func (a App) handleErrorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, a.keys.Back) || key.Matches(msg, a.keys.Enter) {
		a.statusMsg = ""
		a.goBackToPrevious()
		return a, nil
	}
	return a, nil
}

func (a *App) handleCommandResult(msg commandResultMsg) {
	a.statusMsg = ""

	if msg.err != nil {
		a.state = StateError
		a.errorMsg = msg.err.Error()
		return
	}

	if len(msg.items) == 0 {
		a.state = StateError
		a.errorMsg = "No results found."
		return
	}

	a.state = StateResults
	title := a.selectedCommand.Name + " results"
	a.results = newResultList(title, msg.items, a.layoutWidth(), a.listHeight())
}

func (a App) executeCommand(command string, inputs map[string]string) (tea.Model, tea.Cmd) {
	a.state = StateLoading
	return a, tea.Batch(
		a.spinner.Tick,
		func() tea.Msg {
			items, err := a.executor(command, inputs)
			return commandResultMsg{items: items, err: err}
		},
	)
}

func (a App) updateCurrent(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.state {
	case StateMenu:
		var cmd tea.Cmd
		a.menu, cmd = a.menu.Update(msg)
		return a, cmd
	case StateResults:
		var cmd tea.Cmd
		a.results, cmd = a.results.Update(msg)
		return a, cmd
	case StateDetail:
		var cmd tea.Cmd
		_, cmd = a.detail.Update(msg)
		return a, cmd
	case StateInput, StateLoading, StateError:
		// No passthrough update needed.
	}
	return a, nil
}

func (a App) loadingView() string {
	return fmt.Sprintf(
		"\n  %s Running %s...\n\n  %s",
		a.spinner.View(),
		SelectedItemStyle.Render(a.selectedCommand.Name),
		DimStyle.Render("Waiting for the API response"),
	)
}

func (a App) errorView() string {
	return "\n" + ErrorStyle.Render("  Error") + "\n\n" +
		"  " + a.errorMsg
}

func (a App) layoutWidth() int {
	width := a.width
	if width <= 0 {
		width = 80
	}
	if a.compact {
		width = min(width, compactMaxWidth)
	}
	return width
}

func (a App) listHeight() int {
	height := a.height
	if height <= 0 {
		height = 24
	}
	listHeight := max(height-2, 5)
	if a.compact {
		listHeight = min(listHeight, compactMaxListHeight)
	}
	return listHeight
}

func (a App) detailHeight() int {
	height := a.height
	if height <= 0 {
		height = 24
	}
	if a.compact {
		height = min(height, compactMaxDetailHeight)
	}
	return height
}

func (a App) footerView() string {
	width := a.layoutWidth()

	status := a.footerStatus()
	bindings := a.footerHelpBindings()
	if status == "" && len(bindings.short) == 0 && len(bindings.full) == 0 {
		return ""
	}

	innerWidth := max(width-2, 20)
	parts := make([]string, 0, 2)
	availableHelpWidth := innerWidth

	if status != "" {
		maxStatusWidth := max((innerWidth*2)/3, 18)
		statusRendered := FooterStatusStyle.MaxWidth(maxStatusWidth).Render(status)
		parts = append(parts, statusRendered)
		availableHelpWidth = max(innerWidth-lipgloss.Width(statusRendered)-2, 0)
	}

	helpView := ""
	if len(bindings.short) > 0 || len(bindings.full) > 0 {
		h := a.help
		h.Width = availableHelpWidth
		helpView = h.View(bindings)
		if helpView != "" {
			if len(parts) > 0 {
				parts = append(parts, "  ")
			}
			parts = append(parts, helpView)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return StatusBarStyle.Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Top, parts...))
}

func (a App) footerStatus() string {
	left := ""

	switch a.state {
	case StateMenu:
		left = a.listStatus(a.menu, "command")
	case StateInput:
		left = fmt.Sprintf("%s input", a.selectedCommand.Name)
	case StateLoading:
		left = fmt.Sprintf("Running %s", a.selectedCommand.Name)
	case StateResults:
		left = a.listStatus(a.results, "result")
	case StateDetail:
		left = strings.TrimSpace(a.detail.title + "  •  " + a.detail.scrollInfo())
	case StateError:
		left = "Request failed"
	}

	if a.statusMsg != "" {
		left = strings.TrimSpace(left + "  •  " + a.statusMsg)
	}

	return left
}

type helpBindings struct {
	short []key.Binding
	full  [][]key.Binding
}

func newHelpBindings(short []key.Binding, full [][]key.Binding) helpBindings {
	return helpBindings{short: short, full: full}
}

func (h helpBindings) ShortHelp() []key.Binding { return h.short }

func (h helpBindings) FullHelp() [][]key.Binding { return h.full }

func (a App) footerHelpBindings() helpBindings {
	switch a.state {
	case StateMenu:
		return newHelpBindings(
			[]key.Binding{a.keys.Enter, a.keys.Search, a.keys.Up, a.keys.Down, a.keys.Quit},
			[][]key.Binding{{a.keys.Enter, a.keys.Search, a.keys.Up, a.keys.Down}, {a.keys.Quit}},
		)
	case StateInput:
		return newHelpBindings(
			[]key.Binding{a.keys.Back, a.keys.Tab, a.keys.Enter},
			[][]key.Binding{{a.keys.Back, a.keys.Tab, a.keys.Enter}},
		)
	case StateLoading:
		return newHelpBindings(
			[]key.Binding{a.keys.Quit},
			[][]key.Binding{{a.keys.Quit}},
		)
	case StateResults:
		return newHelpBindings(
			[]key.Binding{a.keys.Enter, a.keys.Search, a.keys.Open, a.keys.Yank, a.keys.Back},
			[][]key.Binding{{a.keys.Enter, a.keys.Search, a.keys.Back}, {a.keys.Open, a.keys.Yank, a.keys.Up, a.keys.Down}},
		)
	case StateDetail:
		return newHelpBindings(
			[]key.Binding{a.keys.Back, a.keys.Open, a.keys.Yank, a.keys.Up, a.keys.Down},
			[][]key.Binding{{a.keys.Back, a.keys.Open, a.keys.Yank}, {a.keys.Up, a.keys.Down, a.keys.PageUp, a.keys.PageDown}},
		)
	case StateError:
		return newHelpBindings(
			[]key.Binding{a.keys.Enter, a.keys.Back},
			[][]key.Binding{{a.keys.Enter, a.keys.Back}},
		)
	default:
		return helpBindings{}
	}
}

func (a App) listStatus(m list.Model, noun string) string {
	shown := len(m.VisibleItems())
	total := len(m.Items())
	if total == 0 {
		return fmt.Sprintf("No %ss", noun)
	}
	if m.IsFiltered() || m.SettingFilter() {
		filter := m.FilterValue()
		if filter == "" {
			return fmt.Sprintf("Showing %d of %d %ss", shown, total, noun)
		}
		return fmt.Sprintf("Showing %d of %d %ss  •  filter: %q", shown, total, noun, filter)
	}
	if shown == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", shown, noun)
}

func (a App) backTargetLabel() string {
	if len(a.selectedCommand.Fields) > 0 {
		return "edit query"
	}
	return "commands"
}

func (a *App) goBackToPrevious() {
	if len(a.selectedCommand.Fields) > 0 {
		a.state = StateInput
		return
	}
	a.state = StateMenu
}

// openBrowser opens a URL in the default browser.
func openBrowser(url string) {
	if url == "" {
		return
	}
	// Basic validation.
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return
	}

	ctx := context.Background()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}
