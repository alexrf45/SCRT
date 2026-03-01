package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
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
