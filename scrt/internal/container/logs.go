package container

import (
	"context"
	"fmt"
	"io"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// LogParams configures a container log stream. (CS-5)
type LogParams struct {
	Project    string
	Follow     bool   // stream new log lines as they arrive
	Tail       string // number of lines from the end, or "all" (default)
	Timestamps bool   // prefix each line with an RFC3339 timestamp
}

// Logs returns a plain-text stream of the container's logs. The caller must
// close the returned ReadCloser.
//
// Docker multiplexes stdout and stderr into a single framed stream unless the
// container was created with a TTY. scrt containers run with a TTY, so their
// logs arrive raw; for any non-TTY container the framed stream is demultiplexed
// transparently here, so callers always receive clean text.
func (m *Manager) Logs(ctx context.Context, p LogParams) (io.ReadCloser, error) {
	tail := p.Tail
	if tail == "" {
		tail = "all"
	}

	inspect, err := m.client.ContainerInspect(ctx, p.Project)
	if err != nil {
		return nil, fmt.Errorf("inspect container %s: %w", p.Project, err)
	}

	rc, err := m.client.ContainerLogs(ctx, p.Project, containertypes.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     p.Follow,
		Tail:       tail,
		Timestamps: p.Timestamps,
	})
	if err != nil {
		return nil, fmt.Errorf("container logs %s: %w", p.Project, err)
	}

	if inspect.Config != nil && inspect.Config.Tty {
		return rc, nil // raw stream, no multiplexing
	}
	return demuxLogs(rc), nil
}

// demuxLogs converts Docker's multiplexed stdout/stderr stream into a single
// plain-text stream. The returned reader must be closed by the caller; closing
// it also releases the source stream.
func demuxLogs(src io.ReadCloser) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		_, err := stdcopy.StdCopy(pw, pw, src)
		_ = src.Close()
		_ = pw.CloseWithError(err)
	}()
	return &demuxReadCloser{pr: pr, src: src}
}

// demuxReadCloser couples the pipe reader with the underlying source so that
// closing it stops the demux goroutine and frees the Docker stream.
type demuxReadCloser struct {
	pr  *io.PipeReader
	src io.ReadCloser
}

func (d *demuxReadCloser) Read(p []byte) (int, error) { return d.pr.Read(p) }

func (d *demuxReadCloser) Close() error {
	_ = d.src.Close()
	return d.pr.Close()
}
