package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"

	charmlog "github.com/charmbracelet/log"
)

// ExecSession holds an active TTY exec session into a container.
// The caller must close Conn when the session is complete.
type ExecSession struct {
	ExecID string
	// Conn accepts stdin writes for the exec process.
	Conn net.Conn
	// Reader delivers raw PTY output (stdout+stderr combined when Tty=true).
	Reader io.Reader
}

// WebExecParams holds parameters for WebExec. (CS-5)
type WebExecParams struct {
	Project string
	Shell   string
}

// ResizeExecParams holds parameters for ResizeExec. (CS-5)
type ResizeExecParams struct {
	ExecID string
	Rows   uint
	Cols   uint
}

// Info holds displayable container metadata.
type Info struct {
	Name      string `json:"Name"`
	Status    string `json:"Status"`
	State     string `json:"State"`
	Image     string `json:"Image"`
	CreatedAt string `json:"CreatedAt"`
}

// ImportParams holds the parameters for importing a backup tar archive.
type ImportParams struct {
	File string // path to .tar file
	Repo string // repository name (e.g. fonalex45/scrt)
	Tag  string // tag (e.g. imported)
}

// Manager orchestrates Docker container lifecycle operations for SCRT.
// Non-interactive operations use the Docker Go SDK; TTY-attached sessions
// (docker run -it, docker exec -it) shell out to the docker CLI.
type Manager struct {
	logger *charmlog.Logger
	client *dockerclient.Client
}

// NewManager creates a Manager and verifies Docker is available via the SDK.
// Returns ErrDockerUnavailable if the daemon is not reachable.
func NewManager(ctx context.Context, logger *charmlog.Logger) (*Manager, error) {
	c, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDockerUnavailable, err)
	}

	if _, err := c.ServerVersion(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("%w: %w", ErrDockerUnavailable, err)
	}

	return &Manager{logger: logger, client: c}, nil
}

// Close releases the Docker client connection.
func (m *Manager) Close() error {
	return m.client.Close()
}

// Start creates and starts a new interactive container via docker CLI (TTY required).
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

	// docker run -it requires direct terminal access; use CLI exec.
	cmd := exec.CommandContext(ctx, "docker", fullArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run %s: %w", params.Project, err)
	}

	return nil
}

// Enter opens an interactive shell in an existing container via docker CLI (TTY required).
// If the container is stopped, it is started first via the Docker SDK.
func (m *Manager) Enter(ctx context.Context, project, shell string) error {
	exists, running := m.containerState(ctx, project)
	if !exists {
		return fmt.Errorf("%w: %s", ErrContainerNotFound, project)
	}

	if !running {
		m.logger.Info("starting stopped container", "project", project)
		if err := m.client.ContainerStart(ctx, project, containertypes.StartOptions{}); err != nil {
			return fmt.Errorf("start container %s: %w", project, err)
		}
	}

	// docker exec -it requires direct terminal access; use CLI exec.
	cmd := exec.CommandContext(ctx, "docker", "exec", "-it", project, shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec into %s: %w", project, err)
	}

	return nil
}

// StartExisting restarts a stopped container via the Docker SDK.
// It is a no-op if the container is already running.
func (m *Manager) StartExisting(ctx context.Context, project string) error {
	exists, running := m.containerState(ctx, project)
	if !exists {
		return fmt.Errorf("%w: %s", ErrContainerNotFound, project)
	}
	if running {
		return nil // already running; not an error
	}
	if err := m.client.ContainerStart(ctx, project, containertypes.StartOptions{}); err != nil {
		return fmt.Errorf("start container %s: %w", project, err)
	}
	return nil
}

// CopyFrom copies a file or directory from the container as a tar stream.
// The caller must close the returned ReadCloser.
func (m *Manager) CopyFrom(ctx context.Context, project, srcPath string) (io.ReadCloser, error) {
	reader, _, err := m.client.CopyFromContainer(ctx, project, srcPath)
	if err != nil {
		return nil, fmt.Errorf("copy from container %s:%s: %w", project, srcPath, err)
	}
	return reader, nil
}

// CopyTo copies a tar-encoded stream into the container at dstPath.
// The content reader must contain a valid tar archive; the Docker daemon
// extracts it into dstPath inside the container.
func (m *Manager) CopyTo(ctx context.Context, project, dstPath string, content io.Reader) error {
	if err := m.client.CopyToContainer(ctx, project, dstPath, content, containertypes.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("copy to container %s:%s: %w", project, dstPath, err)
	}
	return nil
}

