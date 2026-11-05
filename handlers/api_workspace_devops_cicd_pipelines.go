package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — CI/CD: Pipelines (Pro Feature)
// devops/cicd/pipelines
// ════════════════════════════════════════════════════════════════

type ProCICDPipeline struct {
	ID              string
	Name            string
	Repository      string
	Branch          string
	Trigger         string // "push", "merge_request", "tag", "schedule", "manual"
	BuildTool       string // "Docker", "Buildpack", "Kaniko", "Bazel"
	Runtime         string // "Node 20", "Go 1.22", "Python 3.12", "Java 21", ".NET 8"
	StageCount      int
	LastRunStatus   string // "success", "failed", "running", "cancelled"
	LastRunDuration int    // seconds
	TotalRuns       int
	SuccessRate     float64
	AvgDuration     int // seconds
	Artifacts       bool
	CacheEnabled    bool
	AutoDeploy      bool
	DeployTarget    string // "k8s-prod", "lambda", "static-cdn", "vm-fleet"
	Region          string
	CostPerMonth    float64
	Status          string // "active", "paused", "failing"
	Updated         string
}

type proCICDPipelineStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProCICDPipeline
}

var proCICDPipelineStore = &proCICDPipelineStoreType{
	items: map[string][]ProCICDPipeline{},
}

func (s *proCICDPipelineStoreType) list(slug string) []ProCICDPipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProCICDPipeline, len(src))
	copy(out, src)
	return out
}

func (s *proCICDPipelineStoreType) getByID(slug, id string) (ProCICDPipeline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.items[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return ProCICDPipeline{}, false
}

func (s *proCICDPipelineStoreType) add(slug string, p ProCICDPipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], p)
}

func (s *proCICDPipelineStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, p := range list {
		if p.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProCICDPipeline struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Repository      string  `json:"repository"`
	Branch          string  `json:"branch"`
	Trigger         string  `json:"trigger"`
	BuildTool       string  `json:"build_tool"`
	Runtime         string  `json:"runtime"`
	StageCount      int     `json:"stage_count"`
	LastRunStatus   string  `json:"last_run_status"`
	LastRunDuration int     `json:"last_run_duration_sec"`
	TotalRuns       int     `json:"total_runs"`
	SuccessRate     float64 `json:"success_rate"`
	AvgDuration     int     `json:"avg_duration_sec"`
	Artifacts       bool    `json:"artifacts"`
	CacheEnabled    bool    `json:"cache_enabled"`
	AutoDeploy      bool    `json:"auto_deploy"`
	DeployTarget    string  `json:"deploy_target"`
	Region          string  `json:"region"`
	CostPerMonth    float64 `json:"cost_per_month"`
	Status          string  `json:"status"`
	Updated         string  `json:"updated"`
}

type apiProCICDPipelineList struct {
	Pipelines []apiProCICDPipeline `json:"pipelines"`
	Total     int                  `json:"total"`
}

func toAPIProCICDPipeline(p ProCICDPipeline) apiProCICDPipeline {
	return apiProCICDPipeline{
		ID: p.ID, Name: p.Name, Repository: p.Repository,
		Branch: p.Branch, Trigger: p.Trigger,
		BuildTool: p.BuildTool, Runtime: p.Runtime,
		StageCount: p.StageCount, LastRunStatus: p.LastRunStatus,
		LastRunDuration: p.LastRunDuration, TotalRuns: p.TotalRuns,
		SuccessRate: p.SuccessRate, AvgDuration: p.AvgDuration,
		Artifacts: p.Artifacts, CacheEnabled: p.CacheEnabled,
		AutoDeploy: p.AutoDeploy, DeployTarget: p.DeployTarget,
		Region: p.Region, CostPerMonth: p.CostPerMonth,
		Status: p.Status, Updated: p.Updated,
	}
}

func filterProCICDPipelines(all []ProCICDPipeline, search, status string) []ProCICDPipeline {
	var out []ProCICDPipeline
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Repository + " " + p.Runtime + " " + p.DeployTarget + " " + p.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSDevOpsCICDPipelineListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProCICDPipelines(proCICDPipelineStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProCICDPipeline, len(items))
	for i, p := range items {
		out[i] = toAPIProCICDPipeline(p)
	}
	apiOK(w, apiProCICDPipelineList{Pipelines: out, Total: len(out)})
}

type apiCreateProCICDPipelineRequest struct {
	Name         string `json:"name"`
	Repository   string `json:"repository"`
	Branch       string `json:"branch"`
	Trigger      string `json:"trigger"`
	BuildTool    string `json:"build_tool"`
	Runtime      string `json:"runtime"`
	AutoDeploy   bool   `json:"auto_deploy"`
	DeployTarget string `json:"deploy_target"`
	Region       string `json:"region"`
}

func apiWSDevOpsCICDPipelineCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProCICDPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	repo := strings.TrimSpace(req.Repository)
	if repo == "" {
		repo = "github.com/org/repo"
	}
	branch := strings.TrimSpace(req.Branch)
	if branch == "" {
		branch = "main"
	}
	trigger := strings.TrimSpace(req.Trigger)
	if trigger == "" {
		trigger = "push"
	}
	buildTool := strings.TrimSpace(req.BuildTool)
	if buildTool == "" {
		buildTool = "Docker"
	}
	runtime := strings.TrimSpace(req.Runtime)
	if runtime == "" {
		runtime = "Node 20"
	}
	deployTarget := strings.TrimSpace(req.DeployTarget)
	if !req.AutoDeploy {
		deployTarget = "—"
	} else if deployTarget == "" {
		deployTarget = "k8s-prod"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	pipe := ProCICDPipeline{
		ID: newID(), Name: name, Repository: repo,
		Branch: branch, Trigger: trigger,
		BuildTool: buildTool, Runtime: runtime,
		StageCount: 3, LastRunStatus: "—",
		LastRunDuration: 0, TotalRuns: 0,
		SuccessRate: 0, AvgDuration: 0,
		Artifacts: true, CacheEnabled: true,
		AutoDeploy: req.AutoDeploy, DeployTarget: deployTarget,
		Region: region, CostPerMonth: 15.00,
		Status: "active", Updated: "2026-04-18",
	}
	proCICDPipelineStore.add(slug, pipe)
	apiCreated(w, toAPIProCICDPipeline(pipe))
}

func apiWSDevOpsCICDPipelineGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	p, found := proCICDPipelineStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "Pipeline not found.")
		return
	}
	apiOK(w, toAPIProCICDPipeline(p))
}

func apiWSDevOpsCICDPipelineDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proCICDPipelineStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "Pipeline not found.")
		return
	}
	proCICDPipelineStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "Pipeline deleted."})
}
