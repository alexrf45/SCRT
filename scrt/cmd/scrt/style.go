package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	styleInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("69"))

	styleWarn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

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

// printSuccess prints a green ✓ confirmation line.
func printSuccess(message string) {
	fmt.Printf("%s  %s\n", styleSuccess.Render("✓"), message)
}

// printInfo prints a blue informational line.
func printInfo(message string) {
	fmt.Printf("%s  %s\n", styleInfo.Render("·"), message)
}

// printWarn prints an orange warning line.
func printWarn(message string) {
	fmt.Printf("%s  %s\n", styleWarn.Render("!"), message)
}
