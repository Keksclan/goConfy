package screens

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/components"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/logic"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/state"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/style"
)

// ValidateModel holds state for the validate screen.
type ValidateModel struct {
	VP         viewport.Model
	Result     string // "VALID", "INVALID", or empty before run
	ErrDetail  string // full error text when invalid
	Ready      bool
	Width      int
	Height     int
	ProviderID string // registry provider to validate against
}

// NewValidateModel creates a new validate screen model.
func NewValidateModel() ValidateModel {
	return ValidateModel{Width: 80, Height: 24}
}

// RunValidation executes the validation pipeline and stores the result.
func (m *ValidateModel) RunValidation(cfg *state.Config) {
	if m.ProviderID == "" {
		ids := logic.ListRegistryIDs()
		if len(ids) > 0 {
			m.ProviderID = ids[0]
		} else {
			m.Result = "ERROR"
			m.ErrDetail = "No registry providers available. Register a config type first."
			m.syncVP()
			return
		}
	}

	err := logic.ValidateConfig(cfg, m.ProviderID)
	if err != nil {
		m.Result = "INVALID"
		m.ErrDetail = err.Error()
	} else {
		m.Result = "VALID"
		m.ErrDetail = ""
	}
	m.syncVP()
}

func (m *ValidateModel) syncVP() {
	m.ensureVP()
	if m.ErrDetail != "" {
		m.VP.SetContent(m.ErrDetail)
	} else {
		m.VP.SetContent(m.Result)
	}
	m.VP.GotoTop()
}

func (m *ValidateModel) ensureVP() {
	if !m.Ready {
		m.VP = viewport.New(m.Width, max(m.Height-8, 10))
		m.Ready = true
	}
}

// ValidateUpdate handles key events for the validate screen.
func ValidateUpdate(m ValidateModel, msg tea.Msg, cfg *state.Config) (ValidateModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "v", "r":
			m.RunValidation(cfg)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.VP, cmd = m.VP.Update(msg)
	return m, cmd
}

// ValidateView renders the validate screen.
func ValidateView(m ValidateModel) string {
	s := style.Title.Render("Validate Config") + "\n"

	switch m.Result {
	case "VALID":
		s += style.StatusOK.Render("✓ VALID") + "\n\n"
	case "INVALID":
		s += style.StatusErr.Render("✗ INVALID") + "\n\n"
	case "ERROR":
		s += style.StatusErr.Render("✗ ERROR") + "\n\n"
	default:
		s += style.Muted.Render("Press v to run validation") + "\n\n"
	}

	m.ensureVP()
	s += m.VP.View() + "\n"

	s += components.RenderHelp([]components.HelpBinding{
		{Key: "v", Desc: "validate"},
		{Key: "↑/↓", Desc: "scroll"},
		{Key: "esc", Desc: "back"},
	})
	return s
}
