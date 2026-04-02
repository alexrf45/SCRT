// Package backend provides a runtime-agnostic interface for container operations.
// Commands and the HTTP API depend only on Backend; the concrete implementation
// is selected at startup by New() based on what runtimes are available.
//
// Tier selection order:
//  1. DockerBackend       — Docker daemon reachable (current)
//  2. ContainerdBackend   — /run/containerd/containerd.sock present (k8s nodes, stub)
//  3. OCIBackend          — runc or crun in $PATH (bare metal, post-privesc, stub)
package backend

import (
	"context"
	"io"

	"github.com/alexrf45/scrt/internal/container"
)

// Backend abstracts container runtime operations across tiers.
// Accept this interface in commands and HTTP handlers; never accept
// a concrete type (API-2).
type Backend interface {
	Start(ctx context.Context, p container.RunParams) error
	Enter(ctx context.Context, project, shell string) error
	// StartExisting restarts a stopped container (docker start).
	// It is a no-op if the container is already running.
	StartExisting(ctx context.Context, project string) error
	Stop(ctx context.Context, project string) error
	Destroy(ctx context.Context, project string) error
	List(ctx context.Context) ([]container.Info, error)
	Pull(ctx context.Context, image string) error
	ImportBackup(ctx context.Context, p container.ImportParams) error
	WebExec(ctx context.Context, p container.WebExecParams) (*container.ExecSession, error)
	ResizeExec(ctx context.Context, p container.ResizeExecParams) error
	// CopyFrom copies a file or directory from the container as a tar stream.
	// The caller must close the returned ReadCloser.
	CopyFrom(ctx context.Context, project, srcPath string) (io.ReadCloser, error)
	// CopyTo copies a tar-encoded stream into the container at dstPath.
	CopyTo(ctx context.Context, project, dstPath string, content io.Reader) error
	Close() error
}
