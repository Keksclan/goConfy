package components

import "github.com/charmbracelet/bubbles/viewport"

// NewViewport creates a viewport with sensible defaults.
func NewViewport(width, height int) viewport.Model {
	vp := viewport.New(width, height)
	return vp
}
