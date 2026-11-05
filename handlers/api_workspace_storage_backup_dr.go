package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Storage API — Backup: Disaster Recovery
// storage/backup/dr
// ════════════════════════════════════════════════════════════════

type DRPlan struct {
	ID              string
	Name            string
	Region          string
	Status          string
	SourceRegion    string
	TargetRegion    string
	RPO             string
	RTO             string
	ReplicationType string
	LastTestedAt    string
	Protected       int
	Encrypted       bool
}

type drPlanStoreType struct {
	mu    sync.RWMutex
	plans map[string][]DRPlan
}

var drPlanStore = &drPlanStoreType{
	plans: map[string][]DRPlan{},
}

func (s *drPlanStoreType) list(slug string) []DRPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.plans[slug]
	out := make([]DRPlan, len(src))
	copy(out, src)
	return out
}

func (s *drPlanStoreType) getByID(slug, id string) (DRPlan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.plans[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return DRPlan{}, false
}

func (s *drPlanStoreType) add(slug string, p DRPlan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans[slug] = append(s.plans[slug], p)
}

func (s *drPlanStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.plans[slug]
	for i, p := range list {
		if p.ID == id {
			s.plans[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDRPlan struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Region          string `json:"region"`
	Status          string `json:"status"`
	SourceRegion    string `json:"source_region"`
	TargetRegion    string `json:"target_region"`
	RPO             string `json:"rpo"`
	RTO             string `json:"rto"`
	ReplicationType string `json:"replication_type"`
	LastTestedAt    string `json:"last_tested_at"`
	Protected       int    `json:"protected_resources"`
	Encrypted       bool   `json:"encrypted"`
}

type apiDRPlanList struct {
	Plans []apiDRPlan `json:"plans"`
	Total int         `json:"total"`
}

func toAPIDRPlan(p DRPlan) apiDRPlan {
	return apiDRPlan{
		ID: p.ID, Name: p.Name, Region: p.Region,
		Status: p.Status, SourceRegion: p.SourceRegion,
		TargetRegion: p.TargetRegion, RPO: p.RPO,
		RTO: p.RTO, ReplicationType: p.ReplicationType,
		LastTestedAt: p.LastTestedAt, Protected: p.Protected,
		Encrypted: p.Encrypted,
	}
}

func filterDRPlans(all []DRPlan, search, status string) []DRPlan {
	var out []DRPlan
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Region + " " + p.SourceRegion + " " + p.TargetRegion + " " + p.ReplicationType)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSStorageBackupDRListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	plans := filterDRPlans(drPlanStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDRPlan, len(plans))
	for i, p := range plans {
		out[i] = toAPIDRPlan(p)
	}
	apiOK(w, apiDRPlanList{Plans: out, Total: len(out)})
}

type apiCreateDRPlanRequest struct {
	Name            string `json:"name"`
	Region          string `json:"region"`
	SourceRegion    string `json:"source_region"`
	TargetRegion    string `json:"target_region"`
	RPO             string `json:"rpo"`
	RTO             string `json:"rto"`
	ReplicationType string `json:"replication_type"`
	Encrypted       bool   `json:"encrypted"`
}

func apiWSStorageBackupDRCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDRPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	sourceRegion := strings.TrimSpace(req.SourceRegion)
	if sourceRegion == "" {
		apiError(w, http.StatusBadRequest, "Source region is required.")
		return
	}
	targetRegion := strings.TrimSpace(req.TargetRegion)
	if targetRegion == "" {
		apiError(w, http.StatusBadRequest, "Target region is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = sourceRegion
	}
	rpo := strings.TrimSpace(req.RPO)
	if rpo == "" {
		rpo = "1 hour"
	}
	rto := strings.TrimSpace(req.RTO)
	if rto == "" {
		rto = "4 hours"
	}
	replicationType := strings.TrimSpace(req.ReplicationType)
	if replicationType == "" {
		replicationType = "scheduled"
	}
	plan := DRPlan{
		ID: newID(), Name: name, Region: region,
		Status: "active", SourceRegion: sourceRegion,
		TargetRegion: targetRegion, RPO: rpo,
		RTO: rto, ReplicationType: replicationType,
		LastTestedAt: "—", Protected: 0,
		Encrypted: req.Encrypted,
	}
	drPlanStore.add(slug, plan)
	apiCreated(w, toAPIDRPlan(plan))
}

func apiWSStorageBackupDRGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	plan, found := drPlanStore.getByID(slug, r.PathValue("drid"))
	if !found {
		apiError(w, http.StatusNotFound, "DR plan not found.")
		return
	}
	apiOK(w, toAPIDRPlan(plan))
}

func apiWSStorageBackupDRDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := drPlanStore.getByID(slug, r.PathValue("drid")); !found {
		apiError(w, http.StatusNotFound, "DR plan not found.")
		return
	}
	drPlanStore.delete(slug, r.PathValue("drid"))
	apiOK(w, map[string]any{"ok": true, "message": "DR plan deleted."})
}
