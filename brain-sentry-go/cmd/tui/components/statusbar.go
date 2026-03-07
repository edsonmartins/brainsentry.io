package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/integraltech/brainsentry/cmd/tui/theme"
)

// StatusBar renders the bottom status bar with user info and hints.
type StatusBar struct {
	Width    int
	User     string
	Tenant   string
	View     string
	HintKeys string
}

// Render returns the rendered status bar string.
func (s StatusBar) Render() string {
	left := theme.StatusBarAccentStyle.Render(fmt.Sprintf(" %s ", s.View))

	right := ""
	if s.User != "" {
		right = theme.StatusBarStyle.Render(fmt.Sprintf(" %s @ %s ", s.User, s.Tenant))
	}

	hints := ""
	if s.HintKeys != "" {
		hints = theme.StatusBarStyle.Render(s.HintKeys)
	}

	// Fill the gap
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right) + lipgloss.Width(hints)
	gap := s.Width - leftLen - rightLen
	if gap < 0 {
		gap = 0
	}

	filler := lipgloss.NewStyle().
		Background(theme.Surface0).
		Render(fmt.Sprintf("%*s", gap, ""))

	return left + filler + hints + right
}
