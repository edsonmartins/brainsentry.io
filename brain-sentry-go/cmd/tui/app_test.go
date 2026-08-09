package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/integraltech/brainsentry/internal/client"
)

func TestAppGlobalNavigationAndBackStack(t *testing.T) {
	model := NewAppModel(client.New("http://127.0.0.1:1", "tenant"))
	model.activeView = ViewDashboard

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(AppModel)
	if model.activeView != ViewMemoryList || len(model.viewStack) != 1 || model.viewStack[0] != ViewDashboard {
		t.Fatalf("unexpected memory navigation: view=%v stack=%v", model.activeView, model.viewStack)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model = updated.(AppModel)
	if !model.showHelp {
		t.Fatal("help key must open help")
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(AppModel)
	if model.showHelp {
		t.Fatal("any key must close help before reaching the active view")
	}
}

func TestAppNewMemoryAndWindowSize(t *testing.T) {
	model := NewAppModel(client.New("http://127.0.0.1:1", "tenant"))
	model.activeView = ViewDashboard

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model = updated.(AppModel)
	if model.width != 100 || model.height != 40 || model.statusBar.Width != 100 {
		t.Fatalf("window size not propagated: %#v", model)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updated.(AppModel)
	if model.activeView != ViewMemoryForm {
		t.Fatalf("new shortcut did not open form: %v", model.activeView)
	}
}
