package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/integraltech/brainsentry/cmd/tui/components"
	tuikeys "github.com/integraltech/brainsentry/cmd/tui/keys"
	"github.com/integraltech/brainsentry/cmd/tui/theme"
	"github.com/integraltech/brainsentry/internal/client"
	"github.com/integraltech/brainsentry/internal/dto"
)

const (
	colID         = "id"
	colContent    = "content"
	colCategory   = "category"
	colImportance = "importance"
	colCreated    = "created"
	colFullID     = "_fullID"
)

// MemoriesLoadedMsg is sent when memory list finishes loading.
type MemoriesLoadedMsg struct {
	Response *dto.MemoryListResponse
}

// MemoriesErrorMsg is sent when memory list loading fails.
type MemoriesErrorMsg struct{ Err error }

// MemoryDeletedMsg is sent when a memory is deleted.
type MemoryDeletedMsg struct{ ID string }

// MemorySelectedMsg requests navigation to memory detail.
type MemorySelectedMsg struct{ ID string }

// MemoryListModel displays a paginated list of memories.
type MemoryListModel struct {
	client     *client.Client
	keys       tuikeys.KeyMap
	table      table.Model
	memories   []dto.MemoryResponse
	page       int
	totalPages int
	totalItems int64
	pageSize   int
	spinner    components.LoadingSpinner
	confirm    components.Confirm
	loading    bool
	err        string
	width      int
	height     int
}

// NewMemoryListModel creates a new memory list view.
func NewMemoryListModel(c *client.Client) MemoryListModel {
	return MemoryListModel{
		client:   c,
		keys:     tuikeys.DefaultKeyMap(),
		page:     1,
		pageSize: 20,
		spinner:  components.NewLoadingSpinner("Loading memories..."),
		loading:  true,
		table:    newMemoryTable(nil, 80),
	}
}

func newMemoryTable(memories []dto.MemoryResponse, width int) table.Model {
	columns := []table.Column{
		table.NewColumn(colID, "ID", 10).
			WithStyle(lipgloss.NewStyle().Foreground(theme.Overlay0)),
		table.NewFlexColumn(colContent, "Content", 3),
		table.NewColumn(colCategory, "Category", 14),
		table.NewColumn(colImportance, "Importance", 12),
		table.NewColumn(colCreated, "Created", 12).
			WithStyle(lipgloss.NewStyle().Foreground(theme.Subtext0)),
	}

	rows := make([]table.Row, 0, len(memories))
	for _, mem := range memories {
		content := mem.Content
		if len(content) > 60 {
			content = content[:57] + "..."
		}

		id := mem.ID
		if len(id) > 8 {
			id = id[:8]
		}

		catCell := table.NewStyledCell(
			string(mem.Category),
			lipgloss.NewStyle().Foreground(theme.ColorForCategory(string(mem.Category))),
		)

		impCell := table.NewStyledCell(
			string(mem.Importance),
			lipgloss.NewStyle().Foreground(theme.ColorForImportance(string(mem.Importance))).Bold(true),
		)

		rows = append(rows, table.NewRow(table.RowData{
			colID:         id,
			colContent:    content,
			colCategory:   catCell,
			colImportance: impCell,
			colCreated:    mem.CreatedAt.Format(time.DateOnly),
			colFullID:     mem.ID,
		}))
	}

	t := table.New(columns).
		WithRows(rows).
		WithTargetWidth(width).
		WithPageSize(20).
		Focused(true).
		SortByAsc(colCreated).
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

// Init starts loading the memory list.
func (m MemoryListModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick(), m.loadMemories())
}

func (m MemoryListModel) loadMemories() tea.Cmd {
	c := m.client
	page := m.page
	size := m.pageSize

	return func() tea.Msg {
		resp, err := c.ListMemories(page, size)
		if err != nil {
			return MemoriesErrorMsg{Err: err}
		}
		return MemoriesLoadedMsg{Response: resp}
	}
}

// Update handles messages for the memory list.
func (m MemoryListModel) Update(msg tea.Msg) (MemoryListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.loading {
			m.table = newMemoryTable(m.memories, msg.Width-2)
		}

	case MemoriesLoadedMsg:
		m.loading = false
		m.memories = msg.Response.Memories
		m.totalPages = msg.Response.TotalPages
		m.totalItems = msg.Response.TotalElements
		m.table = newMemoryTable(m.memories, m.width-2)
		return m, nil

	case MemoriesErrorMsg:
		m.loading = false
		m.err = msg.Err.Error()
		return m, nil

	case MemoryDeletedMsg:
		m.loading = true
		return m, tea.Batch(m.spinner.Tick(), m.loadMemories())

	case components.ConfirmResult:
		if msg.Tag == "delete-memory" && msg.Confirmed {
			row := m.table.HighlightedRow()
			if fullID, ok := row.Data[colFullID].(string); ok {
				c := m.client
				return m, func() tea.Msg {
					if err := c.DeleteMemory(fullID); err != nil {
						return MemoriesErrorMsg{Err: err}
					}
					return MemoryDeletedMsg{ID: fullID}
				}
			}
		}
		m.confirm = components.Confirm{}
		return m, nil

	case tea.KeyMsg:
		if m.confirm.Active {
			var cmd tea.Cmd
			m.confirm, cmd = m.confirm.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case key.Matches(msg, m.keys.Enter):
			row := m.table.HighlightedRow()
			if fullID, ok := row.Data[colFullID].(string); ok {
				return m, func() tea.Msg { return MemorySelectedMsg{ID: fullID} }
			}
		case key.Matches(msg, m.keys.Delete):
			row := m.table.HighlightedRow()
			if id, ok := row.Data[colFullID].(string); ok {
				short := id
				if len(short) > 8 {
					short = short[:8]
				}
				m.confirm = components.NewConfirm(
					fmt.Sprintf("Delete memory %s?", short),
					"delete-memory",
				)
				return m, nil
			}
		case key.Matches(msg, m.keys.PageDown):
			if m.page < m.totalPages {
				m.page++
				m.loading = true
				return m, tea.Batch(m.spinner.Tick(), m.loadMemories())
			}
		case key.Matches(msg, m.keys.PageUp):
			if m.page > 1 {
				m.page--
				m.loading = true
				return m, tea.Batch(m.spinner.Tick(), m.loadMemories())
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

// View renders the memory list.
func (m MemoryListModel) View() string {
	if m.confirm.Active {
		return lipgloss.Place(m.width, m.height-1,
			lipgloss.Center, lipgloss.Center,
			m.confirm.View(),
		)
	}

	if m.loading {
		return lipgloss.Place(m.width, m.height-1,
			lipgloss.Center, lipgloss.Center,
			m.spinner.View(),
		)
	}

	if m.err != "" {
		return theme.ErrorStyle.Render("Error: " + m.err)
	}

	header := theme.TitleStyle.Render(fmt.Sprintf("  Memories (%d total)", m.totalItems))

	pagination := theme.HintStyle.Render(
		fmt.Sprintf("  Page %d of %d  |  Ctrl+D/U: next/prev  /: filter  n: new  d: delete  Enter: view  Esc: back",
			m.page, m.totalPages),
	)

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", header, m.table.View(), pagination)
}

// Refresh reloads the memory list.
func (m *MemoryListModel) Refresh() tea.Cmd {
	m.loading = true
	return tea.Batch(m.spinner.Tick(), m.loadMemories())
}
