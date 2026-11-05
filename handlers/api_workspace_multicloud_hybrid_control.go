package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Hybrid: Unified Control Plane
// multicloud/hybrid/control
// ════════════════════════════════════════════════════════════════

type ControlPlane struct {
	ID              string
	Name            string
	Provider        string
	Region          string
	Clusters        int
	ManagedServices int
	Policies        int
	SyncInterval    string
	HealthScore     string
	Status          string
	Updated         string
}

type controlPlaneStoreType struct {
	mu    sync.RWMutex
	items map[string][]ControlPlane
}

var controlPlaneStore = &controlPlaneStoreType{
	items: map[string][]ControlPlane{
		"buyvm": {
			{ID: newID(), Name: "primary-control", Provider: "AWS + On-Prem", Region: "us-east-1 / NYC DC", Clusters: 8, ManagedServices: 42, Policies: 18, SyncInterval: "30s", HealthScore: "98.5%", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "eu-control-plane", Provider: "Azure + GCP", Region: "eu-west-1 / eu-central", Clusters: 5, ManagedServices: 28, Policies: 15, SyncInterval: "30s", HealthScore: "97.2%", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "edge-controller", Provider: "Multi-edge", Region: "Global (12 PoPs)", Clusters: 12, ManagedServices: 15, Policies: 8, SyncInterval: "60s", HealthScore: "95.8%", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "staging-control", Provider: "AWS", Region: "us-west-2", Clusters: 2, ManagedServices: 10, Policies: 5, SyncInterval: "120s", HealthScore: "100%", Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "legacy-bridge", Provider: "On-Prem only", Region: "Chicago DC", Clusters: 1, ManagedServices: 4, Policies: 3, SyncInterval: "300s", HealthScore: "88.0%", Status: "degraded", Updated: "2026-03-20"},
		},
		"hosting": {
			{ID: newID(), Name: "global-control-plane", Provider: "AWS + Azure + GCP", Region: "Multi-region", Clusters: 15, ManagedServices: 65, Policies: 30, SyncInterval: "15s", HealthScore: "99.1%", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "apac-control", Provider: "AWS + Alibaba", Region: "ap-southeast-1 / cn-east", Clusters: 6, ManagedServices: 22, Policies: 12, SyncInterval: "30s", HealthScore: "96.7%", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "dev-sandbox-ctrl", Provider: "AWS", Region: "us-east-2", Clusters: 1, ManagedServices: 5, Policies: 2, SyncInterval: "60s", HealthScore: "100%", Status: "active", Updated: "2026-04-14"},
		},
	},
}

func (s *controlPlaneStoreType) list(slug string) []ControlPlane {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ControlPlane, len(src))
	copy(out, src)
	return out
}

func (s *controlPlaneStoreType) getByID(slug, id string) (ControlPlane, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.items[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return ControlPlane{}, false
}

func (s *controlPlaneStoreType) add(slug string, c ControlPlane) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], c)
}

func (s *controlPlaneStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, c := range list {
		if c.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiControlPlane struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Provider        string `json:"provider"`
	Region          string `json:"region"`
	Clusters        int    `json:"clusters"`
	ManagedServices int    `json:"managed_services"`
	Policies        int    `json:"policies"`
	SyncInterval    string `json:"sync_interval"`
	HealthScore     string `json:"health_score"`
	Status          string `json:"status"`
	Updated         string `json:"updated"`
}

type apiControlPlaneList struct {
	ControlPlanes []apiControlPlane `json:"control_planes"`
	Total         int               `json:"total"`
}

func toAPIControlPlane(c ControlPlane) apiControlPlane {
	return apiControlPlane{
		ID: c.ID, Name: c.Name, Provider: c.Provider,
		Region: c.Region, Clusters: c.Clusters, ManagedServices: c.ManagedServices,
		Policies: c.Policies, SyncInterval: c.SyncInterval, HealthScore: c.HealthScore,
		Status: c.Status, Updated: c.Updated,
	}
}

func filterControlPlanes(all []ControlPlane, search, status string) []ControlPlane {
	var out []ControlPlane
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Name + " " + c.Provider + " " + c.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func apiWSMulticloudHybridControlListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterControlPlanes(controlPlaneStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiControlPlane, len(items))
	for i, c := range items {
		out[i] = toAPIControlPlane(c)
	}
	apiOK(w, apiControlPlaneList{ControlPlanes: out, Total: len(out)})
}

type apiCreateControlPlaneRequest struct {
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	Region       string `json:"region"`
	SyncInterval string `json:"sync_interval"`
}

func apiWSMulticloudHybridControlCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateControlPlaneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "AWS"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	syncInterval := strings.TrimSpace(req.SyncInterval)
	if syncInterval == "" {
		syncInterval = "60s"
	}
	cp := ControlPlane{
		ID: newID(), Name: name, Provider: provider,
		Region: region, Clusters: 0, ManagedServices: 0,
		Policies: 0, SyncInterval: syncInterval, HealthScore: "—",
		Status: "active", Updated: "2026-04-17",
	}
	controlPlaneStore.add(slug, cp)
	apiCreated(w, toAPIControlPlane(cp))
}

func apiWSMulticloudHybridControlGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	cp, found := controlPlaneStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "Control plane not found.")
		return
	}
	apiOK(w, toAPIControlPlane(cp))
}

func apiWSMulticloudHybridControlDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := controlPlaneStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "Control plane not found.")
		return
	}
	controlPlaneStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "Control plane deleted."})
}
