package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — Compliance: Regulatory Frameworks
// security/compliance/regulatory
// ════════════════════════════════════════════════════════════════

type RegulatoryFramework struct {
	ID         string
	Name       string
	Standard   string
	Region     string
	Scope      string
	Status     string
	Compliance string
	Updated    string
}

type regulatoryFrameworkStoreType struct {
	mu         sync.RWMutex
	frameworks map[string][]RegulatoryFramework
}

var regulatoryFrameworkStore = &regulatoryFrameworkStoreType{
	frameworks: map[string][]RegulatoryFramework{},
}

func (s *regulatoryFrameworkStoreType) list(slug string) []RegulatoryFramework {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.frameworks[slug]
	out := make([]RegulatoryFramework, len(src))
	copy(out, src)
	return out
}

func (s *regulatoryFrameworkStoreType) getByID(slug, id string) (RegulatoryFramework, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.frameworks[slug] {
		if f.ID == id {
			return f, true
		}
	}
	return RegulatoryFramework{}, false
}

func (s *regulatoryFrameworkStoreType) add(slug string, f RegulatoryFramework) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.frameworks[slug] = append(s.frameworks[slug], f)
}

func (s *regulatoryFrameworkStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.frameworks[slug]
	for i, f := range list {
		if f.ID == id {
			s.frameworks[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiRegulatoryFramework struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Standard   string `json:"standard"`
	Region     string `json:"region"`
	Scope      string `json:"scope"`
	Status     string `json:"status"`
	Compliance string `json:"compliance"`
	Updated    string `json:"updated"`
}

type apiRegulatoryFrameworkList struct {
	Frameworks []apiRegulatoryFramework `json:"frameworks"`
	Total      int                      `json:"total"`
}

func toAPIRegulatoryFramework(f RegulatoryFramework) apiRegulatoryFramework {
	return apiRegulatoryFramework{
		ID: f.ID, Name: f.Name, Standard: f.Standard,
		Region: f.Region, Scope: f.Scope, Status: f.Status,
		Compliance: f.Compliance, Updated: f.Updated,
	}
}

func filterRegulatoryFrameworks(all []RegulatoryFramework, search, status string) []RegulatoryFramework {
	var out []RegulatoryFramework
	q := strings.ToLower(search)
	for _, f := range all {
		if status != "" && f.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(f.Name + " " + f.Standard + " " + f.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, f)
	}
	return out
}

func apiWSSecCompRegulatoryListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	frameworks := filterRegulatoryFrameworks(regulatoryFrameworkStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiRegulatoryFramework, len(frameworks))
	for i, f := range frameworks {
		out[i] = toAPIRegulatoryFramework(f)
	}
	apiOK(w, apiRegulatoryFrameworkList{Frameworks: out, Total: len(out)})
}

type apiCreateRegulatoryFrameworkRequest struct {
	Name     string `json:"name"`
	Standard string `json:"standard"`
	Region   string `json:"region"`
	Scope    string `json:"scope"`
}

func apiWSSecCompRegulatoryCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateRegulatoryFrameworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	standard := strings.TrimSpace(req.Standard)
	if standard == "" {
		standard = "Internal"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "Global"
	}
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "All resources"
	}
	framework := RegulatoryFramework{
		ID: newID(), Name: name, Standard: standard,
		Region: region, Scope: scope, Status: "active",
		Compliance: "pending", Updated: "2026-04-17",
	}
	regulatoryFrameworkStore.add(slug, framework)
	apiCreated(w, toAPIRegulatoryFramework(framework))
}

func apiWSSecCompRegulatoryGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	framework, found := regulatoryFrameworkStore.getByID(slug, r.PathValue("fid"))
	if !found {
		apiError(w, http.StatusNotFound, "Regulatory framework not found.")
		return
	}
	apiOK(w, toAPIRegulatoryFramework(framework))
}

func apiWSSecCompRegulatoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := regulatoryFrameworkStore.getByID(slug, r.PathValue("fid")); !found {
		apiError(w, http.StatusNotFound, "Regulatory framework not found.")
		return
	}
	regulatoryFrameworkStore.delete(slug, r.PathValue("fid"))
	apiOK(w, map[string]any{"ok": true, "message": "Regulatory framework deleted."})
}
