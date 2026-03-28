package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
	executor CommandExecutor

	selectedCommand Command
	errorMsg        string
	statusMsg       string
	width           int
	height          int
}

// NewApp creates a new TUI application.
func NewApp(executor CommandExecutor) App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	keys := DefaultKeyMap()

	return App{
		state:    StateMenu,
		keys:     keys,
		spinner:  s,
		executor: executor,
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

	// Append status message if any.
	if a.statusMsg != "" {
		content += "\n" + HelpStyle.Render(a.statusMsg)
	}

	return content
}

// State returns the current app state (for testing).
func (a App) State() AppState {
	return a.state
}

// Run starts the TUI application.
func Run(executor CommandExecutor) error {
	app := NewApp(executor)
	app.menu = newMenuList(app.width, app.height-2)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) handleResize() {
	menuHeight := max(a.height-2, 5)

	switch a.state {
	case StateMenu:
		a.menu.SetSize(a.width, menuHeight)
	case StateInput:
		a.input.SetWidth(a.width)
	case StateResults:
		a.results.SetSize(a.width, menuHeight)
	case StateDetail:
		a.detail.SetSize(a.width, a.height)
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
		if len(item.Fields) == 0 {
			// No input needed, execute directly.
			return a.executeCommand(item.Name, nil)
		}
		a.state = StateInput
		a.input = NewInputModel(item.Name, item.Fields, a.width)
		return a, nil
	}

	var cmd tea.Cmd
	a.menu, cmd = a.menu.Update(msg)
	return a, cmd
}

func (a App) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Back):
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
		a.state = StateMenu
		return a, nil

	case key.Matches(msg, a.keys.Enter), key.Matches(msg, a.keys.Right):
		item, ok := a.results.SelectedItem().(ResultItem)
		if !ok {
			return a, nil
		}
		a.state = StateDetail
		a.detail = NewDetailModel(item, a.width, a.height)
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
		a.state = StateMenu
		return a, nil
	}
	return a, nil
}

func (a *App) handleCommandResult(msg commandResultMsg) {
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
	a.results = newResultList(title, msg.items, a.width, a.height-2)
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
		DimStyle.Render("Press Ctrl+C to cancel"),
	)
}

func (a App) errorView() string {
	return "\n" + ErrorStyle.Render("  Error") + "\n\n" +
		"  " + a.errorMsg + "\n\n" +
		DimStyle.Render("  Press Enter or Esc to go back")
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
