package components

import (
	"fmt"

	"github.com/keksclan/goConfy/tools/internal/tools/tui/style"
)

// RenderList draws a vertical list with the cursor at the selected index.
func RenderList(items []string, cursor int) string {
	var s string
	for i, item := range items {
		if i == cursor {
			s += style.MenuItemSelected.Render("") + style.MenuItemSelected.Render(item) + "\n"
		} else {
			s += style.MenuItem.Render(item) + "\n"
		}
	}
	return s
}

// RenderToggle draws a labelled boolean toggle.
func RenderToggle(label string, value bool) string {
	val := style.ToggleOff.Render("OFF")
	if value {
		val = style.ToggleOn.Render("ON")
	}
	return fmt.Sprintf("%s %s", style.Label.Render(label), val)
}
