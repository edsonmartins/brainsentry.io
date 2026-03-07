package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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
	colSearchScore   = "score"
	colSearchContent = "content"
	colSearchCat     = "category"
	colSearchDate    = "created"
	colSearchFullID  = "_fullID"
)

// SearchResultsMsg is sent when search completes.
type SearchResultsMsg struct {
	Results      []dto.MemoryResponse
	Total        int
	SearchTimeMs int64
}

// SearchErrorMsg is sent when search fails.
type SearchErrorMsg struct{ Err error }

// SearchModel handles semantic search.
type SearchModel struct {
	client      *client.Client
	keys        tuikeys.KeyMap
	searchInput textinput.Model
	results     []dto.MemoryResponse
	resultsTable table.Model
	total       int
	searchTime  int64
	spinner     components.LoadingSpinner
	searching   bool
	hasSearched bool
	width       int
	height      int
}

// NewSearchModel creates a new search view.
func NewSearchModel(c *client.Client) SearchModel {
	si := textinput.New()
	si.Placeholder = "Enter your search query..."
	si.Width = 60
	si.Focus()
	si.PromptStyle = lipgloss.NewStyle().Foreground(theme.Primary)
	si.TextStyle = lipgloss.NewStyle().Foreground(theme.Text)
	si.PlaceholderStyle = lipgloss.NewStyle().Foreground(theme.Overlay0)

	return SearchModel{
		client:      c,
		keys:        tuikeys.DefaultKeyMap(),
		searchInput: si,
		spinner:     components.NewLoadingSpinner("Searching..."),
	}
}

func newSearchResultsTable(results []dto.MemoryResponse, width int) table.Model {
	columns := []table.Column{
		table.NewColumn(colSearchScore, "Score", 8),
		table.NewColumn(colSearchCat, "Category", 14),
		table.NewFlexColumn(colSearchContent, "Content", 3),
		table.NewColumn(colSearchDate, "Created", 12),
	}

	rows := make([]table.Row, 0, len(results))
	for _, mem := range results {
		content := mem.Content
		if len(content) > 70 {
			content = content[:67] + "..."
		}

		scoreCell := table.NewStyledCell(
			fmt.Sprintf("%.2f", mem.RelevanceScore),
			lipgloss.NewStyle().Foreground(theme.Accent).Bold(true),
		)

		catCell := table.NewStyledCell(
			string(mem.Category),
			lipgloss.NewStyle().Foreground(theme.ColorForCategory(string(mem.Category))),
		)

		rows = append(rows, table.NewRow(table.RowData{
			colSearchScore:   scoreCell,
			colSearchContent: content,
			colSearchCat:     catCell,
			colSearchDate:    mem.CreatedAt.Format(time.DateOnly),
			colSearchFullID:  mem.ID,
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

// Init initializes the search view.
func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the search view.
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.hasSearched {
			m.resultsTable = newSearchResultsTable(m.results, msg.Width-2)
		}

	case SearchResultsMsg:
		m.searching = false
		m.hasSearched = true
		m.results = msg.Results
		m.total = msg.Total
		m.searchTime = msg.SearchTimeMs
		m.resultsTable = newSearchResultsTable(m.results, m.width-2)
		return m, nil

	case SearchErrorMsg:
		m.searching = false
		m.hasSearched = true
		m.results = nil
		m.total = 0
		m.resultsTable = newSearchResultsTable(nil, m.width-2)
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			return m, nil
		}

		// If search input is focused
		if m.searchInput.Focused() {
			switch msg.String() {
			case "enter":
				query := m.searchInput.Value()
				if query == "" {
					return m, nil
				}
				m.searching = true
				m.searchInput.Blur()
				c := m.client
				return m, tea.Batch(m.spinner.Tick(), func() tea.Msg {
					resp, err := c.SearchMemories(&dto.SearchRequest{
						Query: query,
						Limit: 20,
					})
					if err != nil {
						return SearchErrorMsg{Err: err}
					}
					return SearchResultsMsg{
						Results:      resp.Results,
						Total:        resp.Total,
						SearchTimeMs: resp.SearchTimeMs,
					}
				})
			case "esc":
				if m.hasSearched {
					m.searchInput.Blur()
					return m, nil
				}
				return m, func() tea.Msg { return NavigateBackMsg{} }
			}
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		// Results navigation
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case key.Matches(msg, m.keys.Enter):
			if m.hasSearched && len(m.results) > 0 {
				row := m.resultsTable.HighlightedRow()
				if fullID, ok := row.Data[colSearchFullID].(string); ok {
					return m, func() tea.Msg {
						return MemorySelectedMsg{ID: fullID}
					}
				}
			}
		case key.Matches(msg, m.keys.Search), msg.String() == "/":
			m.searchInput.Focus()
			return m, textinput.Blink
		}
	}

	if m.searching {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.hasSearched && !m.searchInput.Focused() {
		var cmd tea.Cmd
		m.resultsTable, cmd = m.resultsTable.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the search view.
func (m SearchModel) View() string {
	header := theme.TitleStyle.Render("  Semantic Search")

	searchBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Surface1).
		Padding(0, 1).
		Width(min(70, m.width-4)).
		Render(m.searchInput.View())

	searchBar := "\n  " + searchBox + "\n"

	if m.searching {
		return fmt.Sprintf("\n%s\n%s\n  %s", header, searchBar, m.spinner.View())
	}

	if !m.hasSearched {
		hint := theme.HintStyle.Render("\n  Type a query and press Enter to search.\n  Results are ranked by semantic relevance.\n")
		return fmt.Sprintf("\n%s\n%s%s", header, searchBar, hint)
	}

	// Results
	info := theme.DimStyle.Render(fmt.Sprintf("\n  %d results in %dms\n", m.total, m.searchTime))

	hints := theme.HintStyle.Render("  [/] new search  [j/k] navigate  [Enter] view  [Esc] back")

	if len(m.results) == 0 {
		noResults := theme.DimStyle.Render("  No results found")
		return fmt.Sprintf("\n%s\n%s%s\n%s\n\n%s", header, searchBar, info, noResults, hints)
	}

	return fmt.Sprintf("\n%s\n%s%s\n%s\n\n%s", header, searchBar, info, m.resultsTable.View(), hints)
}
