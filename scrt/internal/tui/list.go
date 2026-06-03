// Package tui provides interactive terminal user interfaces for SCRT.
package tui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	// OnLogs returns a plain-text log stream for the named container.
	OnLogs func(name string) (io.ReadCloser, error)
	// OnCopyTo uploads a host file into the container at dstPath.
	OnCopyTo func(name, hostPath, dstPath string) error
	// OnCopyFrom downloads srcPath from the container into a host directory.
	OnCopyFrom func(name, srcPath, destDir string) error
}

const (
	listHeaderRow       = 0
	listDataStart       = 1
	statusClearDelay    = 3 * time.Second
	autoRefreshInterval = 5 * time.Second
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

	var (
		lastErr error
		busy    bool           // guards concurrent Docker ops; only touched on the main goroutine
		master  = p.Containers // full, unfiltered container set; the source of truth
		filter  string         // active case-insensitive filter query
		auto    bool           // whether background auto-refresh is enabled
	)

	table := tview.NewTable().SetBorders(false).SetSelectable(true, false)

	// setContainers updates the master set and redraws the table through the
	// active filter. Main-goroutine only.
	setContainers := func(containers []container.Info) {
		master = containers
		populateTable(table, filterContainers(master, filter))
	}
	setContainers(p.Containers)

	footer := tview.NewTextView().SetDynamicColors(true)
	footerHints(footer)

	// filterInput is a hidden bar revealed with '/'. Typing filters the table
	// live; Enter keeps the filter, Esc clears it. Both return focus to the table.
	filterInput := tview.NewInputField().SetLabel("/ ")
	filterInput.SetChangedFunc(func(text string) {
		filter = text
		populateTable(table, filterContainers(master, filter))
	})

	mainPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(filterInput, 0, 0, false). // hidden until '/'
		AddItem(footer, 1, 0, false)

	closeFilter := func() {
		mainPage.ResizeItem(filterInput, 0, 0)
		app.SetFocus(table)
	}
	filterInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			filterInput.SetText("") // ChangedFunc clears filter and repopulates
		}
		closeFilter()
	})

	pages := tview.NewPages().AddPage("main", mainPage, true, true)

	// startOp marks the UI as busy and updates the footer. Returns false if
	// another op is already in flight; callers must check and bail out.
	// Must only be called on the tview main goroutine.
	startOp := func(msg string) bool {
		if busy {
			return false
		}
		busy = true
		footer.SetText("[yellow]" + msg)
		return true
	}

	// finishOp clears the busy flag, shows a success or error message, and
	// schedules the hint bar to restore after statusClearDelay.
	// Must only be called via app.QueueUpdateDraw (i.e. on the main goroutine).
	finishOp := func(successMsg string, err error) {
		busy = false
		if err != nil {
			footer.SetText("[red]✗  " + err.Error())
		} else {
			footer.SetText("[green]✓  " + successMsg)
		}
		scheduleHintRestore(app, footer, &busy)
	}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()

		// Navigation and quit — always responsive even while ops are running.
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
		case '/':
			mainPage.ResizeItem(filterInput, 1, 0)
			app.SetFocus(filterInput)
			return nil
		case 'a':
			auto = !auto
			if auto {
				footer.SetText("[green]auto-refresh on (every 5s)")
			} else {
				footer.SetText("[yellow]auto-refresh off")
			}
			scheduleHintRestore(app, footer, &busy)
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			app.Stop()
			return nil
		}

		// Action keys require a data row.
		if row < listDataStart {
			return event
		}

		name := table.GetCell(row, 0).Text

		switch event.Rune() {
		case 'e':
			if busy {
				return nil
			}
			// Suspend hands terminal control to docker exec -it, then resumes tview.
			app.Suspend(func() {
				if err := p.OnEnter(name); err != nil {
					lastErr = err
				}
			})
			return nil

		case 'l':
			if p.OnLogs == nil {
				return nil
			}
			showLogs(app, pages, table, name, p.OnLogs)
			return nil

		case 'u':
			if busy || p.OnCopyTo == nil {
				return nil
			}
			showCopyForm(app, pages, table, fmt.Sprintf(" Upload to %q ", name),
				"Host file", "Container dir", func(hostPath, dstPath string) {
					if !startOp("Uploading to " + name + "…") {
						return
					}
					go func() {
						err := p.OnCopyTo(name, hostPath, dstPath)
						app.QueueUpdateDraw(func() { finishOp("Uploaded to "+name, err) })
					}()
				})
			return nil

		case 'g':
			if busy || p.OnCopyFrom == nil {
				return nil
			}
			showCopyForm(app, pages, table, fmt.Sprintf(" Download from %q ", name),
				"Container path", "Host dir", func(srcPath, destDir string) {
					if !startOp("Downloading from " + name + "…") {
						return
					}
					go func() {
						err := p.OnCopyFrom(name, srcPath, destDir)
						app.QueueUpdateDraw(func() { finishOp("Downloaded from "+name, err) })
					}()
				})
			return nil

		case 's':
			if busy {
				return nil
			}
			showConfirm(app, pages, table, fmt.Sprintf("Stop container %q?", name), func() {
				if !startOp("Stopping " + name + "…") {
					return
				}
				go func() {
					err := p.OnStop(name)
					app.QueueUpdateDraw(func() {
						if err == nil {
							if containers, rerr := p.OnRefresh(); rerr == nil {
								setContainers(containers)
							}
						}
						finishOp("Stopped "+name, err)
					})
				}()
			})
			return nil

		case 'd':
			if busy {
				return nil
			}
			showConfirm(app, pages, table,
				fmt.Sprintf("Destroy %q?\nThis cannot be undone.", name),
				func() {
					if !startOp("Destroying " + name + "…") {
						return
					}
					go func() {
						err := p.OnDestroy(name)
						app.QueueUpdateDraw(func() {
							if err == nil {
								if containers, rerr := p.OnRefresh(); rerr == nil {
									setContainers(containers)
								}
							}
							finishOp("Destroyed "+name, err)
						})
					}()
				},
			)
			return nil

		case 'b':
			if !startOp("Backing up " + name + "…") {
				return nil
			}
			go func() {
				err := p.OnBackup(name)
				app.QueueUpdateDraw(func() {
					finishOp("Backup complete — "+name, err)
				})
			}()
			return nil

		case 'r':
			if !startOp("Refreshing…") {
				return nil
			}
			go func() {
				containers, err := p.OnRefresh()
				app.QueueUpdateDraw(func() {
					if err == nil {
						setContainers(containers)
					}
					finishOp("Refreshed", err)
				})
			}()
			return nil
		}

		return event
	})

	// Auto-refresh ticker. The goroutine only handles timing; all state
	// (auto, busy) is read and mutated on the main goroutine inside
	// QueueUpdateDraw, so no locking is needed. The actual Docker call runs in a
	// detached goroutine to avoid blocking the UI, and refreshes silently
	// (no footer churn) to keep the hint bar intact.
	stopTicker := make(chan struct{})
	go func() {
		ticker := time.NewTicker(autoRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stopTicker:
				return
			case <-ticker.C:
				app.QueueUpdateDraw(func() {
					if !auto || busy {
						return
					}
					busy = true
					go func() {
						containers, err := p.OnRefresh()
						app.QueueUpdateDraw(func() {
							busy = false
							if err == nil {
								setContainers(containers)
							}
						})
					}()
				})
			}
		}
	}()

	err := app.SetRoot(pages, true).EnableMouse(true).Run()
	close(stopTicker)
	if err != nil {
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

// footerHints sets the key-hint bar in the footer.
func footerHints(tv *tview.TextView) {
	tv.SetText("[yellow]jk[white] nav  [yellow]/[white] filter  [yellow]e[white]nter  [yellow]l[white]ogs  [yellow]s[white]top  [yellow]d[white]estroy  [yellow]b[white]ackup  [yellow]u[white]pload  [yellow]g[white]et  [yellow]r[white]efresh  [yellow]a[white]uto  [yellow]q[white]uit")
}

// scheduleHintRestore restores the footer key-hint bar after statusClearDelay,
// unless an operation is in flight when the timer fires.
func scheduleHintRestore(app *tview.Application, footer *tview.TextView, busy *bool) {
	go func() {
		time.Sleep(statusClearDelay)
		app.QueueUpdateDraw(func() {
			if !*busy {
				footerHints(footer)
			}
		})
	}()
}

// filterContainers returns the containers whose name, image, or state contains
// the query (case-insensitive). An empty query returns the input unchanged.
func filterContainers(containers []container.Info, query string) []container.Info {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return containers
	}
	out := make([]container.Info, 0, len(containers))
	for _, c := range containers {
		if strings.Contains(strings.ToLower(c.Name), query) ||
			strings.Contains(strings.ToLower(c.Image), query) ||
			strings.Contains(strings.ToLower(c.State), query) {
			out = append(out, c)
		}
	}
	return out
}

// showConfirm displays a modal confirmation dialog. focusBack receives focus
// when the dialog is dismissed. onConfirm is called only on "Confirm".
func showConfirm(app *tview.Application, pages *tview.Pages, focusBack tview.Primitive, msg string, onConfirm func()) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"Confirm", "Cancel"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage("confirm")
			app.SetFocus(focusBack)
			if buttonLabel == "Confirm" {
				onConfirm()
			}
		})

	pages.AddPage("confirm", modal, false, true)
	app.SetFocus(modal)
}
