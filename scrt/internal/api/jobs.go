package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// jobStatus represents the lifecycle state of a background exec job.
type jobStatus string

const (
	jobRunning jobStatus = "running"
	jobDone    jobStatus = "done"
	jobError   jobStatus = "error"
)

// ExecJob holds the state of a single background exec job.
type ExecJob struct {
	ID        string     `json:"id"`
	Container string     `json:"container"`
	Cmd       string     `json:"cmd"`
	Status    jobStatus  `json:"status"`
	Output    string     `json:"output,omitempty"`
	ExitCode  int        `json:"exit_code"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	cancel    context.CancelFunc
}

// jobStore is a thread-safe in-memory store for background exec jobs.
type jobStore struct {
	mu   sync.Mutex
	jobs map[string]*ExecJob
}

func newJobStore() *jobStore {
	return &jobStore{jobs: make(map[string]*ExecJob)}
}

// create registers a new running job and returns it.
func (js *jobStore) create(container, cmd string, cancel context.CancelFunc) *ExecJob {
	js.mu.Lock()
	defer js.mu.Unlock()

	id := newJobID()
	j := &ExecJob{
		ID:        id,
		Container: container,
		Cmd:       cmd,
		Status:    jobRunning,
		StartedAt: time.Now().UTC(),
		cancel:    cancel,
	}
	js.jobs[id] = j
	return j
}

// get returns the job for id, or (nil, false) if not found.
func (js *jobStore) get(id string) (*ExecJob, bool) {
	js.mu.Lock()
	defer js.mu.Unlock()
	j, ok := js.jobs[id]
	return j, ok
}

// list returns a shallow copy of all jobs, newest first.
func (js *jobStore) list() []*ExecJob {
	js.mu.Lock()
	defer js.mu.Unlock()

	out := make([]*ExecJob, 0, len(js.jobs))
	for _, j := range js.jobs {
		out = append(out, j)
	}
	// Stable newest-first sort without importing sort package in hot path
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].StartedAt.After(out[j-1].StartedAt); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// finish records the result of a completed job.
func (js *jobStore) finish(id string, output []byte, exitCode int, execErr error) {
	js.mu.Lock()
	defer js.mu.Unlock()

	j, ok := js.jobs[id]
	if !ok {
		return
	}

	now := time.Now().UTC()
	j.EndedAt = &now
	j.Output = string(output)
	j.ExitCode = exitCode

	if execErr != nil {
		j.Status = jobError
		if j.Output == "" {
			j.Output = execErr.Error()
		}
	} else if exitCode != 0 {
		j.Status = jobError
	} else {
		j.Status = jobDone
	}
}

// cancel cancels a running job's context. Returns false if the job is not found.
func (js *jobStore) cancel(id string) bool {
	js.mu.Lock()
	defer js.mu.Unlock()

	j, ok := js.jobs[id]
	if !ok {
		return false
	}
	if j.cancel != nil {
		j.cancel()
	}
	return true
}

// remove deletes a job from the store after cancelling it.
func (js *jobStore) remove(id string) bool {
	js.mu.Lock()
	defer js.mu.Unlock()

	j, ok := js.jobs[id]
	if !ok {
		return false
	}
	if j.cancel != nil {
		j.cancel()
	}
	delete(js.jobs, id)
	return true
}

// newJobID returns a random 8-byte hex job identifier.
func newJobID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
