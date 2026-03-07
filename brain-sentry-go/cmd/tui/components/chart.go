package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/integraltech/brainsentry/cmd/tui/theme"
)

// BarChartItem represents a single bar in a horizontal bar chart.
type BarChartItem struct {
	Label string
	Value int64
	Color string
}

// HorizontalBarChart renders a horizontal bar chart.
func HorizontalBarChart(title string, items []BarChartItem, maxWidth int) string {
	if len(items) == 0 {
		return ""
	}

	// Find max value
	var maxVal int64
	for _, item := range items {
		if item.Value > maxVal {
			maxVal = item.Value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	barWidth := maxWidth - 26
	if barWidth < 10 {
		barWidth = 10
	}

	barLabelStyle := lipgloss.NewStyle().
		Foreground(theme.Subtext0).
		Width(16).
		Align(lipgloss.Right)

	barValueStyle := lipgloss.NewStyle().
		Foreground(theme.Text).
		Width(6)

	var b strings.Builder
	b.WriteString(theme.SectionStyle.Render(title) + "\n")

	for _, item := range items {
		label := barLabelStyle.Render(item.Label)
		value := barValueStyle.Render(fmt.Sprintf("%d", item.Value))

		width := int(float64(barWidth) * float64(item.Value) / float64(maxVal))
		if width < 1 && item.Value > 0 {
			width = 1
		}

		color := item.Color
		if color == "" {
			color = string(theme.Primary)
		}
		barStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(color))

		bar := barStyle.Render(strings.Repeat("█", width))
		b.WriteString(fmt.Sprintf("%s %s %s\n", label, bar, value))
	}

	return b.String()
}

// Sparkline renders a sparkline chart from data points.
func Sparkline(data []float64, width int) string {
	if len(data) == 0 {
		return ""
	}

	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	valRange := maxVal - minVal
	if valRange == 0 {
		valRange = 1
	}

	samples := data
	if len(data) > width {
		samples = resample(data, width)
	}

	var b strings.Builder
	for _, v := range samples {
		idx := int(math.Round((v - minVal) / valRange * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		b.WriteRune(blocks[idx])
	}

	return lipgloss.NewStyle().
		Foreground(theme.Primary).
		Render(b.String())
}

// VerticalBarChart renders a vertical column chart.
func VerticalBarChart(labels []string, values []int64, height int, colors []string) string {
	if len(values) == 0 {
		return ""
	}

	var maxVal int64
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	var rows []string
	for row := height; row > 0; row-- {
		var cols []string
		threshold := float64(row) / float64(height)
		for i, v := range values {
			ratio := float64(v) / float64(maxVal)
			color := string(theme.Primary)
			if i < len(colors) && colors[i] != "" {
				color = colors[i]
			}
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			if ratio >= threshold {
				cols = append(cols, style.Render(" ██ "))
			} else {
				cols = append(cols, "    ")
			}
		}
		rows = append(rows, strings.Join(cols, ""))
	}

	sep := strings.Repeat("────", len(values))

	var labelParts []string
	for _, l := range labels {
		if len(l) > 4 {
			l = l[:4]
		}
		labelParts = append(labelParts, fmt.Sprintf("%-4s", l))
	}
	labelRow := lipgloss.NewStyle().
		Foreground(theme.Overlay0).
		Render(strings.Join(labelParts, ""))

	return strings.Join(rows, "\n") + "\n" + sep + "\n" + labelRow
}

func resample(data []float64, targetLen int) []float64 {
	if targetLen >= len(data) {
		return data
	}
	result := make([]float64, targetLen)
	ratio := float64(len(data)) / float64(targetLen)
	for i := 0; i < targetLen; i++ {
		idx := int(float64(i) * ratio)
		if idx >= len(data) {
			idx = len(data) - 1
		}
		result[i] = data[idx]
	}
	return result
}
