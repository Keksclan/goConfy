package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/components"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/screens"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/state"
)

// Model is the top-level bubbletea model that holds all screen models and
// routes key events to the active screen.
type Model struct {
	Cfg *state.Config

	// Current screen
	Screen Screen
	// Target screen after file picker (which workflow triggered it).
	NextScreen Screen

	// Screen models
	Home       screens.HomeModel
	FilePicker screens.FilePickerModel
	Inspect    screens.InspectModel
	Validate   screens.ValidateModel
	Fmt        screens.FmtModel
	Dump       screens.DumpModel
	InitScr    screens.InitModel
	Settings   screens.SettingsModel

	// Global state
	ShowHelp bool
	Width    int
	Height   int
}

// NewModel creates the initial application model with default state.
func NewModel() Model {
	cfg := state.DefaultConfig()
	return Model{
		Cfg:        cfg,
		Screen:     ScreenHome,
		Home:       screens.NewHomeModel(),
		FilePicker: screens.NewFilePickerModel(),
		Inspect:    screens.NewInspectModel(),
		Validate:   screens.NewValidateModel(),
		Fmt:        screens.NewFmtModel(),
		Dump:       screens.NewDumpModel(),
		InitScr:    screens.NewInitModel(),
		Settings:   screens.NewSettingsModel(cfg),
	}
}

// Init implements tea.Model. No initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. It dispatches messages to the active screen and
// handles global keys (quit, help toggle, shortcut navigation).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Propagate to all screens that care about size.
		m.Inspect, _ = screens.InspectUpdate(m.Inspect, msg, m.Cfg)
		m.Validate, _ = screens.ValidateUpdate(m.Validate, msg, m.Cfg)
		m.Fmt, _ = screens.FmtUpdate(m.Fmt, msg, m.Cfg)
		m.Dump, _ = screens.DumpUpdate(m.Dump, msg, m.Cfg)
		m.InitScr, _ = screens.InitUpdate(m.InitScr, msg, m.Cfg)
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		// Global quit from home screen.
		if m.Screen == ScreenHome && (key == "q" || key == "ctrl+c") {
			return m, tea.Quit
		}
		// ctrl+c always quits.
		if key == "ctrl+c" {
			return m, tea.Quit
		}

		// Global help toggle.
		if key == "?" {
			m.ShowHelp = !m.ShowHelp
			return m, nil
		}

		// Esc / q goes back (except when in text editing or browse mode).
		if key == "esc" || key == "q" {
			if m.Screen != ScreenHome {
				// Special: if file picker is in browse mode, let it handle esc.
				if m.Screen == ScreenFilePicker && m.FilePicker.BrowseMode {
					break // fall through to screen handler
				}
				// Special: if settings is editing text, let it handle esc.
				if m.Screen == ScreenSettings && m.Settings.EditingText {
					break
				}
				// Special: if fmt is in confirm, let it handle esc.
				if m.Screen == ScreenFmt && m.Fmt.View == screens.FmtViewConfirm {
					break
				}
				// Special: if init is in confirm, let it handle esc.
				if m.Screen == ScreenInit && m.InitScr.Phase == screens.InitPhaseConfirm {
					break
				}
				m.Screen = ScreenHome
				return m, nil
			}
		}

		// Global shortcut keys from any screen (except when editing text).
		if m.Screen == ScreenHome {
			switch key {
			case "i":
				m.Screen = ScreenInit
				m.InitScr = screens.NewInitModel()
				return m, nil
			}
		}
	}

	// Dispatch to active screen.
	var cmd tea.Cmd
	switch m.Screen {
	case ScreenHome:
		m.Home, cmd = screens.HomeUpdate(m.Home, msg)
		// Check if enter was pressed.
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "enter" {
			m.handleHomeSelection()
		}

	case ScreenFilePicker:
		m.FilePicker, cmd = screens.FilePickerUpdate(m.FilePicker, msg)
		// Check if enter was pressed (and not in browse mode).
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "enter" && !m.FilePicker.BrowseMode {
			m.applyFilePicker()
			m.Screen = m.NextScreen
			// Auto-trigger operations for certain screens.
			switch m.NextScreen {
			case ScreenInspect:
				m.Inspect = screens.NewInspectModel()
				m.Inspect.Width = m.Width
				m.Inspect.Height = m.Height
				m.Inspect.LoadContent(m.Cfg)
			case ScreenValidate:
				m.Validate = screens.NewValidateModel()
				m.Validate.Width = m.Width
				m.Validate.Height = m.Height
				m.Validate.RunValidation(m.Cfg)
			case ScreenDump:
				m.Dump = screens.NewDumpModel()
				m.Dump.Width = m.Width
				m.Dump.Height = m.Height
				m.Dump.RunDump(m.Cfg)
			}
		}

	case ScreenInspect:
		m.Inspect, cmd = screens.InspectUpdate(m.Inspect, msg, m.Cfg)

	case ScreenValidate:
		m.Validate, cmd = screens.ValidateUpdate(m.Validate, msg, m.Cfg)

	case ScreenFmt:
		m.Fmt, cmd = screens.FmtUpdate(m.Fmt, msg, m.Cfg)

	case ScreenDump:
		m.Dump, cmd = screens.DumpUpdate(m.Dump, msg, m.Cfg)

	case ScreenInit:
		m.InitScr, cmd = screens.InitUpdate(m.InitScr, msg, m.Cfg)

	case ScreenSettings:
		m.Settings, cmd = screens.SettingsUpdate(m.Settings, msg, m.Cfg)
	}

	return m, cmd
}

