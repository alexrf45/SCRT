// Package container manages Docker container lifecycle for SCRT.
package container

import "errors"

// Sentinel errors for container operations.
// Use errors.Is for control flow (ERR-2).
var (
	// ErrDockerUnavailable indicates the Docker daemon cannot be reached.
	ErrDockerUnavailable = errors.New("docker daemon is not accessible")

	// ErrInvalidProject indicates a project name failed validation.
	ErrInvalidProject = errors.New("invalid project name: use only alphanumeric characters, hyphens, and underscores")

	// ErrContainerExists indicates a container with the given name already exists.
	ErrContainerExists = errors.New("container already exists")

	// ErrContainerNotFound indicates no container with the given name was found.
	ErrContainerNotFound = errors.New("container not found")

	// ErrX11Unavailable indicates X11 forwarding was requested but the host
	// environment does not have DISPLAY set or Xauthority is missing.
	ErrX11Unavailable = errors.New("X11 forwarding requested but DISPLAY or Xauthority unavailable")

	// ErrContainerNotRunning indicates the container exists but is not in a running state.
	ErrContainerNotRunning = errors.New("container is not running")
)
