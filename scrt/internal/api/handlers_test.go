package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexrf45/scrt/internal/config"
	"github.com/alexrf45/scrt/internal/container"
)

func TestHandleHealthz(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, &mockBackend{})
	w := do(t, srv, http.MethodGet, "/healthz", nil, false)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); body != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}

func TestHandleUI(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, &mockBackend{})
	w := do(t, srv, http.MethodGet, "/", nil, false)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html prefix", ct)
	}
	if body := w.Body.String(); !strings.Contains(body, "SCRT") {
		t.Error("UI body does not contain expected content 'SCRT'")
	}
}

func TestHandleListContainers(t *testing.T) {
	t.Parallel()

	sampleContainers := []container.Info{
		{Name: "proj1", State: "running", Status: "Up 2 hours", Image: "fonalex45/scrt:latest"},
		{Name: "proj2", State: "exited", Status: "Exited (0)", Image: "fonalex45/scrt:dev"},
	}

	tests := []struct {
		name       string
		backend    *mockBackend
		authed     bool
		wantStatus int
		wantLen    int
	}{
		{
			name:       "unauthorized",
			backend:    &mockBackend{},
			authed:     false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "success with containers",
			backend: &mockBackend{
				listFn: func(_ context.Context) ([]container.Info, error) {
					return sampleContainers, nil
				},
			},
			authed:     true,
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name: "success empty slice",
			backend: &mockBackend{
				listFn: func(_ context.Context) ([]container.Info, error) {
					return []container.Info{}, nil
				},
			},
			authed:     true,
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name: "backend error",
			backend: &mockBackend{
				listFn: func(_ context.Context) ([]container.Info, error) {
					return nil, errors.New("daemon unreachable")
				},
			},
			authed:     true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer(t, tt.backend)
			w := do(t, srv, http.MethodGet, "/api/v1/containers", nil, tt.authed)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var got []container.Info
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("container count = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestHandleStopContainer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		project    string
		backend    *mockBackend
		authed     bool
		wantStatus int
	}{
		{
			name:       "unauthorized",
			project:    "myproj",
			backend:    &mockBackend{},
			authed:     false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid project name",
			project:    "bad.name",
			backend:    &mockBackend{},
			authed:     true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "success",
			project: "myproj",
			backend: &mockBackend{
				stopFn: func(_ context.Context, _ string) error { return nil },
			},
			authed:     true,
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "container not found",
			project: "myproj",
			backend: &mockBackend{
				stopFn: func(_ context.Context, _ string) error {
					return fmt.Errorf("stop: %w", container.ErrContainerNotFound)
				},
			},
			authed:     true,
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "backend error",
			project: "myproj",
			backend: &mockBackend{
				stopFn: func(_ context.Context, _ string) error {
					return errors.New("connection refused")
				},
			},
			authed:     true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer(t, tt.backend)
			w := do(t, srv, http.MethodPost, "/api/v1/containers/"+tt.project+"/stop", nil, tt.authed)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandleDestroyContainer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		project    string
		setupDir   bool
		backend    *mockBackend
		authed     bool
		wantStatus int
	}{
		{
			name:       "unauthorized",
			project:    "myproj",
			backend:    &mockBackend{},
			authed:     false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid project name",
			project:    "bad.name",
			backend:    &mockBackend{},
			authed:     true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "success without project dir",
			project: "myproj",
			backend: &mockBackend{
				destroyFn: func(_ context.Context, _ string) error { return nil },
			},
			authed:     true,
			setupDir:   false,
			wantStatus: http.StatusNoContent,
		},
		{
			name:     "success removes project dir",
			project:  "myproj",
			setupDir: true,
			backend: &mockBackend{
				destroyFn: func(_ context.Context, _ string) error { return nil },
			},
			authed:     true,
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "container not found still succeeds",
			project: "myproj",
			backend: &mockBackend{
				destroyFn: func(_ context.Context, _ string) error {
					return fmt.Errorf("destroy: %w", container.ErrContainerNotFound)
				},
			},
			authed:     true,
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "backend error",
			project: "myproj",
			backend: &mockBackend{
				destroyFn: func(_ context.Context, _ string) error {
					return errors.New("daemon unreachable")
				},
			},
			authed:     true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workDir := t.TempDir()
			if tt.setupDir {
				if err := os.MkdirAll(filepath.Join(workDir, tt.project), 0o755); err != nil {
					t.Fatalf("setup project dir: %v", err)
				}
			}

			srv := newTestServer(t, tt.backend, config.Config{WorkDirBase: workDir})
			w := do(t, srv, http.MethodPost, "/api/v1/containers/"+tt.project+"/destroy", nil, tt.authed)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.setupDir && tt.wantStatus == http.StatusNoContent {
				if _, err := os.Stat(filepath.Join(workDir, tt.project)); !os.IsNotExist(err) {
					t.Error("expected project directory to be removed after destroy")
				}
			}
		})
	}
}

func TestHandleBackupContainer(t *testing.T) {
	// Cannot run in parallel: uses t.Chdir which changes process-level working
	// directory (config.DefaultBackupDir = "./backups" is relative to CWD).

	tests := []struct {
		name       string
		project    string
		setupDir   bool
		wantStatus int
	}{
		{
			name:       "invalid project name",
			project:    "bad.name",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "project directory missing",
			project:    "missing",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "success",
			project:    "myproj",
			setupDir:   true,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect ./backups into an isolated temp dir per sub-test.
			t.Chdir(t.TempDir())

			workDir := t.TempDir()
			if tt.setupDir {
				if err := os.MkdirAll(filepath.Join(workDir, tt.project), 0o755); err != nil {
					t.Fatalf("setup project dir: %v", err)
				}
			}

			srv := newTestServer(t, &mockBackend{}, config.Config{WorkDirBase: workDir})
			w := do(t, srv, http.MethodPost, "/api/v1/containers/"+tt.project+"/backup", nil, true)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK {
				var result map[string]string
				if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
					t.Fatalf("decode response body: %v", err)
				}
				if result["file"] == "" {
					t.Error("expected non-empty 'file' field in backup response")
				}
			}
		})
	}
}

func TestHandlePullImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		backend    *mockBackend
		authed     bool
		wantStatus int
	}{
		{
			name:       "unauthorized",
			body:       `{"image":"fonalex45/scrt:latest"}`,
			backend:    &mockBackend{},
			authed:     false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "success",
			body: `{"image":"fonalex45/scrt:latest"}`,
			backend: &mockBackend{
				pullFn: func(_ context.Context, _ string) error { return nil },
			},
			authed:     true,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "empty image field",
			body:       `{"image":""}`,
			backend:    &mockBackend{},
			authed:     true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing image field",
			body:       `{}`,
			backend:    &mockBackend{},
			authed:     true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed json",
			body:       `not json`,
			backend:    &mockBackend{},
			authed:     true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "backend error",
			body: `{"image":"fonalex45/scrt:latest"}`,
			backend: &mockBackend{
				pullFn: func(_ context.Context, _ string) error {
					return errors.New("registry unreachable")
				},
			},
			authed:     true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer(t, tt.backend)
			w := do(t, srv, http.MethodPost, "/api/v1/images/pull", strings.NewReader(tt.body), tt.authed)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
