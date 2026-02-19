// goconfytui is an interactive TUI for goConfy.
//
// It provides the same workflows as the goconfygen CLI (init, validate, fmt,
// dump) in a terminal user interface powered by Charm's bubbletea.
//
// Build:
//
//	go build ./cmd/goconfytui
//
// Install:
//
//	go install github.com/keksclan/goConfy/cmd/goconfytui@latest
//
// Run:
//
//	./goconfytui
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/keksclan/goConfy/internal/tui/app"
)

func main() {
	p := tea.NewProgram(app.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "goconfytui: %v\n", err)
		os.Exit(1)
	}
}
