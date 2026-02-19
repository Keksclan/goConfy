package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/internal/tui/components"
	"github.com/keksclan/goConfy/internal/tui/state"
	"github.com/keksclan/goConfy/internal/tui/style"
)

// SettingsField enumerates editable settings fields.
type SettingsField int

const (
	SettFieldStrict SettingsField = iota
	SettFieldDotEnvOpt
	SettFieldOSWins
	SettFieldProfileVar
	SettFieldRedactPaths
)

const settingsCount = 5

// SettingsModel holds state for the settings screen.
type SettingsModel struct {
	Cursor        int
	ProfileVarIn  textinput.Model
	RedactPathsIn textinput.Model
	EditingText   bool // true when a text input is focused for editing
}

// NewSettingsModel creates a new settings screen model.
func NewSettingsModel(cfg *state.Config) SettingsModel {
	pv := components.NewTextInput("APP_PROFILE")
	pv.SetValue(cfg.ProfileEnvVar)

	rp := components.NewTextInput("redis.password,postgres.url")
	rp.SetValue(strings.Join(cfg.RedactionPaths, ","))

	return SettingsModel{
		ProfileVarIn:  pv,
		RedactPathsIn: rp,
	}
}

// SettingsUpdate handles key events for the settings screen.
func SettingsUpdate(m SettingsModel, msg tea.Msg, cfg *state.Config) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// When editing a text input, delegate until enter/esc.
		if m.EditingText {
			switch key {
			case "enter", "esc":
				m.EditingText = false
				m.ProfileVarIn.Blur()
				m.RedactPathsIn.Blur()
				// Apply values back to config.
				cfg.ProfileEnvVar = m.ProfileVarIn.Value()
				cfg.RedactionPaths = splitPaths(m.RedactPathsIn.Value())
				return m, nil
			}
			var cmd tea.Cmd
			if SettingsField(m.Cursor) == SettFieldProfileVar {
				m.ProfileVarIn, cmd = m.ProfileVarIn.Update(msg)
			} else {
				m.RedactPathsIn, cmd = m.RedactPathsIn.Update(msg)
			}
			return m, cmd
		}

		switch key {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < settingsCount-1 {
				m.Cursor++
			}
		case "enter", " ":
			switch SettingsField(m.Cursor) {
			case SettFieldStrict:
				cfg.StrictMode = !cfg.StrictMode
			case SettFieldDotEnvOpt:
				cfg.DotEnvOptional = !cfg.DotEnvOptional
			case SettFieldOSWins:
				cfg.DotEnvOSWins = !cfg.DotEnvOSWins
			case SettFieldProfileVar:
				m.EditingText = true
				m.ProfileVarIn.Focus()
			case SettFieldRedactPaths:
				m.EditingText = true
				m.RedactPathsIn.Focus()
			}
		}
	}
	return m, nil
}

// splitPaths splits a comma-separated string into trimmed path segments.
func splitPaths(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// SettingsView renders the settings screen.
func SettingsView(m SettingsModel, cfg *state.Config) string {
	s := style.Title.Render("Settings") + "\n\n"

	labels := []string{
		"Strict mode:           " + boolStr(cfg.StrictMode),
		"Dotenv optional:       " + boolStr(cfg.DotEnvOptional),
		"OS env wins:           " + boolStr(cfg.DotEnvOSWins),
		"Profile env var:       " + m.ProfileVarIn.View(),
		"Redaction dot-paths:   " + m.RedactPathsIn.View(),
	}

	s += components.RenderList(labels, m.Cursor)
	s += "\n"

	if m.EditingText {
		s += style.Muted.Render("Type to edit, press enter to confirm") + "\n"
	}

	s += "\n"
	s += style.Muted.Render("Settings are stored in-memory for this session.") + "\n"
	s += components.RenderHelp([]components.HelpBinding{
		{Key: "↑/↓", Desc: "navigate"},
		{Key: "enter", Desc: "toggle/edit"},
		{Key: "esc", Desc: "back"},
	})
	return s
}
