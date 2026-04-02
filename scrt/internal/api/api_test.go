package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexrf45/scrt/internal/backend"
	"github.com/alexrf45/scrt/internal/config"
	"github.com/alexrf45/scrt/internal/container"
)


const testToken = "test-token-abc123xyz"

// Compile-time assertion: mockBackend must satisfy backend.Backend.
var _ backend.Backend = (*mockBackend)(nil)

// mockBackend implements backend.Backend for testing.
// Any nil field returns its zero value without error.
type mockBackend struct {
	startFn            func(context.Context, container.RunParams) error
	enterFn            func(context.Context, string, string) error
	startExistingFn    func(context.Context, string) error
	stopFn             func(context.Context, string) error
	destroyFn          func(context.Context, string) error
	listFn             func(context.Context) ([]container.Info, error)
	pullFn             func(context.Context, string) error
	importBackupFn     func(context.Context, container.ImportParams) error
	webExecFn          func(context.Context, container.WebExecParams) (*container.ExecSession, error)
	resizeExecFn       func(context.Context, container.ResizeExecParams) error
	copyFromFn         func(context.Context, string, string) (io.ReadCloser, error)
	copyToFn           func(context.Context, string, string, io.Reader) error
	execBackgroundFn   func(context.Context, container.BackgroundExecParams) ([]byte, int, error)
	closeFn            func() error
}

func (m *mockBackend) Start(ctx context.Context, p container.RunParams) error {
	if m.startFn != nil {
		return m.startFn(ctx, p)
	}
	return nil
}

func (m *mockBackend) Enter(ctx context.Context, project, shell string) error {
	if m.enterFn != nil {
		return m.enterFn(ctx, project, shell)
	}
	return nil
}

func (m *mockBackend) StartExisting(ctx context.Context, project string) error {
	if m.startExistingFn != nil {
		return m.startExistingFn(ctx, project)
	}
	return nil
}

func (m *mockBackend) CopyFrom(ctx context.Context, project, srcPath string) (io.ReadCloser, error) {
	if m.copyFromFn != nil {
		return m.copyFromFn(ctx, project, srcPath)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *mockBackend) CopyTo(ctx context.Context, project, dstPath string, content io.Reader) error {
	if m.copyToFn != nil {
		return m.copyToFn(ctx, project, dstPath, content)
	}
	return nil
}

func (m *mockBackend) ExecBackground(ctx context.Context, p container.BackgroundExecParams) ([]byte, int, error) {
	if m.execBackgroundFn != nil {
		return m.execBackgroundFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockBackend) Stop(ctx context.Context, project string) error {
	if m.stopFn != nil {
		return m.stopFn(ctx, project)
	}
	return nil
}

func (m *mockBackend) Destroy(ctx context.Context, project string) error {
	if m.destroyFn != nil {
		return m.destroyFn(ctx, project)
	}
	return nil
}

func (m *mockBackend) List(ctx context.Context) ([]container.Info, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockBackend) Pull(ctx context.Context, image string) error {
	if m.pullFn != nil {
		return m.pullFn(ctx, image)
	}
	return nil
}

func (m *mockBackend) ImportBackup(ctx context.Context, p container.ImportParams) error {
	if m.importBackupFn != nil {
		return m.importBackupFn(ctx, p)
	}
	return nil
}

func (m *mockBackend) WebExec(ctx context.Context, p container.WebExecParams) (*container.ExecSession, error) {
	if m.webExecFn != nil {
		return m.webExecFn(ctx, p)
	}
	return nil, nil
}

func (m *mockBackend) ResizeExec(ctx context.Context, p container.ResizeExecParams) error {
	if m.resizeExecFn != nil {
		return m.resizeExecFn(ctx, p)
	}
	return nil
}

func (m *mockBackend) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

// newTestServer creates a Server wired with the given backend.
// cfg is optional; if omitted a default with t.TempDir() as WorkDirBase is used.
func newTestServer(t *testing.T, b backend.Backend, cfg ...config.Config) *Server {
	t.Helper()
	c := config.Config{WorkDirBase: t.TempDir()}
	if len(cfg) > 0 {
		c = cfg[0]
	}
	return New(ServeParams{
		Addr:    ":0",
		Token:   testToken,
		Backend: b,
		Config:  c,
	})
}

// do issues method+path through srv.mux and returns the recorded response.
// Sets Authorization header when authed is true.
// Sets Content-Type: application/json when body is non-nil.
func do(t *testing.T, srv *Server, method, path string, body io.Reader, authed bool) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if authed {
		req.Header.Set("Authorization", "Bearer "+testToken)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Routing integration tests
// ---------------------------------------------------------------------------

func TestUnauthenticatedRoutes(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, &mockBackend{})

	tests := []struct {
		path string
	}{
		{"/healthz"},
		{"/"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			w := do(t, srv, http.MethodGet, tt.path, nil, false)
			if w.Code != http.StatusOK {
				t.Errorf("GET %s (no auth) = %d, want 200", tt.path, w.Code)
			}
		})
	}
}

func TestAPIRoutesRequireAuth(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, &mockBackend{
		listFn:    func(_ context.Context) ([]container.Info, error) { return nil, nil },
		stopFn:    func(_ context.Context, _ string) error { return nil },
		destroyFn: func(_ context.Context, _ string) error { return nil },
		pullFn:    func(_ context.Context, _ string) error { return nil },
	})

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/containers", ""},
		{http.MethodPost, "/api/v1/containers/proj/start", ""},
		{http.MethodPost, "/api/v1/containers/proj/stop", ""},
		{http.MethodPost, "/api/v1/containers/proj/destroy", ""},
		{http.MethodPost, "/api/v1/containers/proj/backup", ""},
		{http.MethodGet, "/api/v1/containers/proj/files?path=/tmp/flag.txt", ""},
		{http.MethodPost, "/api/v1/containers/proj/files?path=/tmp/", ""},
		{http.MethodPost, "/api/v1/containers/proj/exec", `{"cmd":"id"}`},
		{http.MethodGet, "/api/v1/jobs", ""},
		{http.MethodGet, "/api/v1/jobs/abc123", ""},
		{http.MethodDelete, "/api/v1/jobs/abc123", ""},
		{http.MethodPost, "/api/v1/images/pull", `{"image":"test:latest"}`},
	}

	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			t.Parallel()
			var body io.Reader
			if r.body != "" {
				body = strings.NewReader(r.body)
			}
			w := do(t, srv, r.method, r.path, body, false)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s %s without auth: got %d, want 401", r.method, r.path, w.Code)
			}
			if w.Header().Get("WWW-Authenticate") == "" {
				t.Errorf("%s %s: missing WWW-Authenticate header on 401", r.method, r.path)
			}
		})
	}
}
