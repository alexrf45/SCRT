// Package tui provides interactive terminal user interfaces for SCRT.
package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexrf45/scrt/internal/container"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-isatty"
	"github.com/rivo/tview"
)

// ListParams holds configuration and action callbacks for the container browser.
type ListParams struct {
	Containers []container.Info
	OnEnter    func(name string) error
	OnStop     func(name string) error
	OnDestroy  func(name string) error
	OnBackup   func(name string) error
	OnRefresh  func() ([]container.Info, error)
}

const (
	listHeaderRow = 0
	listDataStart = 1
)

// RunList launches the interactive container browser. Blocks until the user quits.
// Falls back to static table output when stdout is not a terminal.
func RunList(p ListParams) error {
	if !isTerminal() {
		return renderStaticList(p.Containers)
	}
	return runInteractiveList(p)
}

func isTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

// renderStaticList prints a plain-text table to stdout for non-TTY environments.
func renderStaticList(containers []container.Info) error {
	if len(containers) == 0 {
		fmt.Println("No SCRT containers found.")
		return nil
	}

	fmt.Printf("%-22s %-12s %-38s %-30s\n", "NAME", "STATE", "IMAGE", "STATUS")
	fmt.Println(strings.Repeat("-", 106))
	for _, c := range containers {
		fmt.Printf("%-22s %-12s %-38s %-30s\n", c.Name, c.State, c.Image, c.Status)
	}
	return nil
}

func runInteractiveList(p ListParams) error {
	app := tview.NewApplication()

	var lastErr error

	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	populateTable(table, p.Containers)

	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]↑↓/jk[white] nav  [yellow]e[white]nter  [yellow]s[white]top  [yellow]d[white]estroy  [yellow]b[white]ackup  [yellow]r[white]efresh  [yellow]q[white]uit")

	mainPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(footer, 1, 0, false)

	pages := tview.NewPages().AddPage("main", mainPage, true, true)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()

		// Global navigation and quit.
		switch event.Rune() {
		case 'q':
			app.Stop()
			return nil
		case 'j':
			if row < table.GetRowCount()-1 {
				table.Select(row+1, 0)
			}
			return nil
		case 'k':
			if row > listDataStart {
				table.Select(row-1, 0)
			}
			return nil
		}

		if event.Key() == tcell.KeyEscape {
			app.Stop()
			return nil
		}

		// Actions below require a data row to be selected.
		if row < listDataStart {
			return event
		}

		name := table.GetCell(row, 0).Text

		switch event.Rune() {
		case 'e':
			app.Suspend(func() {
				if err := p.OnEnter(name); err != nil {
					lastErr = err
				}
			})
			return nil

		case 's':
			showConfirm(app, pages, fmt.Sprintf("Stop container %q?", name), func() {
				if err := p.OnStop(name); err != nil {
					lastErr = err
					return
				}
				if containers, err := p.OnRefresh(); err == nil {
					populateTable(table, containers)
					app.Draw()
				}
			})
			return nil

		case 'd':
			showConfirm(app, pages,
				fmt.Sprintf("Destroy container %q?\nThis action cannot be undone.", name),
				func() {
					if err := p.OnDestroy(name); err != nil {
						lastErr = err
						return
					}
					if containers, err := p.OnRefresh(); err == nil {
						populateTable(table, containers)
						app.Draw()
					}
				},
			)
			return nil

		case 'b':
			if err := p.OnBackup(name); err != nil {
				lastErr = err
			}
			return nil

		case 'r':
			if containers, err := p.OnRefresh(); err == nil {
				populateTable(table, containers)
				app.Draw()
			}
			return nil
		}

		return event
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("tui list: %w", err)
	}

	return lastErr
}

// populateTable fills the table with a header row and one row per container.
func populateTable(table *tview.Table, containers []container.Info) {
	table.Clear()

	headers := []string{"NAME", "STATE", "IMAGE", "STATUS"}
	for col, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(tcell.ColorAqua).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1)
		table.SetCell(listHeaderRow, col, cell)
	}

	for i, c := range containers {
		row := i + listDataStart

		stateColor := tcell.ColorOrange
		switch strings.ToLower(c.State) {
		case "running":
			stateColor = tcell.ColorGreen
		case "exited":
			stateColor = tcell.ColorRed
		}

		table.SetCell(row, 0, tview.NewTableCell(c.Name).SetMaxWidth(22).SetExpansion(1))
		table.SetCell(row, 1, tview.NewTableCell(c.State).SetTextColor(stateColor).SetMaxWidth(12).SetExpansion(1))
		table.SetCell(row, 2, tview.NewTableCell(c.Image).SetMaxWidth(38).SetExpansion(2))
		table.SetCell(row, 3, tview.NewTableCell(c.Status).SetMaxWidth(30).SetExpansion(1))
	}

	if len(containers) > 0 {
		table.Select(listDataStart, 0)
	}
}

// showConfirm displays a modal confirmation dialog over the given pages.
// onConfirm is called only when the user selects "Confirm".
func showConfirm(app *tview.Application, pages *tview.Pages, msg string, onConfirm func()) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"Confirm", "Cancel"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage("confirm")
			if buttonLabel == "Confirm" {
				onConfirm()
			}
		})

	pages.AddPage("confirm", modal, false, true)
	app.SetFocus(modal)
}
