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

	"github.com/alexrf45/scrt/internal/container"
)

// Backend abstracts container runtime operations across tiers.
// Accept this interface in commands and HTTP handlers; never accept
// a concrete type (API-2).
type Backend interface {
	Start(ctx context.Context, p container.RunParams) error
	Enter(ctx context.Context, project, shell string) error
	Stop(ctx context.Context, project string) error
	Destroy(ctx context.Context, project string) error
	List(ctx context.Context) ([]container.Info, error)
	Pull(ctx context.Context, image string) error
	ImportBackup(ctx context.Context, p container.ImportParams) error
	Close() error
}
