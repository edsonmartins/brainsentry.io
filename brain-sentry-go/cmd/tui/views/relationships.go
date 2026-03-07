package views

import (
	"fmt"
	"strings"

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
	colRelFrom     = "from"
	colRelTo       = "to"
	colRelType     = "type"
	colRelStrength = "strength"
	colRelTargetID = "_targetID"
)

// RelationshipsLoadedMsg is sent when relationships finish loading.
type RelationshipsLoadedMsg struct {
	Relationships []domain.MemoryRelationship
}

// RelationshipsErrorMsg is sent when loading fails.
type RelationshipsErrorMsg struct{ Err error }

// RelationshipsModel displays relationships for a memory.
type RelationshipsModel struct {
	client        *client.Client
	keys          tuikeys.KeyMap
	memoryID      string
	relationships []domain.MemoryRelationship
	table         table.Model
	spinner       components.LoadingSpinner
	loading       bool
	err           string
	width         int
	height        int
}

// NewRelationshipsModel creates a new relationships view.
func NewRelationshipsModel(c *client.Client, memoryID string) RelationshipsModel {
	return RelationshipsModel{
		client:   c,
		keys:     tuikeys.DefaultKeyMap(),
		memoryID: memoryID,
		spinner:  components.NewLoadingSpinner("Loading relationships..."),
		loading:  true,
		table:    newRelationshipsTable(nil, "", 80),
	}
}

func strengthBar(strength float64) string {
	filled := int(strength * 10)
	if filled < 0 {
		filled = 0
	}
	if filled > 10 {
		filled = 10
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
	return fmt.Sprintf("%s %.0f%%", bar, strength*100)
}

func newRelationshipsTable(rels []domain.MemoryRelationship, memoryID string, width int) table.Model {
	columns := []table.Column{
		table.NewColumn(colRelFrom, "From", 10).
			WithStyle(lipgloss.NewStyle().Foreground(theme.Overlay0)),
		table.NewColumn(colRelTo, "To", 10).
			WithStyle(lipgloss.NewStyle().Foreground(theme.Overlay0)),
		table.NewColumn(colRelType, "Type", 16),
		table.NewFlexColumn(colRelStrength, "Strength", 2),
	}

	rows := make([]table.Row, 0, len(rels))
	for _, rel := range rels {
		fromID := rel.FromMemoryID
		toID := rel.ToMemoryID
		if len(fromID) > 8 {
			fromID = fromID[:8]
		}
		if len(toID) > 8 {
			toID = toID[:8]
		}

		typeCell := table.NewStyledCell(
			rel.Type,
			lipgloss.NewStyle().Foreground(theme.Accent).Bold(true),
		)

		strengthCell := table.NewStyledCell(
			strengthBar(rel.Strength),
			lipgloss.NewStyle().Foreground(theme.Secondary),
		)

		targetID := rel.ToMemoryID
		if targetID == memoryID {
			targetID = rel.FromMemoryID
		}

		rows = append(rows, table.NewRow(table.RowData{
			colRelFrom:     fromID,
			colRelTo:       toID,
			colRelType:     typeCell,
			colRelStrength: strengthCell,
			colRelTargetID: targetID,
		}))
	}

	t := table.New(columns).
		WithRows(rows).
		WithTargetWidth(width).
		WithPageSize(20).
		Focused(true).
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

// Init starts loading relationships.
func (m RelationshipsModel) Init() tea.Cmd {
	c := m.client
	id := m.memoryID
	return tea.Batch(m.spinner.Tick(), func() tea.Msg {
		rels, err := c.GetMemoryRelationships(id)
		if err != nil {
			return RelationshipsErrorMsg{Err: err}
		}
		return RelationshipsLoadedMsg{Relationships: rels}
	})
}

// Update handles messages for the relationships view.
func (m RelationshipsModel) Update(msg tea.Msg) (RelationshipsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.loading {
			m.table = newRelationshipsTable(m.relationships, m.memoryID, msg.Width-2)
		}

	case RelationshipsLoadedMsg:
		m.loading = false
		m.relationships = msg.Relationships
		m.table = newRelationshipsTable(m.relationships, m.memoryID, m.width-2)
		return m, nil

	case RelationshipsErrorMsg:
		m.loading = false
		m.err = msg.Err.Error()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case key.Matches(msg, m.keys.Enter):
			if len(m.relationships) > 0 {
				row := m.table.HighlightedRow()
				if targetID, ok := row.Data[colRelTargetID].(string); ok {
					return m, func() tea.Msg { return MemorySelectedMsg{ID: targetID} }
				}
			}
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

// View renders the relationships view.
func (m RelationshipsModel) View() string {
	if m.loading {
		return lipgloss.Place(m.width, m.height-1,
			lipgloss.Center, lipgloss.Center,
			m.spinner.View(),
		)
	}

	if m.err != "" {
		return theme.ErrorStyle.Render("Error: " + m.err)
	}

	id := m.memoryID
	if len(id) > 8 {
		id = id[:8]
	}
	header := theme.TitleStyle.Render(fmt.Sprintf("  Relationships for %s (%d)", id, len(m.relationships)))

	hints := theme.HintStyle.Render("  [j/k] navigate  [Enter] go to memory  [Esc] back")

	if len(m.relationships) == 0 {
		noData := theme.DimStyle.Render("  No relationships found")
		return fmt.Sprintf("\n%s\n\n%s\n\n%s", header, noData, hints)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", header, m.table.View(), hints)
}
