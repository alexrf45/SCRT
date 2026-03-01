package tui

import (
	"fmt"
	"strings"

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

	// Strip any existing tag so we build a clean repo:tag reference.
	repo := stripImageTag(baseImage)

	// Track the currently selected tag option.
	currentTag := "latest"

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(fmt.Sprintf(" Pull: %s ", repo))
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
		result = repo + ":" + tag
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

// stripImageTag returns the image reference without its tag.
// Examples: "fonalex45/scrt:latest" → "fonalex45/scrt"
//
//	"registry.example.com:5000/img:v1" → "registry.example.com:5000/img"
//	"fonalex45/scrt" → "fonalex45/scrt"
func stripImageTag(image string) string {
	idx := strings.LastIndex(image, ":")
	if idx == -1 {
		return image
	}
	// If the substring after the last colon contains a slash it is a
	// registry port (e.g. "registry:5000/img"), not a tag — leave it.
	if strings.Contains(image[idx+1:], "/") {
		return image
	}
	return image[:idx]
}
