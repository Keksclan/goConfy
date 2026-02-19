package screens

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/internal/tui/components"
	"github.com/keksclan/goConfy/internal/tui/logic"
	"github.com/keksclan/goConfy/internal/tui/state"
	"github.com/keksclan/goConfy/internal/tui/style"
)

// InspectTab identifies which preview tab is active.
type InspectTab int

const (
	TabRawYAML  InspectTab = iota // raw file contents
	TabExpanded                   // after macro expansion
	TabMerged                     // after profile merge
	TabRedacted                   // redacted JSON (always safe)
)

var inspectTabLabels = []string{"RAW YAML", "EXPANDED", "MERGED", "REDACTED JSON"}

// InspectModel holds state for the inspect / preview screen.
type InspectModel struct {
	Tab      InspectTab
	VP       viewport.Model
	Content  string // current tab content
	Err      error
	Profiles []string
	HasProf  bool
	Ready    bool
	Width    int
	Height   int
}

// NewInspectModel creates a new inspect screen model.
func NewInspectModel() InspectModel {
	return InspectModel{Width: 80, Height: 24}
}

// LoadContent populates the viewport with the content for the current tab.
func (m *InspectModel) LoadContent(cfg *state.Config) {
	m.Err = nil
	var content string
	var err error

	switch m.Tab {
	case TabRawYAML:
		content, err = logic.LoadRawYAML(cfg.ConfigPath)
		if err == nil {
			// Best-effort YAML redaction.
			content, err = logic.RedactYAML(content, cfg.RedactionPaths)
		}
	case TabExpanded:
		content, err = logic.ExpandedYAML(cfg)
		if err == nil {
			content, err = logic.RedactYAML(content, cfg.RedactionPaths)
		}
	case TabMerged:
		content, err = logic.MergedYAML(cfg)
		if err == nil {
			content, err = logic.RedactYAML(content, cfg.RedactionPaths)
		}
	case TabRedacted:
		// Uses the first available registry provider.
		ids := logic.ListRegistryIDs()
		if len(ids) > 0 {
			content, err = logic.DumpRedactedJSON(cfg, ids[0])
		} else {
			// Fallback: show merged YAML redacted.
			content, err = logic.MergedYAML(cfg)
			if err == nil {
				content, err = logic.RedactYAML(content, cfg.RedactionPaths)
				if err == nil {
					content = "⚠ No registry provider – showing redacted YAML instead.\n\n" + content
				}
			}
		}
	}

	if err != nil {
		m.Err = err
		m.Content = ""
	} else {
		m.Content = content
	}

	// Load profile info.
	if cfg.ConfigPath != "" {
		m.Profiles, m.HasProf, _ = logic.ListProfiles(cfg.ConfigPath)
	}

	m.ensureVP()
	m.VP.SetContent(m.Content)
	m.VP.GotoTop()
}

func (m *InspectModel) ensureVP() {
	if !m.Ready {
		m.VP = viewport.New(m.Width, max(m.Height-8, 10))
		m.Ready = true
	}
}

// InspectUpdate handles key events for the inspect screen.
func InspectUpdate(m InspectModel, msg tea.Msg, cfg *state.Config) (InspectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = false
		m.LoadContent(cfg)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.Tab = InspectTab((int(m.Tab) + 1) % len(inspectTabLabels))
			m.LoadContent(cfg)
			return m, nil
		case "shift+tab":
			m.Tab = InspectTab((int(m.Tab) - 1 + len(inspectTabLabels)) % len(inspectTabLabels))
			m.LoadContent(cfg)
			return m, nil
		case "r":
			m.LoadContent(cfg)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.VP, cmd = m.VP.Update(msg)
	return m, cmd
}

// InspectView renders the inspect screen.
func InspectView(m InspectModel, cfg *state.Config) string {
	var s string
	s += style.Title.Render("Inspect / Preview") + "\n"

	// Tabs
	for i, lbl := range inspectTabLabels {
		if InspectTab(i) == m.Tab {
			s += style.TabActive.Render(lbl)
		} else {
			s += style.TabInactive.Render(lbl)
		}
	}
	s += "\n"

	// Profile info
	profileLine := "Profile: "
	active := cfg.ActiveProfile
	if active == "" {
		active = "(from " + cfg.ProfileEnvVar + ")"
	}
	profileLine += active
	if m.HasProf {
		profileLine += "  |  Available: " + joinStrings(m.Profiles)
	} else {
		profileLine += "  |  No profiles section found"
	}
	s += style.Muted.Render(profileLine) + "\n"

	// Warning about YAML redaction.
	if m.Tab != TabRedacted {
		s += style.Warning.Render("⚠ YAML redaction is best-effort. Use REDACTED JSON tab for guaranteed safety.") + "\n"
	}

	if m.Err != nil {
		s += style.StatusErr.Render("Error: "+m.Err.Error()) + "\n"
	} else {
		m.ensureVP()
		s += m.VP.View() + "\n"
	}

	s += components.RenderHelp([]components.HelpBinding{
		{Key: "tab", Desc: "switch tab"},
		{Key: "r", Desc: "reload"},
		{Key: "↑/↓", Desc: "scroll"},
		{Key: "esc", Desc: "back"},
	})
	return s
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += ", " + s
	}
	return result
}
