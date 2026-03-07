package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/integraltech/brainsentry/cmd/tui/components"
	tuikeys "github.com/integraltech/brainsentry/cmd/tui/keys"
	"github.com/integraltech/brainsentry/cmd/tui/views"
	"github.com/integraltech/brainsentry/internal/client"
)

// ViewID identifies which view is active.
type ViewID int

const (
	ViewLogin ViewID = iota
	ViewDashboard
	ViewMemoryList
	ViewMemoryDetail
	ViewMemoryForm
	ViewSearch
	ViewSessions
	ViewRelationships
)

// AppModel is the root Bubble Tea model that routes between views.
type AppModel struct {
	client *client.Client
	keys   tuikeys.KeyMap

	activeView ViewID
	viewStack  []ViewID // navigation history for back

	// Sub-models
	login         views.LoginModel
	dashboard     views.DashboardModel
	memoryList    views.MemoryListModel
	memoryDetail  views.MemoryDetailModel
	memoryForm    views.MemoryFormModel
	search        views.SearchModel
	sessions      views.SessionsModel
	relationships views.RelationshipsModel

	// Components
	statusBar components.StatusBar
	toast     components.Toast

	// State
	userName   string
	tenantName string
	showHelp   bool
	width      int
	height     int
}

// NewAppModel creates the root app model.
func NewAppModel(c *client.Client) AppModel {
	return AppModel{
		client:     c,
		keys:       tuikeys.DefaultKeyMap(),
		activeView: ViewLogin,
		login:      views.NewLoginModel(c),
	}
}

// Init initializes the app.
func (m AppModel) Init() tea.Cmd {
	return m.login.Init()
}

// pushView pushes current view onto the stack and switches to the new view.
func (m *AppModel) pushView(view ViewID) tea.Cmd {
	m.viewStack = append(m.viewStack, m.activeView)
	m.activeView = view
	return m.initView(view)
}

// popView returns to the previous view in the stack.
func (m *AppModel) popView() tea.Cmd {
	if len(m.viewStack) == 0 {
		return nil
	}
	prev := m.viewStack[len(m.viewStack)-1]
	m.viewStack = m.viewStack[:len(m.viewStack)-1]
	m.activeView = prev
	return m.initView(prev)
}

// navigateTo switches to a top-level view (clears stack, sets dashboard as base).
func (m *AppModel) navigateTo(view ViewID) tea.Cmd {
	if view == ViewDashboard {
		m.viewStack = nil
	} else {
		m.viewStack = []ViewID{ViewDashboard}
	}
	m.activeView = view
	return m.initView(view)
}

// initView initializes a view's model and returns its Init command.
func (m *AppModel) initView(view ViewID) tea.Cmd {
	switch view {
	case ViewDashboard:
		m.dashboard = views.NewDashboardModel(m.client)
		return m.dashboard.Init()
	case ViewMemoryList:
		m.memoryList = views.NewMemoryListModel(m.client)
		return m.memoryList.Init()
	case ViewSearch:
		m.search = views.NewSearchModel(m.client)
		return m.search.Init()
	case ViewSessions:
		m.sessions = views.NewSessionsModel(m.client)
		return m.sessions.Init()
	}
	return nil
}

// isInputView returns true if the current view has text input that needs key events.
func (m AppModel) isInputView() bool {
	return m.activeView == ViewLogin || m.activeView == ViewMemoryForm
}

