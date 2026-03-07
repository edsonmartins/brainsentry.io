package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/integraltech/brainsentry/cmd/tui/components"
	tuikeys "github.com/integraltech/brainsentry/cmd/tui/keys"
	"github.com/integraltech/brainsentry/cmd/tui/theme"
	"github.com/integraltech/brainsentry/internal/client"
	"github.com/integraltech/brainsentry/internal/dto"
)

// MemoryDetailLoadedMsg is sent when memory detail finishes loading.
type MemoryDetailLoadedMsg struct {
	Memory *dto.MemoryResponse
}

// MemoryDetailErrorMsg is sent when loading fails.
type MemoryDetailErrorMsg struct{ Err error }

// NavigateBackMsg requests navigation back to the previous view.
type NavigateBackMsg struct{}

// EditMemoryMsg requests navigation to edit a memory.
type EditMemoryMsg struct{ ID string }

// ViewRelationshipsMsg requests navigation to relationships view.
type ViewRelationshipsMsg struct{ ID string }

// MemoryDetailModel displays full details of a single memory.
type MemoryDetailModel struct {
	client   *client.Client
	keys     tuikeys.KeyMap
	memory   *dto.MemoryResponse
	viewport viewport.Model
	spinner  components.LoadingSpinner
	confirm  components.Confirm
	loading  bool
	err      string
	memoryID string
	width    int
	height   int
}

// NewMemoryDetailModel creates a new memory detail view.
func NewMemoryDetailModel(c *client.Client, memoryID string) MemoryDetailModel {
	return MemoryDetailModel{
		client:   c,
		keys:     tuikeys.DefaultKeyMap(),
		memoryID: memoryID,
		spinner:  components.NewLoadingSpinner("Loading memory..."),
		loading:  true,
	}
}

// Init starts loading the memory detail.
func (m MemoryDetailModel) Init() tea.Cmd {
	c := m.client
	id := m.memoryID
	return tea.Batch(m.spinner.Tick(), func() tea.Msg {
		memory, err := c.GetMemory(id)
		if err != nil {
			return MemoryDetailErrorMsg{Err: err}
		}
		return MemoryDetailLoadedMsg{Memory: memory}
	})
}

