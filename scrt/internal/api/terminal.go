package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alexrf45/scrt/internal/container"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	// Origin check is omitted; token auth below enforces access control.
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// handleTerminal upgrades an HTTP GET to a WebSocket and bridges it to a Docker
// exec session with a PTY. The browser passes the token via ?token= because the
// browser WebSocket API cannot set custom request headers during the upgrade.
func (s *Server) handleTerminal(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	if tok == "" || !equalConstantTime(tok, s.p.Token) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade writes its own HTTP error on failure.
		return
	}
	defer ws.Close()

	session, err := s.p.Backend.WebExec(r.Context(), container.WebExecParams{
		Project: name,
		Shell:   s.p.Config.ContainerShell,
	})
	if err != nil {
		_ = ws.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
		return
	}
	defer session.Conn.Close()

	done := make(chan struct{})

	// docker → browser: stream PTY output as binary WebSocket frames.
	go func() {
		defer close(done)
		defer ws.Close()
		buf := make([]byte, 32*1024)
		for {
			n, readErr := session.Reader.Read(buf)
			if n > 0 {
				if writeErr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()

	// browser → docker: forward keyboard input; handle resize control messages.
	go func() {
		defer session.Conn.Close()
		for {
			mt, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}
			switch mt {
			case websocket.BinaryMessage:
				if _, err := session.Conn.Write(msg); err != nil {
					return
				}
			case websocket.TextMessage:
				handleTerminalResize(r.Context(), s, session.ExecID, msg)
			}
		}
	}()

	<-done
}

// terminalResizeMsg is the JSON shape sent by the browser on terminal resize.
type terminalResizeMsg struct {
	Type string `json:"type"`
	Rows uint   `json:"rows"`
	Cols uint   `json:"cols"`
}

func handleTerminalResize(ctx context.Context, s *Server, execID string, data []byte) {
	var msg terminalResizeMsg
	if err := json.Unmarshal(data, &msg); err != nil || msg.Type != "resize" || msg.Rows == 0 || msg.Cols == 0 {
		return
	}
	_ = s.p.Backend.ResizeExec(ctx, container.ResizeExecParams{
		ExecID: execID,
		Rows:   msg.Rows,
		Cols:   msg.Cols,
	})
}
