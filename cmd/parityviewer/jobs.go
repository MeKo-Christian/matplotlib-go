package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	rerenderStatusRunning = "running"
	rerenderStatusDone    = "done"
	rerenderStatusFailed  = "failed"
)

type rerenderRunner func(ctx context.Context, names []string, progress func(string)) error

type rerenderJobStatus struct {
	ID        string `json:"job_id"`
	Status    string `json:"status"`
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Current   string `json:"current,omitempty"`
	Error     string `json:"error,omitempty"`
}

type rerenderJobManager struct {
	mu     sync.Mutex
	nextID atomic.Uint64
	jobs   map[string]*rerenderJob
	active string
	runner rerenderRunner
}

type rerenderJob struct {
	status rerenderJobStatus
}

func newRerenderJobManager(runner rerenderRunner) *rerenderJobManager {
	return &rerenderJobManager{
		jobs:   map[string]*rerenderJob{},
		runner: runner,
	}
}

func (m *rerenderJobManager) start(names []string) (rerenderJobStatus, error) {
	names = normalizeCaseNames(names)
	if len(names) == 0 {
		return rerenderJobStatus{}, errors.New("missing name parameter")
	}

	m.mu.Lock()
	if m.active != "" {
		m.mu.Unlock()
		return rerenderJobStatus{}, errors.New("rerender already running")
	}
	id := "rerender-" + strconv.FormatUint(m.nextID.Add(1), 10)
	status := rerenderJobStatus{
		ID:     id,
		Status: rerenderStatusRunning,
		Total:  len(names),
	}
	job := &rerenderJob{status: status}
	m.jobs[id] = job
	m.active = id
	m.mu.Unlock()

	go func() {
		err := m.runner(context.Background(), names, func(name string) {
			m.markProgress(id, name)
		})
		m.finish(id, err)
	}()

	return status, nil
}

func (m *rerenderJobManager) status(id string) (rerenderJobStatus, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return rerenderJobStatus{}, false
	}
	return job.status, true
}

func (m *rerenderJobManager) markProgress(id, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return
	}
	job.status.Current = name
	if job.status.Completed < job.status.Total {
		job.status.Completed++
	}
}

func (m *rerenderJobManager) finish(id string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return
	}
	if err != nil {
		job.status.Status = rerenderStatusFailed
		job.status.Error = err.Error()
	} else {
		job.status.Status = rerenderStatusDone
		job.status.Completed = job.status.Total
	}
	if m.active == id {
		m.active = ""
	}
}

func handleRerender(w http.ResponseWriter, r *http.Request, opts viewOptions, manager *rerenderJobManager) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !opts.CanRerender {
		http.Error(w, opts.RerenderDisabledMsg, http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	names := r.Form["name"]
	if r.FormValue("all") == "1" {
		names = nil
	}
	job, err := manager.start(names)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "already running") {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, http.StatusAccepted, job)
}

func handleRerenderStatus(w http.ResponseWriter, r *http.Request, manager *rerenderJobManager) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	status, ok := manager.status(id)
	if !ok {
		http.Error(w, "rerender job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