// Stop stops a running container via the Docker SDK.
func (m *Manager) Stop(ctx context.Context, project string) error {
	exists, running := m.containerState(ctx, project)
	if !exists {
		return fmt.Errorf("%w: %s", ErrContainerNotFound, project)
	}

	if !running {
		return nil // already stopped; not an error
	}

	if err := m.client.ContainerStop(ctx, project, containertypes.StopOptions{}); err != nil {
		return fmt.Errorf("stop container %s: %w", project, err)
	}

	return nil
}

// Destroy removes a container forcefully via the Docker SDK.
// The caller is responsible for removing the project directory.
func (m *Manager) Destroy(ctx context.Context, project string) error {
	exists, _ := m.containerState(ctx, project)
	if !exists {
		return fmt.Errorf("%w: %s", ErrContainerNotFound, project)
	}

	if err := m.client.ContainerRemove(ctx, project, containertypes.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("remove container %s: %w", project, err)
	}

	return nil
}

// List returns all SCRT containers (running and stopped), filtered by label.
func (m *Manager) List(ctx context.Context) ([]Info, error) {
	filterArgs := filters.NewArgs(
		filters.Arg("label", FilterLabels()),
	)

	containers, err := m.client.ContainerList(ctx, containertypes.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	infos := make([]Info, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		infos = append(infos, Info{
			Name:   name,
			State:  c.State,
			Status: c.Status,
			Image:  c.Image,
		})
	}

	return infos, nil
}

// Pull downloads a Docker image from the registry via the Docker SDK.
// The response body is drained to ensure the pull completes; callers are
// responsible for progress display (e.g. via tui.RunWithSpinner).
func (m *Manager) Pull(ctx context.Context, image string) error {
	reader, err := m.client.ImagePull(ctx, image, imagetypes.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", image, err)
	}
	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("pull %s: %w", image, err)
	}
	return nil
}

// ImportBackup imports a tar archive as a Docker image via the Docker SDK.
func (m *Manager) ImportBackup(ctx context.Context, p ImportParams) error {
	f, err := os.Open(p.File)
	if err != nil {
		return fmt.Errorf("open backup file %s: %w", p.File, err)
	}
	defer f.Close()

	ref := p.Repo
	if p.Tag != "" {
		ref = ref + ":" + p.Tag
	}

	source := imagetypes.ImportSource{
		Source:     f,
		SourceName: "-",
	}

	resp, err := m.client.ImageImport(ctx, source, ref, imagetypes.ImportOptions{})
	if err != nil {
		return fmt.Errorf("import image from %s: %w", p.File, err)
	}
	defer resp.Close()

	if _, err := io.Copy(io.Discard, resp); err != nil {
		return fmt.Errorf("import %s: %w", p.File, err)
	}
	return nil
}

// WebExec creates a TTY exec session in a running container using the Docker SDK.
// The returned ExecSession.Conn must be closed by the caller when done.
func (m *Manager) WebExec(ctx context.Context, p WebExecParams) (*ExecSession, error) {
	exists, running := m.containerState(ctx, p.Project)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrContainerNotFound, p.Project)
	}
	if !running {
		return nil, fmt.Errorf("container %s is not running", p.Project)
	}

	execResp, err := m.client.ContainerExecCreate(ctx, p.Project, containertypes.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{p.Shell},
	})
	if err != nil {
		return nil, fmt.Errorf("exec create %s: %w", p.Project, err)
	}

	hijack, err := m.client.ContainerExecAttach(ctx, execResp.ID, containertypes.ExecAttachOptions{Tty: true})
	if err != nil {
		return nil, fmt.Errorf("exec attach %s: %w", p.Project, err)
	}

	return &ExecSession{
		ExecID: execResp.ID,
		Conn:   hijack.Conn,
		Reader: hijack.Reader,
	}, nil
}

// ResizeExec resizes the TTY of an active exec session.
func (m *Manager) ResizeExec(ctx context.Context, p ResizeExecParams) error {
	if err := m.client.ContainerExecResize(ctx, p.ExecID, containertypes.ResizeOptions{
		Height: p.Rows,
		Width:  p.Cols,
	}); err != nil {
		return fmt.Errorf("resize exec %s: %w", p.ExecID, err)
	}
	return nil
}

// containerState checks whether a container exists and whether it's running
// using the Docker SDK.
func (m *Manager) containerState(ctx context.Context, name string) (exists bool, running bool) {
	info, err := m.client.ContainerInspect(ctx, name)
	if err != nil {
		return false, false
	}
	return true, info.State.Running
}

// parseContainerList parses newline-delimited JSON from docker ps --format json.
// Retained for unit testing.
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
