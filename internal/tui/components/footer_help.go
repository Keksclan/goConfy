package components

import "github.com/keksclan/goConfy/internal/tui/style"

// HelpBinding describes a single key → action pair shown in the footer.
type HelpBinding struct {
	Key  string
	Desc string
}

// RenderHelp formats a list of key bindings into a single footer line.
func RenderHelp(bindings []HelpBinding) string {
	var s string
	for i, b := range bindings {
		if i > 0 {
			s += "  "
		}
		s += style.TabActive.Render(b.Key) + " " + style.Muted.Render(b.Desc)
	}
	return style.HelpBar.Render(s)
}

// CommonHelp returns the help bindings that are always shown.
func CommonHelp() []HelpBinding {
	return []HelpBinding{
		{Key: "q/esc", Desc: "back"},
		{Key: "?", Desc: "help"},
	}
}
