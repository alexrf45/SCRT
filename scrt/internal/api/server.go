// Package api implements the SCRT HTTP API and embedded web UI.
// It is consumed by the `scrt serve` command and is runtime-agnostic —
// all container operations are delegated to a backend.Backend.
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alexrf45/scrt/internal/backend"
	"github.com/alexrf45/scrt/internal/config"
)

// ServeParams configures the HTTP server. (CS-5)
type ServeParams struct {
	Addr     string
	Token    string
	Backend  backend.Backend
	Config   config.Config
	CertFile string // optional PEM certificate; enables TLS when set with KeyFile
	KeyFile  string // optional PEM private key; enables TLS when set with CertFile
}

// Server serves the SCRT HTTP API and embedded web UI.
type Server struct {
	p   ServeParams
	mux *http.ServeMux
}

// New creates a Server and registers all routes.
func New(p ServeParams) *Server {
	s := &Server{p: p, mux: http.NewServeMux()}
	s.registerRoutes()
	return s
}

// Start begins serving on the configured address and blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.p.Addr,
		Handler:           s.mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		var err error
		if s.p.CertFile != "" && s.p.KeyFile != "" {
			err = srv.ListenAndServeTLS(s.p.CertFile, s.p.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen %s: %w", s.p.Addr, err)
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return <-errCh
	case err := <-errCh:
		return err
	}
}

func (s *Server) registerRoutes() {
	auth := func(h http.HandlerFunc) http.Handler { return s.requireToken(h) }

	// Unauthenticated — web UI and health probe
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /", s.handleUI)

	// Authenticated API
	s.mux.Handle("GET /api/v1/containers", auth(s.handleListContainers))
	s.mux.Handle("POST /api/v1/containers/{name}/start", auth(s.handleStartContainer))
	s.mux.Handle("POST /api/v1/containers/{name}/stop", auth(s.handleStopContainer))
	s.mux.Handle("POST /api/v1/containers/{name}/destroy", auth(s.handleDestroyContainer))
	s.mux.Handle("POST /api/v1/containers/{name}/backup", auth(s.handleBackupContainer))
	s.mux.Handle("GET /api/v1/containers/{name}/files", auth(s.handleDownloadFile))
	s.mux.Handle("POST /api/v1/containers/{name}/files", auth(s.handleUploadFile))
	s.mux.Handle("POST /api/v1/images/pull", auth(s.handlePullImage))

	// WebSocket terminal — auth via ?token= query param (browsers cannot set
	// custom headers on WebSocket upgrades).
	s.mux.HandleFunc("GET /ws/exec/{name}", s.handleTerminal)
}
