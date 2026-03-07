package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoadingSpinner wraps a bubbles spinner with a message.
type LoadingSpinner struct {
	Spinner spinner.Model
	Message string
}

// NewLoadingSpinner creates a spinner with a message.
func NewLoadingSpinner(message string) LoadingSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
	return LoadingSpinner{Spinner: s, Message: message}
}

// Update processes spinner messages.
func (l LoadingSpinner) Update(msg tea.Msg) (LoadingSpinner, tea.Cmd) {
	var cmd tea.Cmd
	l.Spinner, cmd = l.Spinner.Update(msg)
	return l, cmd
}

// View renders the spinner with its message.
func (l LoadingSpinner) View() string {
	return l.Spinner.View() + " " + l.Message
}

// Tick returns the spinner's tick command.
func (l LoadingSpinner) Tick() tea.Cmd {
	return l.Spinner.Tick
}
