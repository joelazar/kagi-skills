package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputField describes a single input field for a command.
type InputField struct {
	Key         string
	Label       string
	Placeholder string
	Required    bool
	Value       string
	Sensitive   bool
	MaxLength   int
}

// InputModel manages a set of text input fields for a command.
type InputModel struct {
	fields  []InputField
	inputs  []textinput.Model
	focused int
	width   int
	command string
	hint    string
}

// NewInputModel creates an input form for the given command and its fields.
func NewInputModel(command, hint string, fields []InputField, width int) InputModel {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		if f.MaxLength > 0 {
			ti.CharLimit = f.MaxLength
		} else {
			ti.CharLimit = 4000
		}
		ti.Width = min(width-10, 60)
		ti.SetValue(f.Value)
		if f.Sensitive {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}

	return InputModel{
		fields:  fields,
		inputs:  inputs,
		focused: 0,
		width:   width,
		command: command,
		hint:    hint,
	}
}

// Values returns the current input values as a map.
func (m *InputModel) Values() map[string]string {
	vals := make(map[string]string, len(m.fields))
	for i, f := range m.fields {
		vals[f.Key] = strings.TrimSpace(m.inputs[i].Value())
	}
	return vals
}

// Validate checks that all required fields have values.
func (m *InputModel) Validate() bool {
	for i, f := range m.fields {
		if f.Required && strings.TrimSpace(m.inputs[i].Value()) == "" {
			return false
		}
	}
	return true
}

// Update handles input messages.
func (m *InputModel) Update(msg tea.Msg) (*InputModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
			return m, nil
		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused - 1 + len(m.inputs)) % len(m.inputs)
			m.inputs[m.focused].Focus()
			return m, nil
		}
	}

	// Update focused input.
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)

	return m, cmd
}

// View renders the input form.
func (m *InputModel) View() string {
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorHighlight)
	requiredStyle := lipgloss.NewStyle().Foreground(ColorError)

	var sb strings.Builder
	sb.WriteString(TitleStyle.Render("  "+m.command) + "\n\n")
	if m.hint != "" {
		sb.WriteString(DimStyle.MaxWidth(max(m.width-4, 20)).Render(m.hint) + "\n\n")
	}

	for i, f := range m.fields {
		label := labelStyle.Render(f.Label)
		if f.Required {
			label += requiredStyle.Render(" *")
		}
		sb.WriteString(label + "\n")
		sb.WriteString(m.inputs[i].View() + "\n\n")
	}

	submitStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
	if m.Validate() {
		sb.WriteString(submitStyle.Render("Press Enter to submit"))
	} else {
		sb.WriteString(DimStyle.Render("Fill in required fields (*)"))
	}

	return sb.String()
}

// SetWidth updates the input widths.
func (m *InputModel) SetWidth(width int) {
	m.width = width
	w := min(width-10, 60)
	for i := range m.inputs {
		m.inputs[i].Width = w
	}
}
