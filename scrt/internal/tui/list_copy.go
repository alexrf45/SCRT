package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCopyForm opens a centered two-field form for a file-copy operation. On
// "Copy" it dismisses the form and calls onSubmit with the two field values;
// Esc or "Cancel" dismisses without submitting. focusBack receives focus when
// the form closes.
func showCopyForm(app *tview.Application, pages *tview.Pages, focusBack tview.Primitive, title, label1, label2 string, onSubmit func(v1, v2 string)) {
	var v1, v2 string

	dismiss := func() {
		pages.RemovePage("copy")
		app.SetFocus(focusBack)
	}

	form := tview.NewForm().
		AddInputField(label1, "", 40, nil, func(t string) { v1 = t }).
		AddInputField(label2, "", 40, nil, func(t string) { v2 = t }).
		AddButton("Copy", func() {
			dismiss()
			onSubmit(v1, v2)
		}).
		AddButton("Cancel", dismiss)

	// SetBorder/SetTitle are Box methods returning *tview.Box, so they must be
	// separate statements, not chained onto the *tview.Form above.
	form.SetBorder(true)
	form.SetTitle(title)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			dismiss()
			return nil
		}
		return event
	})

	pages.AddPage("copy", centered(form, 60, 9), true, true)
	app.SetFocus(form)
}

// centered wraps a primitive in a flex layout that fixes its width and height
// and centers it in the available space.
func centered(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 0, true).
			AddItem(nil, 0, 1, false),
			width, 0, true).
		AddItem(nil, 0, 1, false)
}
