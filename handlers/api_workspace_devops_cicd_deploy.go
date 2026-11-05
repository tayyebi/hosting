package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — CI/CD: Deployment Automation
// devops/cicd/deploy
// ════════════════════════════════════════════════════════════════

type DeployPipeline struct {
	ID          string
	Name        string
	Environment string
	Strategy    string
	Target      string
	Status      string
	LastDeploy  string
	Duration    string
	Updated     string
}

type deployPipelineStoreType struct {
	mu        sync.RWMutex
	pipelines map[string][]DeployPipeline
}

var deployPipelineStore = &deployPipelineStoreType{
	pipelines: map[string][]DeployPipeline{
		"buyvm": {
			{ID: newID(), Name: "Production Deploy", Environment: "production", Strategy: "blue-green", Target: "Kubernetes", Status: "active", LastDeploy: "2026-04-16 18:00", Duration: "3m 45s", Updated: "2026-04-16"},
			{ID: newID(), Name: "Staging Deploy", Environment: "staging", Strategy: "rolling", Target: "Kubernetes", Status: "active", LastDeploy: "2026-04-17 09:30", Duration: "2m 10s", Updated: "2026-04-17"},
			{ID: newID(), Name: "Dev Deploy", Environment: "development", Strategy: "recreate", Target: "Docker Compose", Status: "active", LastDeploy: "2026-04-17 10:15", Duration: "0m 48s", Updated: "2026-04-17"},
			{ID: newID(), Name: "Canary Release", Environment: "production", Strategy: "canary", Target: "Kubernetes", Status: "active", LastDeploy: "2026-04-14 12:00", Duration: "15m 20s", Updated: "2026-04-14"},
			{ID: newID(), Name: "Hotfix Pipeline", Environment: "production", Strategy: "rolling", Target: "Kubernetes", Status: "inactive", LastDeploy: "2026-03-20 22:00", Duration: "2m 05s", Updated: "2026-03-20"},
		},
		"hosting": {
			{ID: newID(), Name: "Production Release", Environment: "production", Strategy: "blue-green", Target: "Kubernetes", Status: "active", LastDeploy: "2026-04-13 16:00", Duration: "4m 30s", Updated: "2026-04-13"},
			{ID: newID(), Name: "Staging Pipeline", Environment: "staging", Strategy: "rolling", Target: "Kubernetes", Status: "active", LastDeploy: "2026-04-14 08:45", Duration: "2m 55s", Updated: "2026-04-14"},
			{ID: newID(), Name: "Edge Deploy", Environment: "edge", Strategy: "canary", Target: "Edge Nodes", Status: "active", LastDeploy: "2026-04-10 14:00", Duration: "8m 12s", Updated: "2026-04-10"},
		},
	},
}

func (s *deployPipelineStoreType) list(slug string) []DeployPipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.pipelines[slug]
	out := make([]DeployPipeline, len(src))
	copy(out, src)
	return out
}

func (s *deployPipelineStoreType) getByID(slug, id string) (DeployPipeline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.pipelines[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return DeployPipeline{}, false
}

func (s *deployPipelineStoreType) add(slug string, p DeployPipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelines[slug] = append(s.pipelines[slug], p)
}

func (s *deployPipelineStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.pipelines[slug]
	for i, p := range list {
		if p.ID == id {
			s.pipelines[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDeployPipeline struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Strategy    string `json:"strategy"`
	Target      string `json:"target"`
	Status      string `json:"status"`
	LastDeploy  string `json:"last_deploy"`
	Duration    string `json:"duration"`
	Updated     string `json:"updated"`
}

type apiDeployPipelineList struct {
	Pipelines []apiDeployPipeline `json:"pipelines"`
	Total     int                 `json:"total"`
}

func toAPIDeployPipeline(p DeployPipeline) apiDeployPipeline {
	return apiDeployPipeline{
		ID: p.ID, Name: p.Name, Environment: p.Environment,
		Strategy: p.Strategy, Target: p.Target, Status: p.Status,
		LastDeploy: p.LastDeploy, Duration: p.Duration, Updated: p.Updated,
	}
}

func filterDeployPipelines(all []DeployPipeline, search, status string) []DeployPipeline {
	var out []DeployPipeline
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Environment + " " + p.Target)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSDevOpsCICDDeployListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipelines := filterDeployPipelines(deployPipelineStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDeployPipeline, len(pipelines))
	for i, p := range pipelines {
		out[i] = toAPIDeployPipeline(p)
	}
	apiOK(w, apiDeployPipelineList{Pipelines: out, Total: len(out)})
}

type apiCreateDeployPipelineRequest struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Strategy    string `json:"strategy"`
	Target      string `json:"target"`
}

func apiWSDevOpsCICDDeployCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDeployPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	env := strings.TrimSpace(req.Environment)
	if env == "" {
		env = "staging"
	}
	strategy := strings.TrimSpace(req.Strategy)
	if strategy == "" {
		strategy = "rolling"
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		target = "Kubernetes"
	}
	pipeline := DeployPipeline{
		ID: newID(), Name: name, Environment: env,
		Strategy: strategy, Target: target, Status: "active",
		LastDeploy: "—", Duration: "—", Updated: "2026-04-17",
	}
	deployPipelineStore.add(slug, pipeline)
	apiCreated(w, toAPIDeployPipeline(pipeline))
}

func apiWSDevOpsCICDDeployGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipeline, found := deployPipelineStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "Deploy pipeline not found.")
		return
	}
	apiOK(w, toAPIDeployPipeline(pipeline))
}

func apiWSDevOpsCICDDeployDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := deployPipelineStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "Deploy pipeline not found.")
		return
	}
	deployPipelineStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "Deploy pipeline deleted."})
}
