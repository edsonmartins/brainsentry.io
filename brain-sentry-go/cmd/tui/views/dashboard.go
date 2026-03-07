package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/integraltech/brainsentry/cmd/tui/components"
	"github.com/integraltech/brainsentry/cmd/tui/theme"
	"github.com/integraltech/brainsentry/internal/client"
	"github.com/integraltech/brainsentry/internal/dto"
)

// StatsLoadedMsg is sent when stats finish loading.
type StatsLoadedMsg struct {
	Stats          *dto.StatsResponse
	RecentMemories []dto.MemoryResponse
}

// StatsErrorMsg is sent when stats loading fails.
type StatsErrorMsg struct{ Err error }

// DashboardModel displays system stats and recent memories.
type DashboardModel struct {
	client         *client.Client
	stats          *dto.StatsResponse
	recentMemories []dto.MemoryResponse
	spinner        components.LoadingSpinner
	loading        bool
	err            string
	width          int
	height         int
}

// NewDashboardModel creates a new dashboard view.
func NewDashboardModel(c *client.Client) DashboardModel {
	return DashboardModel{
		client:  c,
		spinner: components.NewLoadingSpinner("Loading dashboard..."),
		loading: true,
	}
}

// Init starts loading dashboard data.
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick(), m.loadStats())
}

func (m DashboardModel) loadStats() tea.Cmd {
	c := m.client
	return func() tea.Msg {
		stats, err := c.GetStats()
		if err != nil {
			return StatsErrorMsg{Err: err}
		}

		list, err := c.ListMemories(1, 5)
		var recent []dto.MemoryResponse
		if err == nil {
			recent = list.Memories
		}

		return StatsLoadedMsg{Stats: stats, RecentMemories: recent}
	}
}

// Update handles messages for the dashboard.
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case StatsLoadedMsg:
		m.loading = false
		m.stats = msg.Stats
		m.recentMemories = msg.RecentMemories
		return m, nil

	case StatsErrorMsg:
		m.loading = false
		m.err = msg.Err.Error()
		return m, nil
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func statCard(title string, value string, icon string, accent lipgloss.Color, width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Subtext0)

	valueStyle := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)

	iconStyle := lipgloss.NewStyle().
		Foreground(accent)

	inner := iconStyle.Render(icon) + " " + titleStyle.Render(title) + "\n" +
		valueStyle.Render(value)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Surface1).
		Padding(1, 2).
		Width(width).
		Render(inner)
}

func metricRow(label string, value string) string {
	l := lipgloss.NewStyle().Foreground(theme.Subtext0).Width(18).Render(label)
	v := lipgloss.NewStyle().Foreground(theme.Text).Bold(true).Render(value)
	return l + v
}

