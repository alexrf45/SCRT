package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RunPullDialog shows an interactive tag-selection dialog and returns the
// selected image:tag string. Returns an empty string if the user cancels.
// Falls back to an empty string (no dialog) when stdout is not a terminal.
func RunPullDialog(baseImage string) (string, error) {
	if !isTerminal() {
		return "", nil
	}
	return runPullDialog(baseImage)
}

func runPullDialog(baseImage string) (string, error) {
	app := tview.NewApplication()
	result := ""

	// Track the currently selected tag option.
	currentTag := "latest"

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(fmt.Sprintf(" Pull: %s ", baseImage))
	form.SetTitleAlign(tview.AlignLeft)

	customField := tview.NewInputField().
		SetLabel("Custom Tag: ").
		SetFieldWidth(20)

	form.AddDropDown("Tag", []string{"latest", "dev", "custom..."}, 0, func(option string, _ int) {
		currentTag = option
	})
	form.AddFormItem(customField)

	form.AddButton("Pull", func() {
		tag := currentTag
		if tag == "custom..." {
			tag = customField.GetText()
			if tag == "" {
				return // require a custom tag value before proceeding
			}
		}
		result = baseImage + ":" + tag
		app.Stop()
	})

	form.AddButton("Cancel", func() {
		app.Stop()
	})

	// Allow Escape to cancel.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.Stop()
			return nil
		}
		return event
	})

	// Center the form in the terminal.
	center := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 14, 0, true).
			AddItem(nil, 0, 1, false),
			55, 0, true).
		AddItem(nil, 0, 1, false)

	if err := app.SetRoot(center, true).Run(); err != nil {
		return "", fmt.Errorf("pull dialog: %w", err)
	}

	return result, nil
}
