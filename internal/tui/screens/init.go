package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/internal/tui/components"
	"github.com/keksclan/goConfy/internal/tui/logic"
	"github.com/keksclan/goConfy/internal/tui/state"
	"github.com/keksclan/goConfy/internal/tui/style"
)

// InitPhase enumerates the sub-phases of the init screen.
type InitPhase int

const (
	InitPhaseSelect  InitPhase = iota // select registry ID
	InitPhaseOptions                  // set toggles + output path
	InitPhasePreview                  // preview generated YAML
	InitPhaseConfirm                  // confirm write
)

// InitModel holds state for the init / template generation screen.
type InitModel struct {
	Phase       InitPhase
	IDs         []string // available registry IDs
	IDCursor    int
	OutputInput textinput.Model
	VP          viewport.Model
	YAMLContent string
	EnvContent  string
	Written     []string // paths written to disk
	Err         error
	Cursor      int // cursor in options list
	Ready       bool
	Width       int
	Height      int
	ConfirmYes  bool
}

// NewInitModel creates a new init screen model.
func NewInitModel() InitModel {
	ti := components.NewTextInput("./config.yml")
	ti.Focus()
	return InitModel{
		IDs:         logic.ListRegistryIDs(),
		OutputInput: ti,
		Width:       80,
		Height:      24,
	}
}

// InitUpdate handles key events for the init screen.
func InitUpdate(m InitModel, msg tea.Msg, cfg *state.Config) (InitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = false
		return m, nil
	case tea.KeyMsg:
		key := msg.String()

		switch m.Phase {
		case InitPhaseSelect:
			switch key {
			case "up", "k":
				if m.IDCursor > 0 {
					m.IDCursor--
				}
			case "down", "j":
				if m.IDCursor < len(m.IDs)-1 {
					m.IDCursor++
				}
			case "enter":
				if len(m.IDs) > 0 {
					cfg.RegistryID = m.IDs[m.IDCursor]
					m.Phase = InitPhaseOptions
				}
			}
			return m, nil

		case InitPhaseOptions:
			switch key {
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				}
			case "down", "j":
				if m.Cursor < 4 {
					m.Cursor++
				}
			case "enter", " ":
				switch m.Cursor {
				case 0:
					cfg.GenerateDotEnv = !cfg.GenerateDotEnv
				case 1:
					cfg.IncludeProfiles = !cfg.IncludeProfiles
				case 2:
					cfg.ForceOverwrite = !cfg.ForceOverwrite
				case 3:
					// Output path input – focus it.
					m.Phase = InitPhaseOptions // stays, but we handle typing below.
				case 4:
					// Generate preview.
					m.generate(cfg)
				}
			case "i":
				m.generate(cfg)
			}
			// Delegate to output input when cursor is on it.
			if m.Cursor == 3 {
				var cmd tea.Cmd
				m.OutputInput, cmd = m.OutputInput.Update(msg)
				cfg.OutputPath = m.OutputInput.Value()
				return m, cmd
			}
			return m, nil

		case InitPhasePreview:
			switch key {
			case "ctrl+s":
				m.Phase = InitPhaseConfirm
				m.ConfirmYes = false
				return m, nil
			}

		case InitPhaseConfirm:
			switch key {
			case "left", "right", "tab", "h", "l":
				m.ConfirmYes = !m.ConfirmYes
			case "enter":
				if m.ConfirmYes {
					written, err := logic.WriteTemplate(cfg, m.YAMLContent, m.EnvContent)
					if err != nil {
						m.Err = err
					} else {
						m.Written = written
					}
				}
				m.Phase = InitPhasePreview
				return m, nil
			case "esc":
				m.Phase = InitPhasePreview
				return m, nil
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.VP, cmd = m.VP.Update(msg)
	return m, cmd
}

func (m *InitModel) generate(cfg *state.Config) {
	m.Err = nil
	cfg.OutputPath = m.OutputInput.Value()

	yamlContent, envContent, err := logic.GenerateTemplate(cfg)
	if err != nil {
		m.Err = err
		return
	}
	m.YAMLContent = yamlContent
	m.EnvContent = envContent

	preview := yamlContent
	if envContent != "" {
		preview += "\n--- .env ---\n" + envContent
	}

	m.ensureVP()
	m.VP.SetContent(preview)
	m.VP.GotoTop()
	m.Phase = InitPhasePreview
}

func (m *InitModel) ensureVP() {
	if !m.Ready {
		m.VP = viewport.New(m.Width, max(m.Height-8, 10))
		m.Ready = true
	}
}

// InitScreenView renders the init screen.
func InitScreenView(m InitModel, cfg *state.Config) string {
	s := style.Title.Render("Generate Config Template (init)") + "\n"

	if m.Err != nil {
		s += style.StatusErr.Render("Error: "+m.Err.Error()) + "\n\n"
	}
	if len(m.Written) > 0 {
		s += style.StatusOK.Render("✓ Written: "+strings.Join(m.Written, ", ")) + "\n\n"
	}

	switch m.Phase {
	case InitPhaseSelect:
		if len(m.IDs) == 0 {
			s += style.Warning.Render("No registry providers available. Register config types first.") + "\n"
		} else {
			s += style.Muted.Render("Select a config type:") + "\n\n"
			s += components.RenderList(m.IDs, m.IDCursor)
		}
		s += "\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "↑/↓", Desc: "navigate"},
			{Key: "enter", Desc: "select"},
			{Key: "esc", Desc: "back"},
		})

	case InitPhaseOptions:
		s += style.Muted.Render(fmt.Sprintf("Registry ID: %s", cfg.RegistryID)) + "\n\n"
		labels := []string{
			"Generate .env:     " + boolStr(cfg.GenerateDotEnv),
			"Include profiles:  " + boolStr(cfg.IncludeProfiles),
			"Force overwrite:   " + boolStr(cfg.ForceOverwrite),
			"Output path:       " + m.OutputInput.View(),
			"[ Generate ]",
		}
		s += components.RenderList(labels, m.Cursor)
		s += "\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "↑/↓", Desc: "navigate"},
			{Key: "enter", Desc: "toggle/generate"},
			{Key: "i", Desc: "generate"},
			{Key: "esc", Desc: "back"},
		})

	case InitPhasePreview:
		s += "Generated YAML preview:\n"
		m.ensureVP()
		s += m.VP.View() + "\n"

		// Summary of what will be written.
		outPath := cfg.OutputPath
		if outPath == "" {
			outPath = "./config.yml"
		}
		s += style.Warning.Render("Will write: "+outPath) + "\n"
		if cfg.GenerateDotEnv {
			s += style.Warning.Render("Will write: .env alongside config") + "\n"
		}

		s += components.RenderHelp([]components.HelpBinding{
			{Key: "ctrl+s", Desc: "save to disk"},
			{Key: "↑/↓", Desc: "scroll"},
			{Key: "esc", Desc: "back"},
		})

	case InitPhaseConfirm:
		s += components.RenderConfirm("Write generated files to disk?", m.ConfirmYes)
		s += "\n\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "←/→", Desc: "select"},
			{Key: "enter", Desc: "confirm"},
			{Key: "esc", Desc: "cancel"},
		})
	}

	return s
}
