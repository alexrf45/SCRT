package tui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
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
	// Strip any existing tag so we build a clean repo:tag reference.
	repo := stripImageTag(baseImage)

	tag := "latest"
	sel := huh.NewSelect[string]().
		Title(fmt.Sprintf("Pull %s — choose a tag", repo)).
		Options(huh.NewOptions("latest", "dev", "custom...")...).
		Value(&tag)

	if err := sel.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", nil
		}
		return "", fmt.Errorf("pull dialog: %w", err)
	}

	if tag == "custom..." {
		custom := ""
		in := huh.NewInput().
			Title("Custom tag").
			Value(&custom).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("tag must not be empty")
				}
				return nil
			})
		if err := in.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return "", nil
			}
			return "", fmt.Errorf("pull dialog: %w", err)
		}
		tag = strings.TrimSpace(custom)
	}

	return repo + ":" + tag, nil
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
