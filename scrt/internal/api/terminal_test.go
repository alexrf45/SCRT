package api

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexrf45/scrt/internal/container"
	"github.com/gorilla/websocket"
)

// TestHandleTerminalAuth verifies auth and name-validation checks fire before
// the WebSocket upgrade, so a plain HTTP recorder is sufficient.
func TestHandleTerminalAuth(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, &mockBackend{})

	tests := []struct {
		name       string
		token      string
		project    string
		wantStatus int
	}{
		{"no token", "", "myproj", http.StatusUnauthorized},
		{"wrong token", "bad-token", "myproj", http.StatusUnauthorized},
		{"invalid project name", testToken, "bad.name", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/ws/exec/" + tt.project
			if tt.token != "" {
				path += "?token=" + tt.token
			}
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			srv.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestHandleTerminalBackendError verifies that a WebExec failure causes the
// server to close the WebSocket with CloseInternalServerErr (1011).
func TestHandleTerminalBackendError(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, &mockBackend{
		webExecFn: func(_ context.Context, _ container.WebExecParams) (*container.ExecSession, error) {
			return nil, errors.New("container not running")
		},
	})
	ts := httptest.NewServer(srv.mux)
	t.Cleanup(ts.Close)

	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/exec/myproj?token=" + testToken
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Fatal("expected connection to be closed after backend error")
	}
	closeErr, ok := err.(*websocket.CloseError)
	if !ok {
		t.Fatalf("expected *websocket.CloseError, got %T: %v", err, err)
	}
	if closeErr.Code != websocket.CloseInternalServerErr {
		t.Errorf("close code = %d, want %d (CloseInternalServerErr)", closeErr.Code, websocket.CloseInternalServerErr)
	}
}

// TestHandleTerminalDataBridge verifies the three data paths through the terminal handler:
//  1. Docker PTY output flows to the browser as binary WebSocket frames.
//  2. Browser keystrokes (binary frames) are forwarded to the Docker exec conn.
//  3. Resize text frames are parsed and forwarded to Backend.ResizeExec.
func TestHandleTerminalDataBridge(t *testing.T) {
	t.Parallel()

	// net.Pipe simulates the Docker hijacked exec connection.
	// dockerClient = test-controlled side; dockerServer = handler side.
	dockerClient, dockerServer := net.Pipe()

	resizeDone := make(chan container.ResizeExecParams, 1)

	srv := newTestServer(t, &mockBackend{
		webExecFn: func(_ context.Context, _ container.WebExecParams) (*container.ExecSession, error) {
			return &container.ExecSession{
				ExecID: "test-exec-id",
				Conn:   dockerServer,
				Reader: dockerServer,
			}, nil
		},
		resizeExecFn: func(_ context.Context, p container.ResizeExecParams) error {
			select {
			case resizeDone <- p:
			default:
			}
			return nil
		},
	})

	ts := httptest.NewServer(srv.mux)
	// Close the HTTP server before cleaning up the pipe so handler goroutines
	// have exited before the underlying net.Conn is torn down.
	t.Cleanup(func() { dockerClient.Close(); dockerServer.Close() })
	t.Cleanup(ts.Close)

	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/exec/myproj?token=" + testToken
	wsConn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer wsConn.Close()

	// ── Docker → browser ──────────────────────────────────────────────────────
	const ptyOutput = "prompt$ "
	go func() { _, _ = io.WriteString(dockerClient, ptyOutput) }()

	if err := wsConn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	mt, got, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read WebSocket message: %v", err)
	}
	if mt != websocket.BinaryMessage {
		t.Errorf("message type = %d, want BinaryMessage (%d)", mt, websocket.BinaryMessage)
	}
	if string(got) != ptyOutput {
		t.Errorf("PTY output = %q, want %q", got, ptyOutput)
	}

	// ── Browser → Docker ──────────────────────────────────────────────────────
	const keystroke = "ls\r"
	if err := wsConn.WriteMessage(websocket.BinaryMessage, []byte(keystroke)); err != nil {
		t.Fatalf("send keystroke: %v", err)
	}

	if err := dockerClient.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set docker read deadline: %v", err)
	}
	recv := make([]byte, len(keystroke))
	if _, err := io.ReadFull(dockerClient, recv); err != nil {
		t.Fatalf("read from docker side: %v", err)
	}
	if string(recv) != keystroke {
		t.Errorf("docker received %q, want %q", recv, keystroke)
	}

	// ── Resize frame ──────────────────────────────────────────────────────────
	resizeJSON := `{"type":"resize","rows":24,"cols":80}`
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte(resizeJSON)); err != nil {
		t.Fatalf("send resize frame: %v", err)
	}

	select {
	case p := <-resizeDone:
		if p.Rows != 24 || p.Cols != 80 {
			t.Errorf("resize params = {Rows:%d, Cols:%d}, want {Rows:24, Cols:80}", p.Rows, p.Cols)
		}
	case <-time.After(3 * time.Second):
		t.Error("resize message not forwarded to backend within timeout")
	}
}
