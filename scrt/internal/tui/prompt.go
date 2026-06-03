package tui

import (
	"errors"
	"fmt"
	"os"

	"charm.land/huh/v2"
)

// Confirm asks the user a yes/no question and returns their choice. On a
// terminal it shows an interactive huh prompt; otherwise it falls back to huh's
// accessible mode reading from stdin, so piped input (e.g. `echo y | scrt …`)
// still works. A user abort (Ctrl+C / Esc) is reported as a negative answer,
// not an error.
func Confirm(title string, value bool) (bool, error) {
	confirm := huh.NewConfirm().
		Title(title).
		Affirmative("Yes").
		Negative("No").
		Value(&value)

	var err error
	if isTerminal() {
		err = confirm.Run()
	} else {
		err = confirm.RunAccessible(os.Stdout, os.Stdin)
	}
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, fmt.Errorf("confirm: %w", err)
	}
	return value, nil
}
