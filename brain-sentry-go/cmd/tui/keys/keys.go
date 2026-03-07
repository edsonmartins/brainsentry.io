package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines global keybindings for the TUI.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Enter    key.Binding
	Back     key.Binding
	Search   key.Binding
	New      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Save     key.Binding
	Help     key.Binding
	Quit     key.Binding
	Tab      key.Binding
	PageUp   key.Binding
	PageDown key.Binding

	// View navigation
	ViewDashboard  key.Binding
	ViewMemories   key.Binding
	ViewSearch     key.Binding
	ViewSessions   key.Binding
	ViewRelations  key.Binding
}

// DefaultKeyMap returns the default Vim-style keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "page down"),
		),
		ViewDashboard: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "dashboard"),
		),
		ViewMemories: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "memories"),
		),
		ViewSearch: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "search"),
		),
		ViewSessions: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "sessions"),
		),
		ViewRelations: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "relationships"),
		),
	}
}