// View renders the dashboard.
func (m DashboardModel) View() string {
	if m.loading {
		return lipgloss.Place(m.width, m.height-1,
			lipgloss.Center, lipgloss.Center,
			m.spinner.View(),
		)
	}

	if m.err != "" {
		return lipgloss.Place(m.width, m.height-1,
			lipgloss.Center, lipgloss.Center,
			theme.ErrorStyle.Render("Error: "+m.err),
		)
	}

	if m.stats == nil {
		return ""
	}

	// Calculate card width based on terminal width
	cardWidth := (m.width - 8) / 4
	if cardWidth < 20 {
		cardWidth = 20
	}
	if cardWidth > 30 {
		cardWidth = 30
	}

	// Stats cards with icons
	totalCard := statCard("Total Memories", fmt.Sprintf("%d", m.stats.TotalMemories), "◆", theme.Primary, cardWidth)

	categoriesCount := len(m.stats.MemoriesByCategory)
	catCard := statCard("Categories", fmt.Sprintf("%d", categoriesCount), "◈", theme.Info, cardWidth)

	criticalCount := int64(0)
	if v, ok := m.stats.MemoriesByImportance["CRITICAL"]; ok {
		criticalCount = v
	}
	critCard := statCard("Critical", fmt.Sprintf("%d", criticalCount), "▲", theme.Danger, cardWidth)

	activeCard := statCard("Active 24h", fmt.Sprintf("%d", m.stats.ActiveMemories24h), "●", theme.Secondary, cardWidth)

	cards := lipgloss.JoinHorizontal(lipgloss.Top, totalCard, " ", catCard, " ", critCard, " ", activeCard)

	// Build two-column layout for charts
	chartWidth := m.width / 2
	if chartWidth > 55 {
		chartWidth = 55
	}

	// Category bar chart
	var catItems []components.BarChartItem
	for cat, count := range m.stats.MemoriesByCategory {
		catItems = append(catItems, components.BarChartItem{
			Label: cat,
			Value: count,
			Color: string(theme.ColorForCategory(cat)),
		})
	}
	leftChart := ""
	if len(catItems) > 0 {
		leftChart = components.HorizontalBarChart("Memories by Category", catItems, chartWidth)
	}

	// Importance bar chart
	var impItems []components.BarChartItem
	for imp, count := range m.stats.MemoriesByImportance {
		impItems = append(impItems, components.BarChartItem{
			Label: imp,
			Value: count,
			Color: string(theme.ColorForImportance(imp)),
		})
	}
	rightChart := ""
	if len(impItems) > 0 {
		rightChart = components.HorizontalBarChart("Memories by Importance", impItems, chartWidth)
	}

	chartsRow := ""
	if leftChart != "" || rightChart != "" {
		chartsRow = "\n" + lipgloss.JoinHorizontal(lipgloss.Top, leftChart, "  ", rightChart)
	}

	// System metrics in a box
	metricsSection := ""
	if m.stats.RequestsToday > 0 || m.stats.TotalInjections > 0 {
		metricsBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Surface1).
			Padding(0, 2).
			Width(min(60, m.width-4))

		metrics := metricRow("Requests Today", fmt.Sprintf("%d", m.stats.RequestsToday)) + "\n" +
			metricRow("Injections", fmt.Sprintf("%d (rate: %.1f%%)", m.stats.TotalInjections, m.stats.InjectionRate*100)) + "\n" +
			metricRow("Avg Latency", fmt.Sprintf("%.0fms", m.stats.AvgLatencyMs)) + "\n" +
			metricRow("Helpfulness", fmt.Sprintf("%.0f%%", m.stats.HelpfulnessRate*100))

		metricsSection = "\n" + theme.SectionStyle.Render("  System Metrics") + "\n" + metricsBox.Render(metrics)
	}

	// Recent memories with styled cards
	recentSection := "\n" + theme.SectionStyle.Render("  Recent Memories") + "\n"
	if len(m.recentMemories) == 0 {
		recentSection += theme.DimStyle.Render("  No memories yet")
	} else {
		for _, mem := range m.recentMemories {
			content := mem.Content
			if len(content) > 80 {
				content = content[:77] + "..."
			}

			catBadge := theme.Badge(
				string(mem.Category),
				lipgloss.Color("#FFFFFF"),
				theme.ColorForCategory(string(mem.Category)),
			)
			date := lipgloss.NewStyle().Foreground(theme.Overlay0).Render(mem.CreatedAt.Format(time.DateOnly))
			contentLine := lipgloss.NewStyle().Foreground(theme.Text).Render(content)

			recentSection += fmt.Sprintf("  %s %s  %s\n", catBadge, date, contentLine)
		}
	}

	// Quick actions
	hintBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Surface0).
		Padding(0, 1).
		Foreground(theme.Overlay0)

	keys := []string{"[n] new memory", "[/] search", "[2] memory list", "[4] sessions", "[?] help"}
	hint := "\n" + hintBox.Render(strings.Join(keys, "  "))

	return fmt.Sprintf("\n%s\n%s%s\n%s%s", cards, chartsRow, metricsSection, recentSection, hint)
}

// RefreshCmd returns a command to reload dashboard data.
func (m DashboardModel) RefreshCmd() tea.Cmd {
	m.loading = true
	return m.loadStats()
}
