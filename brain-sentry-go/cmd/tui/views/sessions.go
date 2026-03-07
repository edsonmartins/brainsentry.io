package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/integraltech/brainsentry/cmd/tui/components"
	tuikeys "github.com/integraltech/brainsentry/cmd/tui/keys"
	"github.com/integraltech/brainsentry/cmd/tui/theme"
	"github.com/integraltech/brainsentry/internal/client"
	"github.com/integraltech/brainsentry/internal/domain"
)

const (
	colSessID      = "id"
	colSessStatus  = "status"
	colSessStarted = "started"
	colSessMem     = "memories"
	colSessInt     = "interceptions"
	colSessNotes   = "notes"
	colSessFullID  = "_fullID"
)

// SessionsLoadedMsg is sent when sessions finish loading.
type SessionsLoadedMsg struct {
	Sessions []domain.Session
}

// SessionsErrorMsg is sent when sessions loading fails.
type SessionsErrorMsg struct{ Err error }

// SessionsModel displays a list of development sessions.
type SessionsModel struct {
	client   *client.Client
	keys     tuikeys.KeyMap
	sessions []domain.Session
	table    table.Model
	spinner  components.LoadingSpinner
	loading  bool
	err      string
	width    int
	height   int
}

// NewSessionsModel creates a new sessions view.
func NewSessionsModel(c *client.Client) SessionsModel {
	return SessionsModel{
		client:  c,
		keys:    tuikeys.DefaultKeyMap(),
		spinner: components.NewLoadingSpinner("Loading sessions..."),
		loading: true,
		table:   newSessionsTable(nil, 80),
	}
}

func newSessionsTable(sessions []domain.Session, width int) table.Model {
	columns := []table.Column{
		table.NewColumn(colSessID, "ID", 10).
			WithStyle(lipgloss.NewStyle().Foreground(theme.Overlay0)),
		table.NewColumn(colSessStatus, "Status", 12),
		table.NewColumn(colSessStarted, "Started", 18),
		table.NewColumn(colSessMem, "Mem", 6),
		table.NewColumn(colSessInt, "Int", 6),
		table.NewColumn(colSessNotes, "Notes", 6),
	}

	rows := make([]table.Row, 0, len(sessions))
	for _, s := range sessions {
		id := s.ID
		if len(id) > 8 {
			id = id[:8]
		}

		var statusColor lipgloss.Color
		switch s.Status {
		case domain.SessionActive:
			statusColor = theme.Secondary
		case domain.SessionExpired:
			statusColor = theme.Danger
		case domain.SessionCompleted:
			statusColor = theme.Overlay0
		default:
			statusColor = theme.Text
		}

		statusCell := table.NewStyledCell(
			string(s.Status),
			lipgloss.NewStyle().Foreground(statusColor).Bold(true),
		)

		started := s.StartedAt.Format("2006-01-02 15:04")

		rows = append(rows, table.NewRow(table.RowData{
			colSessID:      id,
			colSessStatus:  statusCell,
			colSessStarted: started,
			colSessMem:     fmt.Sprintf("%d", s.MemoryCount),
			colSessInt:     fmt.Sprintf("%d", s.InterceptionCount),
			colSessNotes:   fmt.Sprintf("%d", s.NoteCount),
			colSessFullID:  s.ID,
		}))
	}

	t := table.New(columns).
		WithRows(rows).
		WithTargetWidth(width).
		WithPageSize(20).
		Focused(true).
		SortByDesc(colSessStarted).
		WithBaseStyle(lipgloss.NewStyle().
			Foreground(theme.Text).
			Align(lipgloss.Left)).
		HeaderStyle(lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true).
			Align(lipgloss.Left)).
		HighlightStyle(lipgloss.NewStyle().
			Background(theme.Surface0).
			Foreground(theme.Text).
			Bold(true)).
		BorderRounded()

	return t
}

// Init starts loading sessions.
func (m SessionsModel) Init() tea.Cmd {
	c := m.client
	return tea.Batch(m.spinner.Tick(), func() tea.Msg {
		sessions, err := c.ListSessions()
		if err != nil {
			return SessionsErrorMsg{Err: err}
		}
		return SessionsLoadedMsg{Sessions: sessions}
	})
}

// Update handles messages for the sessions view.
func (m SessionsModel) Update(msg tea.Msg) (SessionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.loading {
			m.table = newSessionsTable(m.sessions, msg.Width-2)
		}

	case SessionsLoadedMsg:
		m.loading = false
		m.sessions = msg.Sessions
		m.table = newSessionsTable(m.sessions, m.width-2)
		return m, nil

	case SessionsErrorMsg:
		m.loading = false
		m.err = msg.Err.Error()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return NavigateBackMsg{} }
		}
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the sessions list.
func (m SessionsModel) View() string {
	if m.loading {
		return lipgloss.Place(m.width, m.height-1,
			lipgloss.Center, lipgloss.Center,
			m.spinner.View(),
		)
	}

	if m.err != "" {
		return theme.ErrorStyle.Render("Error: " + m.err)
	}

	header := theme.TitleStyle.Render(fmt.Sprintf("  Sessions (%d)", len(m.sessions)))

	hints := theme.HintStyle.Render("  [j/k] navigate  [Esc] back")

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", header, m.table.View(), hints)
}
