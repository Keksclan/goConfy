package screens

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/internal/tui/components"
	"github.com/keksclan/goConfy/internal/tui/logic"
	"github.com/keksclan/goConfy/internal/tui/state"
	"github.com/keksclan/goConfy/internal/tui/style"
)

// FmtView enumerates the sub-views of the format screen.
type FmtView int

const (
	FmtViewOptions FmtView = iota // toggle options
	FmtViewBefore                 // before preview
	FmtViewAfter                  // after preview
	FmtViewConfirm                // write confirmation
)

// FmtModel holds state for the format screen.
type FmtModel struct {
	View       FmtView
	VP         viewport.Model
	Before     string
	After      string
	Formatted  []byte // raw formatted bytes ready to write
	WritePath  string // resolved output path after writing
	Err        error
	Cursor     int // cursor in options list
	Ready      bool
	Width      int
	Height     int
	ConfirmYes bool // confirmation selection
}

// NewFmtModel creates a new format screen model.
func NewFmtModel() FmtModel {
	return FmtModel{Width: 80, Height: 24}
}

// fmtOptionLabels returns labels for the toggle options using current state.
func fmtOptionLabels(cfg *state.Config) []string {
	return []string{
		"Strict mode:       " + boolStr(cfg.StrictMode),
		"Expand macros:     " + boolStr(cfg.ExpandMacros),
		"Keep profiles:     " + boolStr(cfg.KeepProfiles),
		"Force overwrite:   " + boolStr(cfg.ForceOverwrite),
		"[ Preview ]",
	}
}

func boolStr(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}

// FmtUpdate handles key events for the format screen.
func FmtUpdate(m FmtModel, msg tea.Msg, cfg *state.Config) (FmtModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = false
		return m, nil
	case tea.KeyMsg:
		key := msg.String()

		switch m.View {
		case FmtViewOptions:
			labels := fmtOptionLabels(cfg)
			switch key {
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				}
			case "down", "j":
				if m.Cursor < len(labels)-1 {
					m.Cursor++
				}
			case "enter", " ":
				switch m.Cursor {
				case 0:
					cfg.StrictMode = !cfg.StrictMode
				case 1:
					cfg.ExpandMacros = !cfg.ExpandMacros
				case 2:
					cfg.KeepProfiles = !cfg.KeepProfiles
				case 3:
					cfg.ForceOverwrite = !cfg.ForceOverwrite
				case 4:
					// Preview
					m.loadPreview(cfg)
					m.View = FmtViewAfter
				}
			case "f":
				m.loadPreview(cfg)
				m.View = FmtViewAfter
			}
			return m, nil

		case FmtViewBefore, FmtViewAfter:
			switch key {
			case "tab":
				if m.View == FmtViewBefore {
					m.View = FmtViewAfter
					m.setVPContent(m.After)
				} else {
					m.View = FmtViewBefore
					m.setVPContent(m.Before)
				}
				return m, nil
			case "ctrl+s":
				m.View = FmtViewConfirm
				m.ConfirmYes = false
				return m, nil
			}

		case FmtViewConfirm:
			switch key {
			case "left", "right", "tab", "h", "l":
				m.ConfirmYes = !m.ConfirmYes
			case "enter":
				if m.ConfirmYes {
					path, err := logic.WriteFormatted(cfg, m.Formatted)
					if err != nil {
						m.Err = err
					} else {
						m.WritePath = path
					}
				}
				m.View = FmtViewAfter
				return m, nil
			case "esc":
				m.View = FmtViewAfter
				return m, nil
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.VP, cmd = m.VP.Update(msg)
	return m, cmd
}

func (m *FmtModel) loadPreview(cfg *state.Config) {
	m.Err = nil

	before, err := logic.LoadRawYAML(cfg.ConfigPath)
	if err != nil {
		m.Err = err
		return
	}
	m.Before = before

	formatted, err := logic.FormatYAML(cfg)
	if err != nil {
		m.Err = err
		return
	}
	m.Formatted = formatted
	m.After = string(formatted)

	m.setVPContent(m.After)
}

func (m *FmtModel) setVPContent(content string) {
	m.ensureVP()
	m.VP.SetContent(content)
	m.VP.GotoTop()
}

func (m *FmtModel) ensureVP() {
	if !m.Ready {
		m.VP = viewport.New(m.Width, max(m.Height-8, 10))
		m.Ready = true
	}
}

// FmtScreenView renders the format screen.
func FmtScreenView(m FmtModel, cfg *state.Config) string {
	s := style.Title.Render("Format (fmt) Config") + "\n"

	if m.Err != nil {
		s += style.StatusErr.Render("Error: "+m.Err.Error()) + "\n\n"
	}
	if m.WritePath != "" {
		s += style.StatusOK.Render("✓ Written to: "+m.WritePath) + "\n\n"
	}

	switch m.View {
	case FmtViewOptions:
		labels := fmtOptionLabels(cfg)
		s += components.RenderList(labels, m.Cursor)
		s += "\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "↑/↓", Desc: "navigate"},
			{Key: "enter", Desc: "toggle/preview"},
			{Key: "f", Desc: "format & preview"},
			{Key: "esc", Desc: "back"},
		})

	case FmtViewBefore:
		s += style.TabActive.Render("BEFORE") + style.TabInactive.Render("AFTER") + "\n"
		m.ensureVP()
		s += m.VP.View() + "\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "tab", Desc: "toggle before/after"},
			{Key: "ctrl+s", Desc: "write"},
			{Key: "esc", Desc: "back"},
		})

	case FmtViewAfter:
		s += style.TabInactive.Render("BEFORE") + style.TabActive.Render("AFTER") + "\n"
		m.ensureVP()
		s += m.VP.View() + "\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "tab", Desc: "toggle before/after"},
			{Key: "ctrl+s", Desc: "write"},
			{Key: "esc", Desc: "back"},
		})

	case FmtViewConfirm:
		target := cfg.ConfigPath
		if cfg.OutputPath != "" {
			target = cfg.OutputPath
		}
		s += components.RenderConfirm("Write formatted output to "+target+"?", m.ConfirmYes)
		s += "\n\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "←/→", Desc: "select"},
			{Key: "enter", Desc: "confirm"},
			{Key: "esc", Desc: "cancel"},
		})
	}

	return s
}
