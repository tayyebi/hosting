package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Workflow: Automation
// apps/workflow/automation
// ════════════════════════════════════════════════════════════════

type WorkflowAutomation struct {
	ID          string
	Name        string
	Trigger     string
	Actions     int
	Runs        string
	SuccessRate string
	LastRun     string
	Status      string
	Updated     string
}

type workflowAutomationStoreType struct {
	mu    sync.RWMutex
	items map[string][]WorkflowAutomation
}

var workflowAutomationStore = &workflowAutomationStoreType{
	items: map[string][]WorkflowAutomation{
		"buyvm": {
			{ID: newID(), Name: "auto-scale-web-tier", Trigger: "CPU > 80%", Actions: 3, Runs: "1,240", SuccessRate: "99.6%", LastRun: "2026-04-17 08:12", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "nightly-db-backup", Trigger: "Cron 02:00 UTC", Actions: 5, Runs: "365", SuccessRate: "100%", LastRun: "2026-04-17 02:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "deploy-on-merge", Trigger: "Git merge to main", Actions: 7, Runs: "892", SuccessRate: "97.2%", LastRun: "2026-04-16 14:33", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "incident-alert-chain", Trigger: "Alert fired", Actions: 4, Runs: "56", SuccessRate: "100%", LastRun: "2026-04-15 22:10", Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "user-onboarding-flow", Trigger: "New user signup", Actions: 6, Runs: "4,310", SuccessRate: "98.8%", LastRun: "2026-04-16 19:45", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "legacy-etl-pipeline", Trigger: "Cron 06:00 UTC", Actions: 12, Runs: "730", SuccessRate: "94.1%", LastRun: "2026-04-17 06:00", Status: "deprecated", Updated: "2025-12-01"},
		},
		"hosting": {
			{ID: newID(), Name: "provision-new-vm", Trigger: "API request", Actions: 8, Runs: "2,150", SuccessRate: "99.1%", LastRun: "2026-04-17 09:30", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "ssl-cert-renewal", Trigger: "Cert expiry < 30d", Actions: 3, Runs: "89", SuccessRate: "100%", LastRun: "2026-04-14 03:00", Status: "active", Updated: "2026-04-14"},
			{ID: newID(), Name: "cleanup-orphan-volumes", Trigger: "Cron Sun 04:00", Actions: 4, Runs: "52", SuccessRate: "100%", LastRun: "2026-04-13 04:00", Status: "active", Updated: "2026-04-13"},
			{ID: newID(), Name: "abuse-detection-scan", Trigger: "Every 15 min", Actions: 5, Runs: "35,040", SuccessRate: "99.9%", LastRun: "2026-04-17 09:45", Status: "active", Updated: "2026-04-17"},
		},
	},
}

func (s *workflowAutomationStoreType) list(slug string) []WorkflowAutomation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]WorkflowAutomation, len(src))
	copy(out, src)
	return out
}

func (s *workflowAutomationStoreType) getByID(slug, id string) (WorkflowAutomation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.items[slug] {
		if w.ID == id {
			return w, true
		}
	}
	return WorkflowAutomation{}, false
}

func (s *workflowAutomationStoreType) add(slug string, w WorkflowAutomation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], w)
}

func (s *workflowAutomationStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, w := range list {
		if w.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiWorkflowAutomation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Trigger     string `json:"trigger"`
	Actions     int    `json:"actions"`
	Runs        string `json:"runs"`
	SuccessRate string `json:"success_rate"`
	LastRun     string `json:"last_run"`
	Status      string `json:"status"`
	Updated     string `json:"updated"`
}

type apiWorkflowAutomationList struct {
	Workflows []apiWorkflowAutomation `json:"workflows"`
	Total     int                     `json:"total"`
}

func toAPIWorkflowAutomation(w WorkflowAutomation) apiWorkflowAutomation {
	return apiWorkflowAutomation{
		ID: w.ID, Name: w.Name, Trigger: w.Trigger,
		Actions: w.Actions, Runs: w.Runs, SuccessRate: w.SuccessRate,
		LastRun: w.LastRun, Status: w.Status, Updated: w.Updated,
	}
}

func filterWorkflowAutomations(all []WorkflowAutomation, search, status string) []WorkflowAutomation {
	var out []WorkflowAutomation
	q := strings.ToLower(search)
	for _, w := range all {
		if status != "" && w.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(w.Name + " " + w.Trigger)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, w)
	}
	return out
}

func apiWSAppsWorkflowAutomationListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterWorkflowAutomations(workflowAutomationStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiWorkflowAutomation, len(items))
	for i, wf := range items {
		out[i] = toAPIWorkflowAutomation(wf)
	}
	apiOK(w, apiWorkflowAutomationList{Workflows: out, Total: len(out)})
}

type apiCreateWorkflowAutomationRequest struct {
	Name    string `json:"name"`
	Trigger string `json:"trigger"`
	Actions int    `json:"actions"`
}

func apiWSAppsWorkflowAutomationCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateWorkflowAutomationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	trigger := strings.TrimSpace(req.Trigger)
	if trigger == "" {
		trigger = "Manual"
	}
	actions := req.Actions
	if actions <= 0 {
		actions = 1
	}
	wf := WorkflowAutomation{
		ID: newID(), Name: name, Trigger: trigger,
		Actions: actions, Runs: "0", SuccessRate: "—",
		LastRun: "—", Status: "active", Updated: "2026-04-17",
	}
	workflowAutomationStore.add(slug, wf)
	apiCreated(w, toAPIWorkflowAutomation(wf))
}

func apiWSAppsWorkflowAutomationGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	wf, found := workflowAutomationStore.getByID(slug, r.PathValue("wfid"))
	if !found {
		apiError(w, http.StatusNotFound, "Workflow automation not found.")
		return
	}
	apiOK(w, toAPIWorkflowAutomation(wf))
}

func apiWSAppsWorkflowAutomationDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := workflowAutomationStore.getByID(slug, r.PathValue("wfid")); !found {
		apiError(w, http.StatusNotFound, "Workflow automation not found.")
		return
	}
	workflowAutomationStore.delete(slug, r.PathValue("wfid"))
	apiOK(w, map[string]any{"ok": true, "message": "Workflow automation deleted."})
}
