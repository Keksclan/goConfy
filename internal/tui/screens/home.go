// Package screens implements individual TUI screens. Each screen exposes an
// Update function (handling key messages) and a View function (rendering).
package screens

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/internal/tui/components"
	"github.com/keksclan/goConfy/internal/tui/style"
)

// HomeItem enumerates the main menu entries.
type HomeItem int

const (
	HomeInspect  HomeItem = iota // Open / Inspect config
	HomeInit                     // Generate config template
	HomeValidate                 // Validate config
	HomeFmt                      // Format config
	HomeDump                     // Dump redacted config
	HomeSettings                 // Settings
	HomeQuit                     // Quit
)

// homeLabels maps each menu item to its display label.
var homeLabels = []string{
	"1) Open / Inspect config",
	"2) Generate config template (init)",
	"3) Validate config",
	"4) Format (fmt) config",
	"5) Dump redacted config",
	"6) Settings",
	"0) Quit",
}

// HomeModel holds the state for the home / workflow menu screen.
type HomeModel struct {
	Cursor int
}

// NewHomeModel creates a HomeModel with the cursor on the first entry.
func NewHomeModel() HomeModel {
	return HomeModel{}
}

// SelectedItem returns the currently highlighted menu item.
func (m HomeModel) SelectedItem() HomeItem {
	return HomeItem(m.Cursor)
}

// HomeUpdate processes key events on the home screen. It returns the updated
// model plus an optional NavigateMsg when the user presses enter.
func HomeUpdate(m HomeModel, msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(homeLabels)-1 {
				m.Cursor++
			}
		case "enter":
			return m, nil // caller inspects SelectedItem()
		case "1":
			m.Cursor = int(HomeInspect)
			return m, nil
		case "2":
			m.Cursor = int(HomeInit)
			return m, nil
		case "3":
			m.Cursor = int(HomeValidate)
			return m, nil
		case "4":
			m.Cursor = int(HomeFmt)
			return m, nil
		case "5":
			m.Cursor = int(HomeDump)
			return m, nil
		case "6":
			m.Cursor = int(HomeSettings)
			return m, nil
		case "0":
			m.Cursor = int(HomeQuit)
			return m, nil
		}
	}
	return m, nil
}

// HomeView renders the home screen.
func HomeView(m HomeModel) string {
	s := style.Title.Render("Workflow Menu") + "\n\n"
	s += components.RenderList(homeLabels, m.Cursor)
	s += "\n"
	s += components.RenderHelp([]components.HelpBinding{
		{Key: "↑/↓", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "q", Desc: "quit"},
		{Key: "?", Desc: "help"},
	})
	return s
}
