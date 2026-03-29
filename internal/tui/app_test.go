package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// mockExecutor returns preset results.
func mockExecutor(items []ResultItem, err error) CommandExecutor {
	return func(_ string, _ map[string]string) ([]ResultItem, error) {
		return items, err
	}
}

func newTestApp(executor CommandExecutor) App {
	return newTestAppWithOptions(executor, RunOptions{})
}

func newTestAppWithOptions(executor CommandExecutor, opts RunOptions) App {
	app := NewAppWithOptions(executor, opts)
	app.menu = newMenuList(app.layoutWidth(), app.listHeight())
	return app
}

func sendSpecialKey(app tea.Model, keyType tea.KeyType) (tea.Model, tea.Cmd) {
	return app.Update(tea.KeyMsg{Type: keyType})
}

func mustApp(t *testing.T, m tea.Model) App {
	t.Helper()
	a, ok := m.(App)
	if !ok {
		t.Fatal("expected App type")
	}
	return a
}

func TestInitialState(t *testing.T) {
	app := newTestApp(nil)
	if app.State() != StateMenu {
		t.Errorf("expected StateMenu, got %v", app.State())
	}
}

func TestMenuToInput(t *testing.T) {
	app := newTestApp(nil)

	// Press enter to select the first command (search).
	m, _ := sendSpecialKey(app, tea.KeyEnter)
	a := mustApp(t, m)

	if a.State() != StateInput {
		t.Errorf("expected StateInput after enter, got %v", a.State())
	}
}

func TestInputBack(t *testing.T) {
	app := newTestApp(nil)

	// Go to input state.
	m, _ := sendSpecialKey(app, tea.KeyEnter)
	a := mustApp(t, m)
	if a.State() != StateInput {
		t.Fatalf("expected StateInput, got %v", a.State())
	}

	// Press esc to go back.
	m, _ = sendSpecialKey(a, tea.KeyEscape)
	a = mustApp(t, m)

	if a.State() != StateMenu {
		t.Errorf("expected StateMenu after esc, got %v", a.State())
	}
}

func TestQuestionMarkInInputDoesNotToggleHelp(t *testing.T) {
	app := newTestApp(nil)
	m, _ := sendSpecialKey(app, tea.KeyEnter)
	a := mustApp(t, m)
	if a.State() != StateInput {
		t.Fatalf("expected StateInput, got %v", a.State())
	}

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	a = mustApp(t, m)

	if a.help.ShowAll {
		t.Fatal("expected ? in input mode to be treated as text, not help toggle")
	}
	if got := a.input.inputs[0].Value(); got != "?" {
		t.Fatalf("expected first input to contain ?, got %q", got)
	}
}

func TestCommandResultSuccess(t *testing.T) {
	items := []ResultItem{
		{Title: "Test Result", URL: "https://example.com", Description: "A test"},
	}
	app := newTestApp(mockExecutor(items, nil))

	// Simulate receiving command results.
	app.handleCommandResult(commandResultMsg{items: items})

	if app.State() != StateResults {
		t.Errorf("expected StateResults, got %v", app.State())
	}
}

func TestCommandResultError(t *testing.T) {
	app := newTestApp(nil)

	app.handleCommandResult(commandResultMsg{err: errors.New("test error")})

	if app.State() != StateError {
		t.Errorf("expected StateError, got %v", app.State())
	}
}

func TestCommandResultEmpty(t *testing.T) {
	app := newTestApp(nil)

	app.handleCommandResult(commandResultMsg{items: nil})

	if app.State() != StateError {
		t.Errorf("expected StateError for empty results, got %v", app.State())
	}
}

func TestErrorBackToMenu(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateError
	app.errorMsg = "test error"

	m, _ := sendSpecialKey(app, tea.KeyEnter)
	a := mustApp(t, m)

	if a.State() != StateMenu {
		t.Errorf("expected StateMenu after enter in error state, got %v", a.State())
	}
}

func TestErrorBackToInputWhenCommandHasFields(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateError
	app.selectedCommand = Command{Name: "search", Fields: []InputField{{Key: "query", Label: "Query", Required: true}}}
	app.input = NewInputModel("search", "", app.selectedCommand.Fields, 80)
	app.errorMsg = "test error"

	m, _ := sendSpecialKey(app, tea.KeyEnter)
	a := mustApp(t, m)

	if a.State() != StateInput {
		t.Errorf("expected StateInput after enter in error state, got %v", a.State())
	}
}

func TestResultsToDetail(t *testing.T) {
	items := []ResultItem{
		{Title: "Test", URL: "https://example.com", Description: "desc", Detail: "# Detail"},
	}
	app := newTestApp(nil)
	app.state = StateResults
	app.results = newResultList("test", items, 80, 20)

	m, _ := sendSpecialKey(app, tea.KeyEnter)
	a := mustApp(t, m)

	if a.State() != StateDetail {
		t.Errorf("expected StateDetail, got %v", a.State())
	}
}

func TestDetailBackToResults(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateDetail
	app.detail = NewDetailModel(ResultItem{Title: "Test"}, 80, 24)

	m, _ := sendSpecialKey(app, tea.KeyEscape)
	a := mustApp(t, m)

	if a.State() != StateResults {
		t.Errorf("expected StateResults after esc in detail, got %v", a.State())
	}
}

func TestResultsBackToMenu(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateResults
	app.results = newResultList("test", nil, 80, 20)

	m, _ := sendSpecialKey(app, tea.KeyEscape)
	a := mustApp(t, m)

	if a.State() != StateMenu {
		t.Errorf("expected StateMenu after esc in results, got %v", a.State())
	}
}

