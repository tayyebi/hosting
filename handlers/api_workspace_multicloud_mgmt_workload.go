package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Management: Workload Portability
// multicloud/mgmt/workload
// ════════════════════════════════════════════════════════════════

type WorkloadProfile struct {
	ID             string
	Name           string
	SourceProvider string
	TargetProvider string
	WorkloadType   string
	Components     int
	Portability    string
	DriftStatus    string
	LastSync       string
	Status         string
	Updated        string
}

type workloadProfileStoreType struct {
	mu    sync.RWMutex
	items map[string][]WorkloadProfile
}

var workloadProfileStore = &workloadProfileStoreType{
	items: map[string][]WorkloadProfile{},
}

func (s *workloadProfileStoreType) list(slug string) []WorkloadProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]WorkloadProfile, len(src))
	copy(out, src)
	return out
}

func (s *workloadProfileStoreType) getByID(slug, id string) (WorkloadProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.items[slug] {
		if w.ID == id {
			return w, true
		}
	}
	return WorkloadProfile{}, false
}

func (s *workloadProfileStoreType) add(slug string, w WorkloadProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], w)
}

func (s *workloadProfileStoreType) delete(slug, id string) {
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

type apiWorkloadProfile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SourceProvider string `json:"source_provider"`
	TargetProvider string `json:"target_provider"`
	WorkloadType   string `json:"workload_type"`
	Components     int    `json:"components"`
	Portability    string `json:"portability"`
	DriftStatus    string `json:"drift_status"`
	LastSync       string `json:"last_sync"`
	Status         string `json:"status"`
	Updated        string `json:"updated"`
}

type apiWorkloadProfileList struct {
	Workloads []apiWorkloadProfile `json:"workloads"`
	Total     int                  `json:"total"`
}

func toAPIWorkloadProfile(w WorkloadProfile) apiWorkloadProfile {
	return apiWorkloadProfile{
		ID: w.ID, Name: w.Name, SourceProvider: w.SourceProvider,
		TargetProvider: w.TargetProvider, WorkloadType: w.WorkloadType,
		Components: w.Components, Portability: w.Portability,
		DriftStatus: w.DriftStatus, LastSync: w.LastSync,
		Status: w.Status, Updated: w.Updated,
	}
}

func filterWorkloadProfiles(all []WorkloadProfile, search, status string) []WorkloadProfile {
	var out []WorkloadProfile
	q := strings.ToLower(search)
	for _, w := range all {
		if status != "" && w.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(w.Name + " " + w.SourceProvider + " " + w.TargetProvider + " " + w.WorkloadType)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, w)
	}
	return out
}

func apiWSMulticloudMgmtWorkloadListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterWorkloadProfiles(workloadProfileStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiWorkloadProfile, len(items))
	for i, wp := range items {
		out[i] = toAPIWorkloadProfile(wp)
	}
	apiOK(w, apiWorkloadProfileList{Workloads: out, Total: len(out)})
}

type apiCreateWorkloadProfileRequest struct {
	Name           string `json:"name"`
	SourceProvider string `json:"source_provider"`
	TargetProvider string `json:"target_provider"`
	WorkloadType   string `json:"workload_type"`
}

func apiWSMulticloudMgmtWorkloadCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateWorkloadProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	src := strings.TrimSpace(req.SourceProvider)
	if src == "" {
		src = "AWS"
	}
	tgt := strings.TrimSpace(req.TargetProvider)
	if tgt == "" {
		tgt = "Azure"
	}
	wt := strings.TrimSpace(req.WorkloadType)
	if wt == "" {
		wt = "Container"
	}
	wp := WorkloadProfile{
		ID: newID(), Name: name, SourceProvider: src,
		TargetProvider: tgt, WorkloadType: wt,
		Components: 0, Portability: "—", DriftStatus: "Pending assessment",
		LastSync: "—", Status: "active", Updated: "2026-04-17",
	}
	workloadProfileStore.add(slug, wp)
	apiCreated(w, toAPIWorkloadProfile(wp))
}

func apiWSMulticloudMgmtWorkloadGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	wp, found := workloadProfileStore.getByID(slug, r.PathValue("wid"))
	if !found {
		apiError(w, http.StatusNotFound, "Workload profile not found.")
		return
	}
	apiOK(w, toAPIWorkloadProfile(wp))
}

func apiWSMulticloudMgmtWorkloadDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := workloadProfileStore.getByID(slug, r.PathValue("wid")); !found {
		apiError(w, http.StatusNotFound, "Workload profile not found.")
		return
	}
	workloadProfileStore.delete(slug, r.PathValue("wid"))
	apiOK(w, map[string]any{"ok": true, "message": "Workload profile deleted."})
}
