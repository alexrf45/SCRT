package tui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/alexrf45/scrt/internal/config"
)

// ErrWizardAborted is returned by RunConfigWizard when the user cancels the
// form (Ctrl+C / Esc). Callers should treat it as a clean no-op, not a failure.
var ErrWizardAborted = errors.New("config wizard aborted")

// RunConfigWizard walks the user through SCRT's settings, prefilled from the
// supplied config, and returns a new Config (CS-5: takes a value, returns a new
// value — the input is not mutated). On a non-terminal it runs in huh's
// accessible mode. Returns ErrWizardAborted if the user cancels.
func RunConfigWizard(current config.Config) (config.Config, error) {
	next := current // copy; scalar fields are edited below, slices are carried over untouched

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Docker image").
				Description("Base image for new containers").
				Value(&next.DockerImage).
				Validate(nonEmpty("docker image")),
			huh.NewInput().
				Title("Container shell").
				Value(&next.ContainerShell).
				Validate(nonEmpty("container shell")),
			huh.NewInput().
				Title("Base working directory").
				Value(&next.WorkDirBase).
				Validate(nonEmpty("working directory")),
			huh.NewConfirm().
				Title("Enable host networking?").
				Value(&next.HostNetworking),
			huh.NewConfirm().
				Title("Enable X11 forwarding?").
				Value(&next.EnableX11),
			huh.NewConfirm().
				Title("Enable GPU passthrough?").
				Value(&next.EnableGPU),
		),
	).WithAccessible(!isTerminal())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return current, ErrWizardAborted
		}
		return current, fmt.Errorf("config wizard: %w", err)
	}

	return next, nil
}

// nonEmpty returns a huh validator that rejects blank input for the named field.
func nonEmpty(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s must not be empty", field)
		}
		return nil
	}
}
