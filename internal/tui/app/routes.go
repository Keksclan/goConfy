// Package app wires together all TUI screens into a single bubbletea program.
package app

// Screen identifies which screen is currently active.
type Screen int

const (
	ScreenHome       Screen = iota // main workflow menu
	ScreenFilePicker               // file selection (before inspect/validate/fmt/dump)
	ScreenInspect                  // inspect / preview
	ScreenValidate                 // validate
	ScreenFmt                      // format
	ScreenDump                     // dump redacted JSON
	ScreenInit                     // init / template generation
	ScreenSettings                 // settings
)
