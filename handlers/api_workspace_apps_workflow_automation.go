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
	items: map[string][]WorkflowAutomation{},
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
