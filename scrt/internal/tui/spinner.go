package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// spinnerModel is a minimal bubbletea model that shows a dot spinner with a
// label while a background goroutine runs the operation.
type spinnerModel struct {
	spinner spinner.Model
	label   string
	done    bool
}

type spinnerDoneMsg struct{ err error }

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s  %s\n", m.spinner.View(), m.label)
}

// RunWithSpinner executes fn in a background goroutine while showing a loading
// spinner on stdout. Returns the error produced by fn.
//
// Callers should suppress any structured logger output before calling to avoid
// interleaved text with the spinner rendering.
func RunWithSpinner(label string, fn func() error) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	m := spinnerModel{spinner: s, label: label}
	p := tea.NewProgram(m, tea.WithOutput(os.Stdout))

	var fnErr error
	go func() {
		fnErr = fn()
		p.Send(spinnerDoneMsg{err: fnErr})
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("spinner: %w", err)
	}
	return fnErr
}
