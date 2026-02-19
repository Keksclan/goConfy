package components

import "github.com/keksclan/goConfy/internal/tui/style"

// RenderConfirm draws a yes/no confirmation prompt.
func RenderConfirm(question string, yesSelected bool) string {
	yes := style.ToggleOff.Render("[ Yes ]")
	no := style.ToggleOff.Render("[ No ]")
	if yesSelected {
		yes = style.ToggleOn.Render("[ Yes ]")
	} else {
		no = style.StatusErr.Render("[ No ]")
	}
	return style.Warning.Render(question) + "\n\n" + yes + "  " + no
}
