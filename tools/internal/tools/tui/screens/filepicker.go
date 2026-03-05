package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/components"
	"github.com/keksclan/goConfy/tools/internal/tools/tui/style"
)

// FilePickerField identifies which path field is being edited.
type FilePickerField int

const (
	FieldConfigPath FilePickerField = iota
	FieldDotEnvPath
	FieldOutputPath
)

// FilePickerModel provides text inputs for config, dotenv, and output paths,
// plus an optional directory browser.
type FilePickerModel struct {
	Inputs      [3]textinput.Model
	Focus       int      // which input is focused (0-2)
	BrowseMode  bool     // true when the directory browser is active
	BrowseDir   string   // current directory in browser
	BrowseItems []string // file/dir names in current directory
	BrowseCur   int      // cursor position in browser
	Err         error
}

// NewFilePickerModel creates a file picker with labelled inputs.
func NewFilePickerModel() FilePickerModel {
	m := FilePickerModel{}
	m.Inputs[FieldConfigPath] = components.NewTextInput("config.yml")
	m.Inputs[FieldDotEnvPath] = components.NewTextInput(".env (optional)")
	m.Inputs[FieldOutputPath] = components.NewTextInput("output path (optional)")
	m.Inputs[FieldConfigPath].Focus()
	return m
}

// Values returns the current text input values.
func (m FilePickerModel) Values() (configPath, dotenvPath, outputPath string) {
	return m.Inputs[FieldConfigPath].Value(),
		m.Inputs[FieldDotEnvPath].Value(),
		m.Inputs[FieldOutputPath].Value()
}

// FilePickerUpdate handles key events for the file picker screen.
func FilePickerUpdate(m FilePickerModel, msg tea.Msg) (FilePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// In browse mode, handle navigation.
		if m.BrowseMode {
			switch key {
			case "esc", "q":
				m.BrowseMode = false
				return m, nil
			case "up", "k":
				if m.BrowseCur > 0 {
					m.BrowseCur--
				}
			case "down", "j":
				if m.BrowseCur < len(m.BrowseItems)-1 {
					m.BrowseCur++
				}
			case "enter":
				if len(m.BrowseItems) > 0 {
					selected := m.BrowseItems[m.BrowseCur]
					if selected == ".." {
						m.BrowseDir = filepath.Dir(m.BrowseDir)
						m.loadBrowseDir()
					} else {
						full := filepath.Join(m.BrowseDir, selected)
						info, err := os.Stat(full)
						if err == nil && info.IsDir() {
							m.BrowseDir = full
							m.loadBrowseDir()
						} else {
							// File selected – set into current input.
							m.Inputs[m.Focus].SetValue(full)
							m.BrowseMode = false
						}
					}
				}
			case "backspace":
				m.BrowseDir = filepath.Dir(m.BrowseDir)
				m.loadBrowseDir()
			}
			return m, nil
		}

		// Normal input mode.
		switch key {
		case "tab", "down":
			m.Inputs[m.Focus].Blur()
			m.Focus = (m.Focus + 1) % len(m.Inputs)
			m.Inputs[m.Focus].Focus()
			return m, nil
		case "shift+tab", "up":
			m.Inputs[m.Focus].Blur()
			m.Focus = (m.Focus - 1 + len(m.Inputs)) % len(m.Inputs)
			m.Inputs[m.Focus].Focus()
			return m, nil
		case "ctrl+b":
			// Open directory browser.
			dir, _ := os.Getwd()
			m.BrowseDir = dir
			m.BrowseMode = true
			m.loadBrowseDir()
			return m, nil
		}
	}

	// Delegate to the focused text input.
	var cmd tea.Cmd
	m.Inputs[m.Focus], cmd = m.Inputs[m.Focus].Update(msg)
	return m, cmd
}

// loadBrowseDir populates BrowseItems with entries from BrowseDir.
func (m *FilePickerModel) loadBrowseDir() {
	m.BrowseCur = 0
	m.BrowseItems = []string{".."}

	entries, err := os.ReadDir(m.BrowseDir)
	if err != nil {
		m.Err = err
		return
	}

	var dirs, files []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && name != ".env" {
			continue // skip hidden files except .env
		}
		if e.IsDir() {
			dirs = append(dirs, name+"/")
		} else {
			files = append(files, name)
		}
	}
	sort.Strings(dirs)
	sort.Strings(files)
	m.BrowseItems = append(m.BrowseItems, dirs...)
	m.BrowseItems = append(m.BrowseItems, files...)
}

// FilePickerView renders the file picker screen.
func FilePickerView(m FilePickerModel) string {
	labels := []string{"Config YAML:", ".env file:", "Output path:"}

	var s string
	s += style.Title.Render("File Selection") + "\n\n"

	if m.BrowseMode {
		s += style.Warning.Render("Directory: "+m.BrowseDir) + "\n\n"
		s += components.RenderList(m.BrowseItems, m.BrowseCur)
		s += "\n"
		s += components.RenderHelp([]components.HelpBinding{
			{Key: "↑/↓", Desc: "navigate"},
			{Key: "enter", Desc: "select"},
			{Key: "backspace", Desc: "parent dir"},
			{Key: "esc", Desc: "close browser"},
		})
		return s
	}

	for i, lbl := range labels {
		cursor := " "
		if i == m.Focus {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s %s\n", cursor, style.Label.Render(lbl), m.Inputs[i].View())
	}

	if m.Err != nil {
		s += "\n" + style.StatusErr.Render(m.Err.Error())
	}

	s += "\n"
	s += components.RenderHelp([]components.HelpBinding{
		{Key: "tab/↑↓", Desc: "switch field"},
		{Key: "ctrl+b", Desc: "browse"},
		{Key: "enter", Desc: "confirm"},
		{Key: "esc", Desc: "back"},
	})
	return s
}
