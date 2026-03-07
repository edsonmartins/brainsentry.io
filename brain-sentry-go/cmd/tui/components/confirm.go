package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmResult is sent when the user confirms or cancels.
type ConfirmResult struct {
	Confirmed bool
	Tag       string // identifies which confirmation this is
}

// Confirm renders a yes/no confirmation dialog.
type Confirm struct {
	Message  string
	Tag      string
	Active   bool
	selected int // 0 = yes, 1 = no
}

var (
	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F59E0B")).
			Padding(1, 3)

	confirmBtnActive = lipgloss.NewStyle().
				Background(lipgloss.Color("#EF4444")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 2).
				Bold(true)

	confirmBtnInactive = lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Foreground(lipgloss.Color("#9CA3AF")).
				Padding(0, 2)
)

// NewConfirm creates a new confirmation dialog.
func NewConfirm(message, tag string) Confirm {
	return Confirm{
		Message:  message,
		Tag:      tag,
		Active:   true,
		selected: 1, // default to No
	}
}

// Update handles key input for the confirmation dialog.
func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) {
	if !c.Active {
		return c, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			c.selected = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			c.selected = 1
		case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
			c.Active = false
			return c, func() tea.Msg { return ConfirmResult{Confirmed: true, Tag: c.Tag} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc"))):
			c.Active = false
			return c, func() tea.Msg { return ConfirmResult{Confirmed: false, Tag: c.Tag} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			c.Active = false
			return c, func() tea.Msg { return ConfirmResult{Confirmed: c.selected == 0, Tag: c.Tag} }
		}
	}

	return c, nil
}

// View renders the confirmation dialog.
func (c Confirm) View() string {
	if !c.Active {
		return ""
	}

	yesStyle := confirmBtnInactive
	noStyle := confirmBtnInactive
	if c.selected == 0 {
		yesStyle = confirmBtnActive
	} else {
		noStyle = confirmBtnActive
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		yesStyle.Render("  Yes  "),
		"  ",
		noStyle.Render("  No  "),
	)

	content := fmt.Sprintf("%s\n\n%s", c.Message, buttons)
	return confirmBoxStyle.Render(content)
}
