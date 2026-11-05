package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — CI/CD: Build Automation
// devops/cicd/build
// ════════════════════════════════════════════════════════════════

type BuildPipeline struct {
	ID       string
	Name     string
	Repo     string
	Trigger  string
	Runtime  string
	Status   string
	LastRun  string
	Duration string
	Updated  string
}

type buildPipelineStoreType struct {
	mu        sync.RWMutex
	pipelines map[string][]BuildPipeline
}

var buildPipelineStore = &buildPipelineStoreType{
	pipelines: map[string][]BuildPipeline{},
}

func (s *buildPipelineStoreType) list(slug string) []BuildPipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.pipelines[slug]
	out := make([]BuildPipeline, len(src))
	copy(out, src)
	return out
}

func (s *buildPipelineStoreType) getByID(slug, id string) (BuildPipeline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.pipelines[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return BuildPipeline{}, false
}

func (s *buildPipelineStoreType) add(slug string, p BuildPipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelines[slug] = append(s.pipelines[slug], p)
}

func (s *buildPipelineStoreType) delete(slug, id string) {
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

type apiBuildPipeline struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Repo     string `json:"repo"`
	Trigger  string `json:"trigger"`
	Runtime  string `json:"runtime"`
	Status   string `json:"status"`
	LastRun  string `json:"last_run"`
	Duration string `json:"duration"`
	Updated  string `json:"updated"`
}

type apiBuildPipelineList struct {
	Pipelines []apiBuildPipeline `json:"pipelines"`
	Total     int                `json:"total"`
}

func toAPIBuildPipeline(p BuildPipeline) apiBuildPipeline {
	return apiBuildPipeline{
		ID: p.ID, Name: p.Name, Repo: p.Repo,
		Trigger: p.Trigger, Runtime: p.Runtime, Status: p.Status,
		LastRun: p.LastRun, Duration: p.Duration, Updated: p.Updated,
	}
}

func filterBuildPipelines(all []BuildPipeline, search, status string) []BuildPipeline {
	var out []BuildPipeline
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Repo + " " + p.Runtime)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSDevOpsCICDBuildListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipelines := filterBuildPipelines(buildPipelineStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiBuildPipeline, len(pipelines))
	for i, p := range pipelines {
		out[i] = toAPIBuildPipeline(p)
	}
	apiOK(w, apiBuildPipelineList{Pipelines: out, Total: len(out)})
}

type apiCreateBuildPipelineRequest struct {
	Name    string `json:"name"`
	Repo    string `json:"repo"`
	Trigger string `json:"trigger"`
	Runtime string `json:"runtime"`
}

func apiWSDevOpsCICDBuildCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateBuildPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	repo := strings.TrimSpace(req.Repo)
	if repo == "" {
		repo = "default"
	}
	trigger := strings.TrimSpace(req.Trigger)
	if trigger == "" {
		trigger = "push to main"
	}
	runtime := strings.TrimSpace(req.Runtime)
	if runtime == "" {
		runtime = "Docker"
	}
	pipeline := BuildPipeline{
		ID: newID(), Name: name, Repo: repo,
		Trigger: trigger, Runtime: runtime, Status: "pending",
		LastRun: "—", Duration: "—", Updated: "2026-04-17",
	}
	buildPipelineStore.add(slug, pipeline)
	apiCreated(w, toAPIBuildPipeline(pipeline))
}

func apiWSDevOpsCICDBuildGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipeline, found := buildPipelineStore.getByID(slug, r.PathValue("bid"))
	if !found {
		apiError(w, http.StatusNotFound, "Build pipeline not found.")
		return
	}
	apiOK(w, toAPIBuildPipeline(pipeline))
}

func apiWSDevOpsCICDBuildDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := buildPipelineStore.getByID(slug, r.PathValue("bid")); !found {
		apiError(w, http.StatusNotFound, "Build pipeline not found.")
		return
	}
	buildPipelineStore.delete(slug, r.PathValue("bid"))
	apiOK(w, map[string]any{"ok": true, "message": "Build pipeline deleted."})
}
