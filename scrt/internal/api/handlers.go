package api

import (
	"encoding/json"
	"errors"
	"net/http"

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
