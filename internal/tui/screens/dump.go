package screens

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/internal/tui/components"
	"github.com/keksclan/goConfy/internal/tui/logic"
	"github.com/keksclan/goConfy/internal/tui/state"
	"github.com/keksclan/goConfy/internal/tui/style"
)

// DumpModel holds state for the dump redacted JSON screen.
type DumpModel struct {
	VP         viewport.Model
	Content    string
	Err        error
	Ready      bool
	Width      int
	Height     int
	ProviderID string
}

// NewDumpModel creates a new dump screen model.
func NewDumpModel() DumpModel {
	return DumpModel{Width: 80, Height: 24}
}

// RunDump executes the dump pipeline and stores the redacted JSON output.
func (m *DumpModel) RunDump(cfg *state.Config) {
	m.Err = nil

	if m.ProviderID == "" {
		ids := logic.ListRegistryIDs()
		if len(ids) > 0 {
			m.ProviderID = ids[0]
		} else {
			// Fallback: show merged YAML redacted when no provider is available.
			content, err := logic.MergedYAML(cfg)
			if err != nil {
				m.Err = err
				return
			}
			content, err = logic.RedactYAML(content, cfg.RedactionPaths)
			if err != nil {
				m.Err = err
				return
			}
			m.Content = "⚠ No registry provider – showing redacted YAML instead.\n\n" + content
			m.syncVP()
			return
		}
	}

	content, err := logic.DumpRedactedJSON(cfg, m.ProviderID)
	if err != nil {
		m.Err = err
		return
	}
	m.Content = content
	m.syncVP()
}

func (m *DumpModel) syncVP() {
	m.ensureVP()
	m.VP.SetContent(m.Content)
	m.VP.GotoTop()
}

func (m *DumpModel) ensureVP() {
	if !m.Ready {
		m.VP = viewport.New(m.Width, max(m.Height-8, 10))
		m.Ready = true
	}
}

// DumpUpdate handles key events for the dump screen.
func DumpUpdate(m DumpModel, msg tea.Msg, cfg *state.Config) (DumpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "d", "r":
			m.RunDump(cfg)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.VP, cmd = m.VP.Update(msg)
	return m, cmd
}

// DumpView renders the dump screen.
func DumpView(m DumpModel) string {
	s := style.Title.Render("Dump Redacted Config (JSON)") + "\n"

	if m.Err != nil {
		s += style.StatusErr.Render("Error: "+m.Err.Error()) + "\n\n"
	}

	if m.Content == "" {
		s += style.Muted.Render("Press d to dump redacted JSON") + "\n\n"
	}

	m.ensureVP()
	s += m.VP.View() + "\n"

	s += components.RenderHelp([]components.HelpBinding{
		{Key: "d", Desc: "dump"},
		{Key: "↑/↓", Desc: "scroll"},
		{Key: "esc", Desc: "back"},
	})
	return s
}
