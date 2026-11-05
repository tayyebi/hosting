package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — MLOps: Experiment Tracking
// ai/mlops/experiments
// ════════════════════════════════════════════════════════════════

type MLExperiment struct {
	ID        string
	Name      string
	Status    string
	BestRun   string
	Metric    string
	Runs      int
	Framework string
	Owner     string
	Created   string
	Region    string
}

type mlExperimentStoreType struct {
	mu          sync.RWMutex
	experiments map[string][]MLExperiment
}

var mlExperimentStore = &mlExperimentStoreType{
	experiments: map[string][]MLExperiment{},
}

func (s *mlExperimentStoreType) list(slug string) []MLExperiment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.experiments[slug]
	out := make([]MLExperiment, len(src))
	copy(out, src)
	return out
}

func (s *mlExperimentStoreType) getByID(slug, id string) (MLExperiment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.experiments[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return MLExperiment{}, false
}

func (s *mlExperimentStoreType) add(slug string, e MLExperiment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.experiments[slug] = append(s.experiments[slug], e)
}

func (s *mlExperimentStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.experiments[slug]
	for i, e := range list {
		if e.ID == id {
			s.experiments[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiMLExperiment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	BestRun   string `json:"best_run"`
	Metric    string `json:"metric"`
	Runs      int    `json:"runs"`
	Framework string `json:"framework"`
	Owner     string `json:"owner"`
	Created   string `json:"created"`
	Region    string `json:"region"`
}

type apiMLExperimentList struct {
	Experiments []apiMLExperiment `json:"experiments"`
	Total       int               `json:"total"`
}

func toAPIMLExperiment(e MLExperiment) apiMLExperiment {
	return apiMLExperiment{
		ID: e.ID, Name: e.Name, Status: e.Status,
		BestRun: e.BestRun, Metric: e.Metric,
		Runs: e.Runs, Framework: e.Framework,
		Owner: e.Owner, Created: e.Created,
		Region: e.Region,
	}
}

func filterMLExperiments(all []MLExperiment, search, status string) []MLExperiment {
	var out []MLExperiment
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Framework + " " + e.Owner + " " + e.Metric + " " + e.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSAIMLOpsExperimentsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	experiments := filterMLExperiments(mlExperimentStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMLExperiment, len(experiments))
	for i, e := range experiments {
		out[i] = toAPIMLExperiment(e)
	}
	apiOK(w, apiMLExperimentList{Experiments: out, Total: len(out)})
}

type apiCreateMLExperimentRequest struct {
	Name      string `json:"name"`
	Framework string `json:"framework"`
	Owner     string `json:"owner"`
	Region    string `json:"region"`
}

func apiWSAIMLOpsExperimentsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMLExperimentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	framework := strings.TrimSpace(req.Framework)
	if framework == "" {
		framework = "PyTorch"
	}
	owner := strings.TrimSpace(req.Owner)
	if owner == "" {
		owner = "unassigned"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	experiment := MLExperiment{
		ID: newID(), Name: name, Status: "active",
		BestRun: "—", Metric: "—",
		Runs: 0, Framework: framework,
		Owner: owner, Created: "2026-04-17",
		Region: region,
	}
	mlExperimentStore.add(slug, experiment)
	apiCreated(w, toAPIMLExperiment(experiment))
}

func apiWSAIMLOpsExperimentsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	experiment, found := mlExperimentStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "Experiment not found.")
		return
	}
	apiOK(w, toAPIMLExperiment(experiment))
}

func apiWSAIMLOpsExperimentsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := mlExperimentStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "Experiment not found.")
		return
	}
	mlExperimentStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "Experiment deleted."})
}