// Update routes messages to the active view.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.Width = msg.Width

	case tea.KeyMsg:
		// Global quit (not in input views or search)
		if key.Matches(msg, m.keys.Quit) && !m.isInputView() && m.activeView != ViewSearch {
			return m, tea.Quit
		}

		// Toggle help (not in input views)
		if key.Matches(msg, m.keys.Help) && !m.isInputView() {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// If help is shown, any key hides it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// Global "/" to open search (from dashboard, sessions — not from views that use "/" internally)
		if key.Matches(msg, m.keys.Search) && !m.isInputView() &&
			m.activeView != ViewSearch && m.activeView != ViewMemoryList {
			return m, m.navigateTo(ViewSearch)
		}

		// All views now handle ESC internally via NavigateBackMsg
		// No global ESC interception needed

		// Global number keys for view navigation (not in input views)
		if !m.isInputView() {
			switch {
			case key.Matches(msg, m.keys.ViewDashboard):
				return m, m.navigateTo(ViewDashboard)
			case key.Matches(msg, m.keys.ViewMemories):
				return m, m.navigateTo(ViewMemoryList)
			case key.Matches(msg, m.keys.ViewSearch):
				return m, m.navigateTo(ViewSearch)
			case key.Matches(msg, m.keys.ViewSessions):
				return m, m.navigateTo(ViewSessions)
			}
		}

	// Login success -> switch to dashboard
	case views.LoginSuccessMsg:
		m.userName = msg.Response.User.Email
		m.tenantName = msg.Response.TenantID
		m.statusBar.User = m.userName
		m.statusBar.Tenant = m.tenantName
		return m, m.navigateTo(ViewDashboard)

	// Navigation messages from sub-views
	case views.MemorySelectedMsg:
		m.memoryDetail = views.NewMemoryDetailModel(m.client, msg.ID)
		cmd := m.pushView(ViewMemoryDetail)
		return m, tea.Batch(cmd, m.memoryDetail.Init())

	case views.NavigateBackMsg:
		return m, m.popView()

	case views.EditMemoryMsg:
		m.memoryForm = views.NewMemoryFormModel(m.client, msg.ID)
		m.pushView(ViewMemoryForm)
		return m, m.memoryForm.Init()

	case views.ViewRelationshipsMsg:
		m.relationships = views.NewRelationshipsModel(m.client, msg.ID)
		cmd := m.pushView(ViewRelationships)
		return m, tea.Batch(cmd, m.relationships.Init())

	case views.MemorySavedMsg:
		// After save, go back to memory list
		m.viewStack = []ViewID{ViewDashboard}
		m.activeView = ViewMemoryList
		m.memoryList = views.NewMemoryListModel(m.client)
		cmd := m.memoryList.Init()
		toastCmd := m.toast.Show("Memory saved!", components.ToastSuccess)
		return m, tea.Batch(cmd, toastCmd)

	case components.ToastDismissMsg:
		m.toast.Dismiss()
		return m, nil
	}

	// Handle "n" for new memory (from list or dashboard)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.New) && (m.activeView == ViewMemoryList || m.activeView == ViewDashboard) {
			m.memoryForm = views.NewMemoryFormModel(m.client, "")
			m.pushView(ViewMemoryForm)
			return m, m.memoryForm.Init()
		}
	}

	// Delegate to active view
	var cmd tea.Cmd
	switch m.activeView {
	case ViewLogin:
		m.login, cmd = m.login.Update(msg)
	case ViewDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case ViewMemoryList:
		m.memoryList, cmd = m.memoryList.Update(msg)
	case ViewMemoryDetail:
		m.memoryDetail, cmd = m.memoryDetail.Update(msg)
	case ViewMemoryForm:
		m.memoryForm, cmd = m.memoryForm.Update(msg)
	case ViewSearch:
		m.search, cmd = m.search.Update(msg)
	case ViewSessions:
		m.sessions, cmd = m.sessions.Update(msg)
	case ViewRelationships:
		m.relationships, cmd = m.relationships.Update(msg)
	}

	return m, cmd
}

// View renders the active view with status bar.
func (m AppModel) View() string {
	if m.showHelp {
		return views.HelpView(m.width, m.height)
	}

	var content string
	switch m.activeView {
	case ViewLogin:
		return m.login.View()
	case ViewDashboard:
		m.statusBar.View = "Dashboard"
		m.statusBar.HintKeys = "[/] search  [?] help  [q] quit"
		content = m.dashboard.View()
	case ViewMemoryList:
		m.statusBar.View = "Memories"
		m.statusBar.HintKeys = "[esc] back  [?] help  [q] quit"
		content = m.memoryList.View()
	case ViewMemoryDetail:
		m.statusBar.View = "Memory Detail"
		m.statusBar.HintKeys = "[esc] back  [?] help"
		content = m.memoryDetail.View()
	case ViewMemoryForm:
		m.statusBar.View = "Memory Form"
		m.statusBar.HintKeys = "[Ctrl+S] save  [Esc] cancel"
		content = m.memoryForm.View()
	case ViewSearch:
		m.statusBar.View = "Search"
		m.statusBar.HintKeys = "[esc] back  [?] help"
		content = m.search.View()
	case ViewSessions:
		m.statusBar.View = "Sessions"
		m.statusBar.HintKeys = "[esc] back  [?] help  [q] quit"
		content = m.sessions.View()
	case ViewRelationships:
		m.statusBar.View = "Relationships"
		m.statusBar.HintKeys = "[esc] back  [?] help"
		content = m.relationships.View()
	}

	// Toast overlay
	toastStr := m.toast.View()
	if toastStr != "" {
		content = toastStr + "\n" + content
	}

	// Status bar at bottom
	statusBar := m.statusBar.Render()

	return content + "\n" + statusBar
}
