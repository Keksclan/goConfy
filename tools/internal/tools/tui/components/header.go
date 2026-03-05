// Package components provides reusable UI building blocks for the TUI.
package components

import "github.com/keksclan/goConfy/tools/internal/tools/tui/style"

// RenderHeader returns the styled application header line.
func RenderHeader() string {
	return style.Header.Render("goConfy TUI") + "\n"
}