func TestResultsBackToInputWhenCommandHasFields(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateResults
	app.selectedCommand = Command{Name: "search", Fields: []InputField{{Key: "query", Label: "Query", Required: true}}}
	app.input = NewInputModel("search", "", app.selectedCommand.Fields, 80)
	app.results = newResultList("test", nil, 80, 20)

	m, _ := sendSpecialKey(app, tea.KeyEscape)
	a := mustApp(t, m)

	if a.State() != StateInput {
		t.Errorf("expected StateInput after esc in results, got %v", a.State())
	}
}

func TestWindowResize(t *testing.T) {
	app := newTestApp(nil)

	m, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	a := mustApp(t, m)

	if a.width != 120 || a.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", a.width, a.height)
	}
}

func TestCompactLayoutCapsDimensions(t *testing.T) {
	app := newTestAppWithOptions(nil, RunOptions{Compact: true})

	m, _ := app.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	a := mustApp(t, m)

	if got := a.layoutWidth(); got != compactMaxWidth {
		t.Fatalf("expected compact width %d, got %d", compactMaxWidth, got)
	}
	if got := a.listHeight(); got != compactMaxListHeight {
		t.Fatalf("expected compact list height %d, got %d", compactMaxListHeight, got)
	}
	if got := a.detailHeight(); got != compactMaxDetailHeight {
		t.Fatalf("expected compact detail height %d, got %d", compactMaxDetailHeight, got)
	}
}

func TestInputValidation(t *testing.T) {
	fields := []InputField{
		{Key: "query", Label: "Query", Required: true},
		{Key: "limit", Label: "Limit", Required: false},
	}
	input := NewInputModel("search", "", fields, 80)

	// Should fail with empty required field.
	if input.Validate() {
		t.Error("expected validation to fail with empty required field")
	}
}

func TestMenuCommandsExist(t *testing.T) {
	commands := MenuCommands()
	if len(commands) == 0 {
		t.Error("expected at least one command")
	}

	names := make(map[string]bool)
	for _, cmd := range commands {
		if cmd.Name == "" {
			t.Error("command with empty name")
		}
		if names[cmd.Name] {
			t.Errorf("duplicate command name: %s", cmd.Name)
		}
		names[cmd.Name] = true
	}
}

func TestFooterShowsFilterStatus(t *testing.T) {
	app := newTestApp(nil)
	app.menu.SetFilterText("search")

	view := app.View()
	if !strings.Contains(view, "filter:") {
		t.Fatalf("expected footer to show active filter status, got %q", view)
	}
}

func TestFooterShowsAdaptiveHelp(t *testing.T) {
	app := newTestApp(nil)

	view := app.View()
	if !strings.Contains(view, "enter") {
		t.Fatalf("expected footer to include adaptive help, got %q", view)
	}
}

func TestHelpKeyTogglesExpandedFooter(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateDetail
	app.detail = NewDetailModel(ResultItem{Title: "Test", Detail: "# Heading\n\nContent"}, 80, 24)

	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	a := mustApp(t, m)

	if !a.help.ShowAll {
		t.Fatal("expected full help to be enabled after pressing ?")
	}
	if !strings.Contains(a.View(), "pgup") {
		t.Fatalf("expected expanded help to include page navigation, got %q", a.View())
	}
}

func TestExpandedHelpReservesVerticalSpace(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateResults
	app.results = newResultList("test", []ResultItem{{Title: "One"}}, 80, 20)
	app.width = 80
	app.height = 24

	baseline := app.listHeight()

	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	a := mustApp(t, m)

	if got := a.listHeight(); got >= baseline {
		t.Fatalf("expected expanded help to reduce list height, baseline=%d got=%d", baseline, got)
	}
}

func TestContextualBackHelpTargetsForm(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateResults
	app.selectedCommand = Command{Name: "search", Fields: []InputField{{Key: "query", Label: "Query", Required: true}}}
	app.input = NewInputModel("search", "", app.selectedCommand.Fields, 80)
	app.results = newResultList("test", []ResultItem{{Title: "One"}}, 80, 20)

	view := app.View()
	if !strings.Contains(view, "form") {
		t.Fatalf("expected footer help to mention returning to the form, got %q", view)
	}
}

func TestLoadingViewShowsElapsedTime(t *testing.T) {
	app := newTestApp(nil)
	app.state = StateLoading
	app.selectedCommand = Command{Name: "search"}
	app.loadingStarted = time.Now().Add(-1500 * time.Millisecond)

	view := app.View()
	if !strings.Contains(view, "Elapsed:") {
		t.Fatalf("expected loading view to show elapsed time, got %q", view)
	}
}

func TestViewRendering(t *testing.T) {
	app := newTestApp(nil)

	// Each state should render without panic.
	states := []AppState{StateMenu, StateInput, StateLoading, StateResults, StateDetail, StateError}
	for _, state := range states {
		app.state = state
		switch state {
		case StateMenu:
			// Already set up.
		case StateInput:
			app.input = NewInputModel("test", "", []InputField{{Key: "q", Label: "Q"}}, 80)
		case StateLoading:
			// No setup needed.
		case StateResults:
			app.results = newResultList("test", nil, 80, 20)
		case StateDetail:
			app.detail = NewDetailModel(ResultItem{Title: "Test"}, 80, 24)
		case StateError:
			app.errorMsg = "test error"
		}

		view := app.View()
		if view == "" && state != StateResults {
			t.Errorf("empty view for state %v", state)
		}
	}
}

func TestTrimToLen(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"", 5, ""},
	}
	for _, tc := range tests {
		got := trimToLen(tc.input, tc.limit)
		if got != tc.expected {
			t.Errorf("trimToLen(%q, %d) = %q, want %q", tc.input, tc.limit, got, tc.expected)
		}
	}
}
