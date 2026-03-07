package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastLevel indicates the severity of a toast notification.
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastDismissMsg is sent when a toast should be dismissed.
type ToastDismissMsg struct{}

// Toast represents a temporary notification message.
type Toast struct {
	Message string
	Level   ToastLevel
	Visible bool
}

var (
	toastInfoStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3B82F6")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2).
			Bold(true)

	toastSuccessStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#10B981")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 2).
				Bold(true)

	toastWarningStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#F59E0B")).
				Foreground(lipgloss.Color("#000000")).
				Padding(0, 2).
				Bold(true)

	toastErrorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#EF4444")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2).
			Bold(true)
)

// Show displays a toast notification.
func (t *Toast) Show(msg string, level ToastLevel) tea.Cmd {
	t.Message = msg
	t.Level = level
	t.Visible = true
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return ToastDismissMsg{}
	})
}

// Dismiss hides the toast.
func (t *Toast) Dismiss() {
	t.Visible = false
}

// View renders the toast notification.
func (t Toast) View() string {
	if !t.Visible {
		return ""
	}

	var style lipgloss.Style
	switch t.Level {
	case ToastSuccess:
		style = toastSuccessStyle
	case ToastWarning:
		style = toastWarningStyle
	case ToastError:
		style = toastErrorStyle
	default:
		style = toastInfoStyle
	}

	return style.Render(t.Message)
}
