package components

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// NewTextInput creates a pre-configured text input with the given placeholder.
func NewTextInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 512
	ti.Width = 60
	return ti
}
