package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
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
	if err := container.ValidateTransferPath(srcPath); err != nil {
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
	if err := container.ValidateTransferPath(dstPath); err != nil {
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
	buf, err := container.TarFile(header.Filename, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := s.p.Backend.CopyTo(r.Context(), name, dstPath, buf); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// execRequest is the body for POST /api/v1/containers/{name}/exec.
type execRequest struct {
	Cmd string `json:"cmd"`
}

// handleExecBackground starts a background exec job inside a container.
// POST /api/v1/containers/{name}/exec   body: {"cmd":"<shell command>"}
// Returns 202 Accepted with {"id":"<job-id>"}.
func (s *Server) handleExecBackground(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := container.ValidateProjectName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req execRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Cmd) == "" {
		http.Error(w, `body must be {"cmd":"<command>"}`, http.StatusBadRequest)
		return
	}

	// Execute via sh -c so operators get full shell features (pipes, redirects, &&).
	cmd := []string{"sh", "-c", req.Cmd}

	jobCtx, cancel := context.WithCancel(context.Background())
	job := s.jobs.create(name, req.Cmd, cancel)

	go func() {
		output, exitCode, execErr := s.p.Backend.ExecBackground(jobCtx, container.BackgroundExecParams{
			Project: name,
			Cmd:     cmd,
		})
		s.jobs.finish(job.ID, output, exitCode, execErr)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": job.ID})
}

// handleListJobs returns all background exec jobs (summary, no output).
// GET /api/v1/jobs
func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.jobs.list())
}

// handleGetJob returns a single job including its output.
// GET /api/v1/jobs/{id}
func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok := s.jobs.get(id)
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("job not found"))
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// handleDeleteJob cancels a running job and removes it from the store.
// DELETE /api/v1/jobs/{id}
func (s *Server) handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.jobs.remove(id) {
		writeError(w, http.StatusNotFound, errors.New("job not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