// handleHomeSelection routes from the home menu to the appropriate screen.
func (m *Model) handleHomeSelection() {
	switch m.Home.SelectedItem() {
	case screens.HomeInspect:
		m.NextScreen = ScreenInspect
		m.FilePicker = screens.NewFilePickerModel()
		m.Screen = ScreenFilePicker
	case screens.HomeInit:
		m.InitScr = screens.NewInitModel()
		m.Screen = ScreenInit
	case screens.HomeValidate:
		m.NextScreen = ScreenValidate
		m.FilePicker = screens.NewFilePickerModel()
		m.Screen = ScreenFilePicker
	case screens.HomeFmt:
		m.NextScreen = ScreenFmt
		m.FilePicker = screens.NewFilePickerModel()
		m.Screen = ScreenFilePicker
	case screens.HomeDump:
		m.NextScreen = ScreenDump
		m.FilePicker = screens.NewFilePickerModel()
		m.Screen = ScreenFilePicker
	case screens.HomeSettings:
		m.Settings = screens.NewSettingsModel(m.Cfg)
		m.Screen = ScreenSettings
	case screens.HomeQuit:
		// Handled by Update's quit logic when we return tea.Quit.
	}
}

// applyFilePicker reads values from the file picker into the shared config.
func (m *Model) applyFilePicker() {
	cfgPath, dotenv, out := m.FilePicker.Values()
	m.Cfg.ConfigPath = cfgPath
	m.Cfg.DotEnvPath = dotenv
	m.Cfg.OutputPath = out
}

// View implements tea.Model. It renders the header, active screen, and help.
func (m Model) View() string {
	s := components.RenderHeader()

	switch m.Screen {
	case ScreenHome:
		s += screens.HomeView(m.Home)
		// If quit is selected and enter pressed, we already called tea.Quit.
	case ScreenFilePicker:
		s += screens.FilePickerView(m.FilePicker)
	case ScreenInspect:
		s += screens.InspectView(m.Inspect, m.Cfg)
	case ScreenValidate:
		s += screens.ValidateView(m.Validate)
	case ScreenFmt:
		s += screens.FmtScreenView(m.Fmt, m.Cfg)
	case ScreenDump:
		s += screens.DumpView(m.Dump)
	case ScreenInit:
		s += screens.InitScreenView(m.InitScr, m.Cfg)
	case ScreenSettings:
		s += screens.SettingsView(m.Settings, m.Cfg)
	}

	if m.ShowHelp {
		s += "\n" + renderGlobalHelp()
	}

	return s
}

// renderGlobalHelp returns a detailed help overlay.
func renderGlobalHelp() string {
	return components.RenderHelp([]components.HelpBinding{
		{Key: "q/esc", Desc: "back / quit (from home)"},
		{Key: "ctrl+c", Desc: "force quit"},
		{Key: "tab", Desc: "switch tabs / fields"},
		{Key: "enter", Desc: "select / confirm"},
		{Key: "ctrl+s", Desc: "save / write"},
		{Key: "r", Desc: "reload from disk"},
		{Key: "v", Desc: "validate"},
		{Key: "f", Desc: "format"},
		{Key: "d", Desc: "dump"},
		{Key: "i", Desc: "init / template"},
		{Key: "?", Desc: "toggle this help"},
	})
}
