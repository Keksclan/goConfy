// Package style provides lipgloss styles used across all TUI screens.
// Keeping them centralised ensures a consistent look and makes it easy
// to tweak colours or padding in one place.
package style

import "github.com/charmbracelet/lipgloss"

// Colours – intentionally kept small so the TUI works well on both dark
// and light terminal backgrounds.
var (
	ColorPrimary   = lipgloss.Color("63")  // purple-ish
	ColorSecondary = lipgloss.Color("39")  // blue
	ColorSuccess   = lipgloss.Color("42")  // green
	ColorError     = lipgloss.Color("196") // red
	ColorMuted     = lipgloss.Color("241") // grey
	ColorWarning   = lipgloss.Color("214") // orange
)

// Header renders the top bar with the application title.
var Header = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPrimary).
	PaddingLeft(1).
	PaddingRight(1)

// Title is used for screen titles below the header.
var Title = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorSecondary).
	MarginBottom(1)

// MenuItem renders a single menu entry.
var MenuItem = lipgloss.NewStyle().
	PaddingLeft(2)

// MenuItemSelected highlights the currently focused menu entry.
var MenuItemSelected = lipgloss.NewStyle().
	PaddingLeft(1).
	Foreground(ColorPrimary).
	Bold(true).
	SetString("> ")

// StatusOK is the style for "VALID" / success messages.
var StatusOK = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorSuccess)

// StatusErr is the style for "INVALID" / error messages.
var StatusErr = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorError)

// Muted is used for help text and secondary information.
var Muted = lipgloss.NewStyle().
	Foreground(ColorMuted)

// Warning renders cautionary text.
var Warning = lipgloss.NewStyle().
	Foreground(ColorWarning)

// HelpBar renders the bottom help / shortcut bar.
var HelpBar = lipgloss.NewStyle().
	Foreground(ColorMuted).
	PaddingTop(1)

// Border wraps content in a rounded border for viewports.
var Border = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorMuted).
	Padding(0, 1)

// TabActive renders the currently active tab label.
var TabActive = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPrimary).
	Underline(true).
	PaddingRight(2)

// TabInactive renders an inactive tab label.
var TabInactive = lipgloss.NewStyle().
	Foreground(ColorMuted).
	PaddingRight(2)

// Label renders a form label.
var Label = lipgloss.NewStyle().
	Bold(true).
	Width(20)

// ToggleOn renders an enabled toggle value.
var ToggleOn = lipgloss.NewStyle().
	Foreground(ColorSuccess).
	Bold(true)

// ToggleOff renders a disabled toggle value.
var ToggleOff = lipgloss.NewStyle().
	Foreground(ColorMuted)
