package tui

import (
	"fmt"
	"io"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showLogs opens a scrollable, full-page log viewer for the named container and
// streams its logs into it. Esc closes the page and stops the stream. focusBack
// receives focus when the page is dismissed.
func showLogs(app *tview.Application, pages *tview.Pages, focusBack tview.Primitive, name string, onLogs func(string) (io.ReadCloser, error)) {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	view.SetChangedFunc(func() { app.Draw() })
	view.SetBorder(true)
	view.SetTitle(fmt.Sprintf(" logs: %s  (Esc to close) ", name))

	rc, err := onLogs(name)
	if err != nil {
		fmt.Fprintf(view, "[red]failed to open logs: %v\n", err)
	} else {
		// Close the stream exactly once, whether logs end naturally or the user
		// closes the page (CC-1: the streaming goroutine owns the reader).
		var once sync.Once
		closeStream := func() { once.Do(func() { _ = rc.Close() }) }

		go func() {
			defer closeStream()
			_, _ = io.Copy(tview.ANSIWriter(view), rc)
		}()

		view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				closeStream()
				pages.RemovePage("logs")
				app.SetFocus(focusBack)
				return nil
			}
			return event
		})
	}

	// When opening the logs failed there is no stream; still let Esc close.
	if err != nil {
		view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				pages.RemovePage("logs")
				app.SetFocus(focusBack)
				return nil
			}
			return event
		})
	}

	pages.AddPage("logs", view, true, true)
	app.SetFocus(view)
}
