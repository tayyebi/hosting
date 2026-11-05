package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — Observability: Logging
// devops/obs/logging
// ════════════════════════════════════════════════════════════════

type LogPipeline struct {
	ID        string
	Name      string
	Source    string
	Format    string
	Storage   string
	Retention string
	Volume    string
	Status    string
	Updated   string
}

type logPipelineStoreType struct {
	mu        sync.RWMutex
	pipelines map[string][]LogPipeline
}

var logPipelineStore = &logPipelineStoreType{
	pipelines: map[string][]LogPipeline{},
}

func (s *logPipelineStoreType) list(slug string) []LogPipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.pipelines[slug]
	out := make([]LogPipeline, len(src))
	copy(out, src)
	return out
}

func (s *logPipelineStoreType) getByID(slug, id string) (LogPipeline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.pipelines[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return LogPipeline{}, false
}

func (s *logPipelineStoreType) add(slug string, p LogPipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelines[slug] = append(s.pipelines[slug], p)
}

func (s *logPipelineStoreType) delete(slug, id string) {
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

type apiLogPipeline struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Format    string `json:"format"`
	Storage   string `json:"storage"`
	Retention string `json:"retention"`
	Volume    string `json:"volume"`
	Status    string `json:"status"`
	Updated   string `json:"updated"`
}

type apiLogPipelineList struct {
	Pipelines []apiLogPipeline `json:"pipelines"`
	Total     int              `json:"total"`
}

func toAPILogPipeline(p LogPipeline) apiLogPipeline {
	return apiLogPipeline{
		ID: p.ID, Name: p.Name, Source: p.Source,
		Format: p.Format, Storage: p.Storage, Retention: p.Retention,
		Volume: p.Volume, Status: p.Status, Updated: p.Updated,
	}
}

func filterLogPipelines(all []LogPipeline, search, status string) []LogPipeline {
	var out []LogPipeline
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Source + " " + p.Format + " " + p.Storage)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSDevOpsObsLoggingListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipelines := filterLogPipelines(logPipelineStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiLogPipeline, len(pipelines))
	for i, p := range pipelines {
		out[i] = toAPILogPipeline(p)
	}
	apiOK(w, apiLogPipelineList{Pipelines: out, Total: len(out)})
}

type apiCreateLogPipelineRequest struct {
	Name      string `json:"name"`
	Source    string `json:"source"`
	Format    string `json:"format"`
	Storage   string `json:"storage"`
	Retention string `json:"retention"`
}

func apiWSDevOpsObsLoggingCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateLogPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "Application"
	}
	format := strings.TrimSpace(req.Format)
	if format == "" {
		format = "JSON structured"
	}
	storage := strings.TrimSpace(req.Storage)
	if storage == "" {
		storage = "Elasticsearch"
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "30 days"
	}
	pipeline := LogPipeline{
		ID: newID(), Name: name, Source: source,
		Format: format, Storage: storage, Retention: retention,
		Volume: "0 B/day", Status: "active", Updated: "2026-04-17",
	}
	logPipelineStore.add(slug, pipeline)
	apiCreated(w, toAPILogPipeline(pipeline))
}

func apiWSDevOpsObsLoggingGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipeline, found := logPipelineStore.getByID(slug, r.PathValue("lid"))
	if !found {
		apiError(w, http.StatusNotFound, "Log pipeline not found.")
		return
	}
	apiOK(w, toAPILogPipeline(pipeline))
}

func apiWSDevOpsObsLoggingDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := logPipelineStore.getByID(slug, r.PathValue("lid")); !found {
		apiError(w, http.StatusNotFound, "Log pipeline not found.")
		return
	}
	logPipelineStore.delete(slug, r.PathValue("lid"))
	apiOK(w, map[string]any{"ok": true, "message": "Log pipeline deleted."})
}
