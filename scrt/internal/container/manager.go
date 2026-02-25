package container

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// Info holds displayable container metadata.
type Info struct {
	Name      string `json:"Names"`
	Status    string `json:"Status"`
	State     string `json:"State"`
	Image     string `json:"Image"`
	CreatedAt string `json:"CreatedAt"`
}

// Manager orchestrates Docker container lifecycle operations for SCRT.
// All operations shell out to the docker CLI via os/exec.
type Manager struct {
	logger *slog.Logger
}

// NewManager creates a Manager and verifies Docker is available.
// Returns ErrDockerUnavailable if the daemon is not reachable.
func NewManager(ctx context.Context, logger *slog.Logger) (*Manager, error) {
	cmd := exec.CommandContext(ctx, "docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDockerUnavailable, err)
	}

	return &Manager{logger: logger}, nil
}

// Close is a no-op for the exec-based manager. Exists for interface compatibility.
func (m *Manager) Close() error {
	return nil
}

// Start creates and starts a new interactive container.
// The caller should have already created the project directory structure.
func (m *Manager) Start(ctx context.Context, params RunParams) error {
	exists, running := m.containerState(ctx, params.Project)
	if exists {
		if running {
			m.logger.Info("container is running — use 'enter' to access it", "project", params.Project)
		} else {
			m.logger.Info("container is stopped — use 'enter' to start and access it", "project", params.Project)
		}
		return fmt.Errorf("%w: %s", ErrContainerExists, params.Project)
	}

	args, err := BuildRunArgs(params)
	if err != nil {
		return fmt.Errorf("build run args: %w", err)
	}

	fullArgs := append([]string{"run"}, args...)
	m.logger.Debug("docker run command", "args", fullArgs)

	// docker run -it needs direct access to the terminal
	cmd := exec.CommandContext(ctx, "docker", fullArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run %s: %w", params.Project, err)
	}

	return nil
}

// Enter opens an interactive shell in an existing container.
// If the container is stopped, it is started first.
func (m *Manager) Enter(ctx context.Context, project, shell string) error {
	exists, running := m.containerState(ctx, project)
	if !exists {
		return fmt.Errorf("%w: %s", ErrContainerNotFound, project)
	}

	if !running {
		m.logger.Info("starting stopped container", "project", project)
		startCmd := exec.CommandContext(ctx, "docker", "container", "start", project)
		if out, err := startCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("start container %s: %s: %w", project, strings.TrimSpace(string(out)), err)
		}
	}

	m.logger.Info("entering container", "project", project)

	cmd := exec.CommandContext(ctx, "docker", "exec", "-it", project, shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec into %s: %w", project, err)
	}

	return nil
}

// Stop stops a running container.
func (m *Manager) Stop(ctx context.Context, project string) error {
	exists, running := m.containerState(ctx, project)
	if !exists {
		return fmt.Errorf("%w: %s", ErrContainerNotFound, project)
	}

	if !running {
		m.logger.Info("container is already stopped", "project", project)
		return nil
	}

	m.logger.Info("stopping container", "project", project)

	cmd := exec.CommandContext(ctx, "docker", "container", "stop", project)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("stop container %s: %s: %w", project, strings.TrimSpace(string(out)), err)
	}

	m.logger.Info("container stopped", "project", project)
	return nil
}

// Destroy removes a container and optionally its data.
// The caller is responsible for removing the project directory.
func (m *Manager) Destroy(ctx context.Context, project string) error {
	exists, _ := m.containerState(ctx, project)
	if !exists {
		return fmt.Errorf("%w: %s", ErrContainerNotFound, project)
	}

	m.logger.Info("destroying container", "project", project)

	cmd := exec.CommandContext(ctx, "docker", "container", "rm", "-f", project)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("remove container %s: %s: %w", project, strings.TrimSpace(string(out)), err)
	}

	m.logger.Info("container destroyed", "project", project)
	return nil
}

// List returns all SCRT containers (running and stopped), filtered by label.
func (m *Manager) List(ctx context.Context) ([]Info, error) {
	cmd := exec.CommandContext(ctx,
		"docker", "ps", "-a",
		"--filter", "label="+FilterLabels(),
		"--format", "json",
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	return parseContainerList(string(out))
}

// Pull downloads a Docker image from the registry.
func (m *Manager) Pull(ctx context.Context, image string) error {
	m.logger.Info("pulling image", "image", image)

	cmd := exec.CommandContext(ctx, "docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pull image %s: %w", image, err)
	}

	m.logger.Info("image pulled", "image", image)
	return nil
}

// containerState checks whether a container exists and whether it's running.
func (m *Manager) containerState(ctx context.Context, name string) (exists bool, running bool) {
	cmd := exec.CommandContext(ctx,
		"docker", "container", "inspect",
		"-f", "{{.State.Running}}",
		name,
	)

	out, err := cmd.Output()
	if err != nil {
		return false, false
	}

	return true, strings.TrimSpace(string(out)) == "true"
}

// parseContainerList parses newline-delimited JSON from docker ps --format json.
func parseContainerList(output string) ([]Info, error) {
	var containers []Info

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}

		var info Info
		if err := json.Unmarshal([]byte(line), &info); err != nil {
			continue // skip malformed lines
		}
		containers = append(containers, info)
	}

	return containers, nil
}
