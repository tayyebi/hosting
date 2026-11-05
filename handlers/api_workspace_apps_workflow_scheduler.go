package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Workflow: Scheduler / Cron
// apps/workflow/scheduler
// ════════════════════════════════════════════════════════════════

type ScheduledJob struct {
	ID       string
	Name     string
	Schedule string
	Command  string
	Runtime  string
	LastRun  string
	NextRun  string
	Duration string
	Status   string
	Updated  string
}

type scheduledJobStoreType struct {
	mu   sync.RWMutex
	jobs map[string][]ScheduledJob
}

var scheduledJobStore = &scheduledJobStoreType{
	jobs: map[string][]ScheduledJob{},
}

func (s *scheduledJobStoreType) list(slug string) []ScheduledJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.jobs[slug]
	out := make([]ScheduledJob, len(src))
	copy(out, src)
	return out
}

func (s *scheduledJobStoreType) getByID(slug, id string) (ScheduledJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs[slug] {
		if j.ID == id {
			return j, true
		}
	}
	return ScheduledJob{}, false
}

func (s *scheduledJobStoreType) add(slug string, j ScheduledJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[slug] = append(s.jobs[slug], j)
}

func (s *scheduledJobStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.jobs[slug]
	for i, j := range list {
		if j.ID == id {
			s.jobs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiScheduledJob struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Runtime  string `json:"runtime"`
	LastRun  string `json:"last_run"`
	NextRun  string `json:"next_run"`
	Duration string `json:"duration"`
	Status   string `json:"status"`
	Updated  string `json:"updated"`
}

type apiScheduledJobList struct {
	Jobs  []apiScheduledJob `json:"jobs"`
	Total int               `json:"total"`
}

func toAPIScheduledJob(j ScheduledJob) apiScheduledJob {
	return apiScheduledJob{
		ID: j.ID, Name: j.Name, Schedule: j.Schedule,
		Command: j.Command, Runtime: j.Runtime, LastRun: j.LastRun,
		NextRun: j.NextRun, Duration: j.Duration,
		Status: j.Status, Updated: j.Updated,
	}
}

func filterScheduledJobs(all []ScheduledJob, search, status string) []ScheduledJob {
	var out []ScheduledJob
	q := strings.ToLower(search)
	for _, j := range all {
		if status != "" && j.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(j.Name + " " + j.Command + " " + j.Runtime)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func apiWSAppsWorkflowSchedulerListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	jobs := filterScheduledJobs(scheduledJobStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiScheduledJob, len(jobs))
	for i, j := range jobs {
		out[i] = toAPIScheduledJob(j)
	}
	apiOK(w, apiScheduledJobList{Jobs: out, Total: len(out)})
}

type apiCreateScheduledJobRequest struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Runtime  string `json:"runtime"`
}

func apiWSAppsWorkflowSchedulerCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateScheduledJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	schedule := strings.TrimSpace(req.Schedule)
	if schedule == "" {
		schedule = "0 * * * *"
	}
	command := strings.TrimSpace(req.Command)
	if command == "" {
		apiError(w, http.StatusBadRequest, "Command is required.")
		return
	}
	runtime := strings.TrimSpace(req.Runtime)
	if runtime == "" {
		runtime = "Shell"
	}
	job := ScheduledJob{
		ID: newID(), Name: name, Schedule: schedule,
		Command: command, Runtime: runtime, LastRun: "—",
		NextRun: "pending", Duration: "—",
		Status: "active", Updated: "2026-04-17",
	}
	scheduledJobStore.add(slug, job)
	apiCreated(w, toAPIScheduledJob(job))
}

func apiWSAppsWorkflowSchedulerGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	job, found := scheduledJobStore.getByID(slug, r.PathValue("jid"))
	if !found {
		apiError(w, http.StatusNotFound, "Scheduled job not found.")
		return
	}
	apiOK(w, toAPIScheduledJob(job))
}

func apiWSAppsWorkflowSchedulerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := scheduledJobStore.getByID(slug, r.PathValue("jid")); !found {
		apiError(w, http.StatusNotFound, "Scheduled job not found.")
		return
	}
	scheduledJobStore.delete(slug, r.PathValue("jid"))
	apiOK(w, map[string]any{"ok": true, "message": "Scheduled job deleted."})
}
