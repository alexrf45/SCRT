package container

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// projectNameRe validates project names: alphanumeric, hyphens, underscores only.
var projectNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateProjectName checks whether a project name is safe for use as a
// container name and filesystem path component.
func ValidateProjectName(name string) error {
	if !projectNameRe.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidProject, name)
	}
	return nil
}

// RunParams configures a new SCRT container. (CS-5, CS-6)
type RunParams struct {
	Project     string
	Image       string
	Version     string
	WorkDirBase string
	Shell       string
	Network     NetworkParams
	Display     DisplayParams
	GPU         GPUParams
	Caps        []string
	ExtraMounts []string
	Env         EnvParams
}

// NetworkParams controls container networking.
type NetworkParams struct {
	HostMode bool
}

// DisplayParams controls X11 forwarding.
// Strict all-or-nothing: creation fails if Enabled is true but the host
// environment lacks DISPLAY or Xauthority.
type DisplayParams struct {
	Enabled bool
}

// GPUParams controls DRI device passthrough (/dev/dri).
type GPUParams struct {
	Enabled bool
}

// EnvParams holds environment variables injected into the container.
type EnvParams struct {
	Project string
	Target  string
	Domain  string
	TZ      string
}

// BuildRunArgs transforms RunParams into a docker run argument list.
// The returned slice does NOT include "docker" or "run" — the caller prepends those.
// Returns ErrX11Unavailable if X11 is enabled but the host environment is not suitable.
func BuildRunArgs(params RunParams) ([]string, error) {
	if err := ValidateProjectName(params.Project); err != nil {
		return nil, err
	}

	var args []string

	// Container name and interactive TTY
	args = append(args, "--name", params.Project, "-it")

	// Labels
	for k, v := range Labels(params.Project, params.Version) {
		args = append(args, "--label", k+"="+v)
	}

	// Network configuration
	if params.Network.HostMode {
		args = append(args, "--net=host")
	}

	// Linux capabilities (deduplicated)
	for _, cap := range deduplicateCaps(params.Caps) {
		args = append(args, "--cap-add="+cap)
	}

	// GPU passthrough (non-fatal if /dev/dri missing)
	if params.GPU.Enabled {
		if _, err := os.Stat("/dev/dri"); err == nil {
			args = append(args, "--device=/dev/dri:/dev/dri")
		}
	}

	// X11 forwarding (strict — fails if DISPLAY or Xauthority missing)
	if params.Display.Enabled {
		x11Args, err := buildX11Args()
		if err != nil {
			return nil, err
		}
		args = append(args, x11Args...)
	}

	// Environment variables
	args = append(args, buildEnvArgs(params.Env)...)

	// Volume mounts
	args = append(args, buildMountArgs(params.Project, params.WorkDirBase, params.ExtraMounts)...)

	// Working directory and entrypoint
	args = append(args, "-w", "/"+params.Project)
	args = append(args, "--entrypoint", params.Shell)

	// Image (must be last before any command args)
	args = append(args, params.Image)

	return args, nil
}

// buildX11Args validates the host X11 environment and returns docker args
// for DISPLAY forwarding. Returns ErrX11Unavailable on failure (strict).
func buildX11Args() ([]string, error) {
	display := os.Getenv("DISPLAY")
	if display == "" {
		return nil, fmt.Errorf("%w: DISPLAY is not set", ErrX11Unavailable)
	}

	xauth := filepath.Join(os.Getenv("HOME"), ".Xauthority")
	if _, err := os.Stat(xauth); err != nil {
		return nil, fmt.Errorf("%w: Xauthority not found at %s", ErrX11Unavailable, xauth)
	}

	return []string{
		"-e", "DISPLAY=" + display,
		"-v", "/tmp/.X11-unix:/tmp/.X11-unix",
		"-v", xauth + ":" + xauth + ":ro",
	}, nil
}

// buildEnvArgs constructs -e flags for container environment variables.
func buildEnvArgs(env EnvParams) []string {
	args := []string{
		"-e", "PROJECT=" + env.Project,
		"-e", "TARGET=" + env.Target,
		"-e", "TZ=" + env.TZ,
	}
	if env.Domain != "" {
		args = append(args, "-e", "DOMAIN="+env.Domain)
	}
	return args
}

// buildMountArgs constructs -v flags for bind mounts.
func buildMountArgs(project, workDirBase string, extraMounts []string) []string {
	projectDir := filepath.Join(workDirBase, project)

	args := []string{
		"-v", projectDir + "/.scrt-logs:/home/kali/.logs:rw",
		"-v", projectDir + ":/" + project,
		"-v", filepath.Join(projectDir, "certs") + ":/certs",
	}

	for _, mount := range extraMounts {
		mount = strings.TrimSpace(mount)
		if mount != "" {
			args = append(args, "-v", mount)
		}
	}

	return args
}

// deduplicateCaps removes duplicate Linux capabilities while preserving order.
func deduplicateCaps(caps []string) []string {
	seen := make(map[string]struct{}, len(caps))
	var out []string
	for _, c := range caps {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, exists := seen[c]; exists {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}
