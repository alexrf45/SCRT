package backend

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	charmlog "github.com/charmbracelet/log"
	"github.com/alexrf45/scrt/internal/container"
)

// Compile-time assertion: *container.Manager must satisfy Backend.
// If the interface and Manager drift apart, this line fails to compile.
var _ Backend = (*container.Manager)(nil)

// ErrNoRuntime is returned when no supported container runtime is detected.
var ErrNoRuntime = errors.New("no supported container runtime found")

// containerdSockets are the well-known containerd socket paths across
// common Kubernetes distributions.
var containerdSockets = []string{
	"/run/containerd/containerd.sock",     // standard k8s, EKS, GKE, AKS
	"/run/k3s/containerd/containerd.sock", // k3s
}

// New detects the available container runtime and returns the appropriate Backend.
// Detection order: Docker daemon → containerd socket → runc/crun binary.
// Fails fast if no runtime is found (CFG-1).
func New(ctx context.Context, logger *charmlog.Logger) (Backend, error) {
	// Tier 1: Docker daemon
	if mgr, err := container.NewManager(ctx, logger); err == nil {
		logger.Debug("backend: Docker tier selected")
		return mgr, nil
	}

	// Tier 2: containerd socket (stub — not yet implemented)
	for _, p := range containerdSockets {
		if _, err := os.Stat(p); err == nil {
			logger.Debug("backend: containerd socket found; tier not yet implemented", "path", p)
			// TODO(feat/containerd): return newContainerdBackend(ctx, logger, p)
		}
	}

	// Tier 3: runc / crun binary (stub — not yet implemented)
	for _, bin := range []string{"runc", "crun"} {
		if _, err := exec.LookPath(bin); err == nil {
			logger.Debug("backend: OCI runtime found; tier not yet implemented", "binary", bin)
			// TODO(feat/oci): return newOCIBackend(ctx, logger, bin)
		}
	}

	return nil, fmt.Errorf("backend: tried Docker, containerd, runc/crun: %w", ErrNoRuntime)
}
