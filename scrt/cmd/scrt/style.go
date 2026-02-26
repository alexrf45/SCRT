package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table column widths used by the list command.
const (
	colWidthName    = 22
	colWidthState   = 12
	colWidthImage   = 38
	colWidthStatus  = 30
)

var (
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true)

	styleCell = lipgloss.NewStyle().Padding(0, 1)

	styleRunning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	styleStopped = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	styleExited = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	styleConfigKey = lipgloss.NewStyle().
			Foreground(lipgloss.Color("69")).
			Width(20)

	styleConfigVal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	styleConfigPath = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	styleBold = lipgloss.NewStyle().Bold(true)

	// styleOp is used for pre-operation banners before docker takes over the terminal.
	styleOp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("69")).
		Bold(true)
)

// printOp prints a styled banner before a docker operation takes over the terminal.
// Format:  →  <message>
func printOp(message string) {
	fmt.Printf("%s  %s\n", styleOp.Render("→"), message)
}

// renderState applies a color style based on the container state string.
func renderState(state string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "running":
		return styleRunning.Render(state)
	case "exited":
		return styleExited.Render(state)
	default:
		return styleStopped.Render(state)
	}
}
