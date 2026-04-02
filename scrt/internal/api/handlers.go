package api

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/alexrf45/scrt/internal/config"
	"github.com/alexrf45/scrt/internal/container"
	"github.com/alexrf45/scrt/internal/project"
)

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *Server) handleListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := s.p.Backend.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, containers)
}

func (s *Server) handleStartContainer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.p.Backend.StartExisting(r.Context(), name); err != nil {
		if errors.Is(err, container.ErrContainerNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStopContainer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.p.Backend.Stop(r.Context(), name); err != nil {
		if errors.Is(err, container.ErrContainerNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDestroyContainer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.p.Backend.Destroy(r.Context(), name); err != nil {
		if !errors.Is(err, container.ErrContainerNotFound) {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if project.Exists(s.p.Config.WorkDirBase, name) {
		if err := project.Remove(s.p.Config.WorkDirBase, name); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleBackupContainer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	backupFile, err := project.Backup(project.BackupParams{
		BaseDir:   s.p.Config.WorkDirBase,
		Project:   name,
		BackupDir: config.DefaultBackupDir,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"file": backupFile})
}

// pullRequest is the request body for POST /api/v1/images/pull.
type pullRequest struct {
	Image string `json:"image"`
}

func (s *Server) handlePullImage(w http.ResponseWriter, r *http.Request) {
	var req pullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Image == "" {
		http.Error(w, `body must be {"image":"<ref>"}`, http.StatusBadRequest)
		return
	}
	if err := s.p.Backend.Pull(r.Context(), req.Image); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDownloadFile copies a file or directory from a container as a tar archive.
// GET /api/v1/containers/{name}/files?path=/container/path
func (s *Server) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	srcPath := r.URL.Query().Get("path")
	if err := validateTransferPath(srcPath); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rc, err := s.p.Backend.CopyFrom(r.Context(), name, srcPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rc.Close()

	base := path.Base(srcPath)
	filename := name + "-" + base + ".tar"
	w.Header().Set("Content-Type", "application/x-tar")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = io.Copy(w, rc)
}

// handleUploadFile copies a file into a container from a multipart form upload.
// POST /api/v1/containers/{name}/files?path=/target/dir/
// Form field: "file" — the file to upload.
func (s *Server) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	dstPath := r.URL.Query().Get("path")
	if err := validateTransferPath(dstPath); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Limit upload to 100 MB.
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse form: %w", err))
		return
	}

	f, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read file field: %w", err))
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("read file data: %w", err))
		return
	}

	// Docker's CopyToContainer expects a tar archive.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{
		Name: filepath.Base(header.Filename),
		Mode: 0o644,
		Size: int64(len(data)),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("build tar header: %w", err))
		return
	}
	if _, err := tw.Write(data); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("write tar data: %w", err))
		return
	}
	tw.Close()

	if err := s.p.Backend.CopyTo(r.Context(), name, dstPath, &buf); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// validateTransferPath rejects empty paths and path traversal attempts.
func validateTransferPath(p string) error {
	if p == "" {
		return errors.New("path query parameter is required")
	}
	for _, part := range strings.Split(p, "/") {
		if part == ".." {
			return errors.New("path may not contain '..' components")
		}
	}
	return nil
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body.
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