// Update handles messages for the memory detail.
func (m MemoryDetailModel) Update(msg tea.Msg) (MemoryDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 5

	case MemoryDetailLoadedMsg:
		m.loading = false
		m.memory = msg.Memory
		m.viewport = viewport.New(m.width-4, m.height-5)
		m.viewport.SetContent(m.renderContent())
		return m, nil

	case MemoryDetailErrorMsg:
		m.loading = false
		m.err = msg.Err.Error()
		return m, nil

	case components.ConfirmResult:
		if msg.Tag == "delete-detail" && msg.Confirmed {
			c := m.client
			id := m.memoryID
			return m, func() tea.Msg {
				if err := c.DeleteMemory(id); err != nil {
					return MemoryDetailErrorMsg{Err: err}
				}
				return NavigateBackMsg{}
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
		case key.Matches(msg, m.keys.Edit):
			return m, func() tea.Msg { return EditMemoryMsg{ID: m.memoryID} }
		case key.Matches(msg, m.keys.Delete):
			m.confirm = components.NewConfirm("Delete this memory?", "delete-detail")
			return m, nil
		case msg.String() == "r":
			return m, func() tea.Msg { return ViewRelationshipsMsg{ID: m.memoryID} }
		}
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m MemoryDetailModel) renderContent() string {
	mem := m.memory
	if mem == nil {
		return ""
	}

	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentWidth > 100 {
		contentWidth = 100
	}

	var b strings.Builder

	// Header with ID
	idBox := lipgloss.NewStyle().
		Foreground(theme.Overlay0).
		Render(mem.ID)
	b.WriteString("  " + idBox + "\n\n")

	// Metadata badges row
	catBadge := theme.Badge(string(mem.Category), lipgloss.Color("#FFFFFF"), theme.ColorForCategory(string(mem.Category)))
	impBadge := theme.Badge(string(mem.Importance), lipgloss.Color("#FFFFFF"), theme.ColorForImportance(string(mem.Importance)))
	statusBadge := theme.Badge(string(mem.ValidationStatus), lipgloss.Color("#FFFFFF"), theme.Surface0)
	typeBadge := theme.Badge(string(mem.MemoryType), lipgloss.Color("#FFFFFF"), theme.Surface0)
	b.WriteString("  " + catBadge + " " + impBadge + " " + statusBadge + " " + typeBadge + "\n\n")

	// Content rendered as markdown with glamour
	contentMd := mem.Content
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dracula"),
		glamour.WithWordWrap(contentWidth),
	)
	if err == nil {
		rendered, renderErr := renderer.Render(contentMd)
		if renderErr == nil {
			contentMd = rendered
		}
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Surface1).
		Padding(0, 1).
		Width(contentWidth).
		Render(contentMd)
	b.WriteString(contentBox + "\n")

	// Summary
	if mem.Summary != "" {
		b.WriteString("\n")
		summaryBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Surface0).
			Padding(0, 1).
			Width(contentWidth).
			Foreground(theme.Subtext0).
			Italic(true).
			Render(mem.Summary)
		b.WriteString(summaryBox + "\n")
	}

	// Tags as badges
	if len(mem.Tags) > 0 {
		b.WriteString("\n  ")
		for _, tag := range mem.Tags {
			b.WriteString(theme.TagStyle.Render(tag) + " ")
		}
		b.WriteString("\n")
	}

	// Code example with glamour
	if mem.CodeExample != "" {
		b.WriteString("\n")
		codeMd := "```\n" + mem.CodeExample + "\n```"
		if renderer != nil {
			rendered, renderErr := renderer.Render(codeMd)
			if renderErr == nil {
				codeMd = rendered
			}
		}
		codeBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Surface1).
			Padding(0, 1).
			Width(contentWidth).
			Render(codeMd)
		b.WriteString(codeBox + "\n")
	}

	// Stats section
	b.WriteString("\n")
	statsTitle := theme.SectionStyle.Render("  Stats")
	b.WriteString(statsTitle + "\n")

	statsContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Surface0).
		Padding(0, 2).
		Width(contentWidth)

	label := lipgloss.NewStyle().Foreground(theme.Subtext0).Width(16)
	val := lipgloss.NewStyle().Foreground(theme.Text)

	statsRows := label.Render("Created") + val.Render(mem.CreatedAt.Format("2006-01-02 15:04")) + "\n" +
		label.Render("Updated") + val.Render(mem.UpdatedAt.Format("2006-01-02 15:04")) + "\n" +
		label.Render("Version") + val.Render(fmt.Sprintf("%d", mem.Version)) + "\n" +
		label.Render("Access Count") + val.Render(fmt.Sprintf("%d", mem.AccessCount)) + "\n" +
		label.Render("Helpfulness") + val.Render(fmt.Sprintf("%.0f%%", mem.HelpfulnessRate*100))

	b.WriteString(statsContent.Render(statsRows) + "\n")

	// Related memories
	if len(mem.RelatedMemories) > 0 {
		b.WriteString("\n")
		relTitle := theme.SectionStyle.Render("  Related Memories")
		b.WriteString(relTitle + "\n")

		relBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Surface0).
			Padding(0, 2).
			Width(contentWidth)

		var relRows string
		for _, rel := range mem.RelatedMemories {
			summary := rel.Summary
			if summary == "" && len(rel.ID) >= 8 {
				summary = rel.ID[:8]
			}
			relType := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render(string(rel.RelationshipType))
			strength := lipgloss.NewStyle().Foreground(theme.Secondary).Render(fmt.Sprintf("%.0f%%", rel.Strength*100))
			relRows += fmt.Sprintf("%s  [%s]  %s\n", summary, relType, strength)
		}
		b.WriteString(relBox.Render(relRows) + "\n")
	}

	return b.String()
}

// View renders the memory detail view.
func (m MemoryDetailModel) View() string {
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

	header := theme.TitleStyle.Render("  Memory Detail")
	hints := theme.HintStyle.Render("  [e] edit  [d] delete  [r] relationships  [j/k] scroll  [Esc] back")

	return fmt.Sprintf("\n%s  %s\n\n%s", header, hints, m.viewport.View())
}
