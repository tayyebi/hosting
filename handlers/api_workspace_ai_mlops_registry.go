package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — MLOps: Model Registry
// ai/mlops/registry
// ════════════════════════════════════════════════════════════════

type ModelRegistryEntry struct {
	ID       string
	Name     string
	Stage    string
	Version  string
	Owner    string
	Format   string
	Size     string
	LastUsed string
	Created  string
	Region   string
}

type modelRegistryStoreType struct {
	mu     sync.RWMutex
	models map[string][]ModelRegistryEntry
}

var modelRegistryStore = &modelRegistryStoreType{
	models: map[string][]ModelRegistryEntry{},
}

func (s *modelRegistryStoreType) list(slug string) []ModelRegistryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.models[slug]
	out := make([]ModelRegistryEntry, len(src))
	copy(out, src)
	return out
}

func (s *modelRegistryStoreType) getByID(slug, id string) (ModelRegistryEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.models[slug] {
		if m.ID == id {
			return m, true
		}
	}
	return ModelRegistryEntry{}, false
}

func (s *modelRegistryStoreType) add(slug string, m ModelRegistryEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.models[slug] = append(s.models[slug], m)
}

func (s *modelRegistryStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.models[slug]
	for i, m := range list {
		if m.ID == id {
			s.models[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiModelRegistryEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Stage    string `json:"stage"`
	Version  string `json:"version"`
	Owner    string `json:"owner"`
	Format   string `json:"format"`
	Size     string `json:"size"`
	LastUsed string `json:"last_used"`
	Created  string `json:"created"`
	Region   string `json:"region"`
}

type apiModelRegistryList struct {
	Models []apiModelRegistryEntry `json:"models"`
	Total  int                     `json:"total"`
}

func toAPIModelRegistryEntry(m ModelRegistryEntry) apiModelRegistryEntry {
	return apiModelRegistryEntry{
		ID: m.ID, Name: m.Name, Stage: m.Stage,
		Version: m.Version, Owner: m.Owner,
		Format: m.Format, Size: m.Size,
		LastUsed: m.LastUsed, Created: m.Created,
		Region: m.Region,
	}
}

func filterModelRegistryEntries(all []ModelRegistryEntry, search, stage string) []ModelRegistryEntry {
	var out []ModelRegistryEntry
	q := strings.ToLower(search)
	for _, m := range all {
		if stage != "" && m.Stage != stage {
			continue
		}
		if q != "" {
			hay := strings.ToLower(m.Name + " " + m.Owner + " " + m.Format + " " + m.Version + " " + m.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func apiWSAIMLOpsRegistryListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	models := filterModelRegistryEntries(modelRegistryStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("stage"))
	out := make([]apiModelRegistryEntry, len(models))
	for i, m := range models {
		out[i] = toAPIModelRegistryEntry(m)
	}
	apiOK(w, apiModelRegistryList{Models: out, Total: len(out)})
}

type apiCreateModelRegistryRequest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Format  string `json:"format"`
	Owner   string `json:"owner"`
	Region  string `json:"region"`
}

func apiWSAIMLOpsRegistryCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateModelRegistryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "v1.0"
	}
	format := strings.TrimSpace(req.Format)
	if format == "" {
		format = "ONNX"
	}
	owner := strings.TrimSpace(req.Owner)
	if owner == "" {
		owner = "unassigned"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	model := ModelRegistryEntry{
		ID: newID(), Name: name, Stage: "staging",
		Version: version, Owner: owner,
		Format: format, Size: "—",
		LastUsed: "—", Created: "2026-04-17",
		Region: region,
	}
	modelRegistryStore.add(slug, model)
	apiCreated(w, toAPIModelRegistryEntry(model))
}

func apiWSAIMLOpsRegistryGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	model, found := modelRegistryStore.getByID(slug, r.PathValue("mid"))
	if !found {
		apiError(w, http.StatusNotFound, "Model not found.")
		return
	}
	apiOK(w, toAPIModelRegistryEntry(model))
}

func apiWSAIMLOpsRegistryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := modelRegistryStore.getByID(slug, r.PathValue("mid")); !found {
		apiError(w, http.StatusNotFound, "Model not found.")
		return
	}
	modelRegistryStore.delete(slug, r.PathValue("mid"))
	apiOK(w, map[string]any{"ok": true, "message": "Model deleted."})
}
