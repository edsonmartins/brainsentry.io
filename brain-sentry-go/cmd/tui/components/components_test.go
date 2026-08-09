package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmKeyboardFlow(t *testing.T) {
	confirm := NewConfirm("Delete memory?", "delete")
	if !confirm.Active || !strings.Contains(confirm.View(), "Delete memory?") {
		t.Fatal("confirmation must start active and render its message")
	}

	confirm, _ = confirm.Update(tea.KeyMsg{Type: tea.KeyLeft})
	confirm, cmd := confirm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if confirm.Active || cmd == nil {
		t.Fatal("enter must finish the confirmation")
	}
	result, ok := cmd().(ConfirmResult)
	if !ok || !result.Confirmed || result.Tag != "delete" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestConfirmEscapeCancels(t *testing.T) {
	confirm, cmd := NewConfirm("Delete?", "delete").Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := cmd().(ConfirmResult)
	if confirm.Active || result.Confirmed {
		t.Fatal("escape must cancel and close")
	}
}

func TestChartsRenderAndHandleDegenerateInput(t *testing.T) {
	if HorizontalBarChart("Usage", nil, 40) != "" || Sparkline(nil, 5) != "" || VerticalBarChart(nil, nil, 4, nil) != "" {
		t.Fatal("empty charts must render no output")
	}
	bar := HorizontalBarChart("Usage", []BarChartItem{{Label: "memory", Value: 5}}, 40)
	if !strings.Contains(bar, "Usage") || !strings.Contains(bar, "5") {
		t.Fatalf("bar chart missing content: %q", bar)
	}
	if got := Sparkline([]float64{1, 2, 3, 4}, 2); len([]rune(got)) < 2 {
		t.Fatalf("sparkline was not rendered: %q", got)
	}
	vertical := VerticalBarChart([]string{"long-label"}, []int64{0}, 2, nil)
	if !strings.Contains(vertical, "long") {
		t.Fatalf("vertical chart label was not rendered: %q", vertical)
	}
}

func TestStatusBarAndToastLifecycle(t *testing.T) {
	bar := StatusBar{Width: 60, View: "Memories", User: "demo", Tenant: "tenant", HintKeys: "q quit"}.Render()
	for _, expected := range []string{"Memories", "demo", "tenant", "q quit"} {
		if !strings.Contains(bar, expected) {
			t.Fatalf("status bar missing %q: %q", expected, bar)
		}
	}

	var toast Toast
	if cmd := toast.Show("Saved", ToastSuccess); cmd == nil || !toast.Visible || !strings.Contains(toast.View(), "Saved") {
		t.Fatal("toast must become visible and schedule dismissal")
	}
	toast.Dismiss()
	if toast.Visible || toast.View() != "" {
		t.Fatal("dismissed toast must be hidden")
	}
}
