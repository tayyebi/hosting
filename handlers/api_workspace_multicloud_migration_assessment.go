package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Migration: Assessment & Discovery
// multicloud/migration/assessment
// ════════════════════════════════════════════════════════════════

type MigrationAssessment struct {
	ID               string
	Name             string
	SourceEnv        string
	TargetEnv        string
	DiscoveredAssets int
	ReadyToMigrate   int
	NeedsRemediation int
	Unsupported      int
	EstimatedCost    string
	Complexity       string
	LastScan         string
	Status           string
	Updated          string
}

type migrationAssessmentStoreType struct {
	mu    sync.RWMutex
	items map[string][]MigrationAssessment
}

var migrationAssessmentStore = &migrationAssessmentStoreType{
	items: map[string][]MigrationAssessment{},
}

func (s *migrationAssessmentStoreType) list(slug string) []MigrationAssessment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]MigrationAssessment, len(src))
	copy(out, src)
	return out
}

func (s *migrationAssessmentStoreType) getByID(slug, id string) (MigrationAssessment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.items[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return MigrationAssessment{}, false
}

func (s *migrationAssessmentStoreType) add(slug string, a MigrationAssessment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], a)
}

func (s *migrationAssessmentStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, a := range list {
		if a.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiMigrationAssessment struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	SourceEnv        string `json:"source_env"`
	TargetEnv        string `json:"target_env"`
	DiscoveredAssets int    `json:"discovered_assets"`
	ReadyToMigrate   int    `json:"ready_to_migrate"`
	NeedsRemediation int    `json:"needs_remediation"`
	Unsupported      int    `json:"unsupported"`
	EstimatedCost    string `json:"estimated_cost"`
	Complexity       string `json:"complexity"`
	LastScan         string `json:"last_scan"`
	Status           string `json:"status"`
	Updated          string `json:"updated"`
}

type apiMigrationAssessmentList struct {
	Assessments []apiMigrationAssessment `json:"assessments"`
	Total       int                      `json:"total"`
}

func toAPIMigrationAssessment(a MigrationAssessment) apiMigrationAssessment {
	return apiMigrationAssessment{
		ID: a.ID, Name: a.Name, SourceEnv: a.SourceEnv,
		TargetEnv: a.TargetEnv, DiscoveredAssets: a.DiscoveredAssets,
		ReadyToMigrate: a.ReadyToMigrate, NeedsRemediation: a.NeedsRemediation,
		Unsupported: a.Unsupported, EstimatedCost: a.EstimatedCost,
		Complexity: a.Complexity, LastScan: a.LastScan,
		Status: a.Status, Updated: a.Updated,
	}
}

func filterMigrationAssessments(all []MigrationAssessment, search, status string) []MigrationAssessment {
	var out []MigrationAssessment
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.SourceEnv + " " + a.TargetEnv + " " + a.Complexity)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSMulticloudMigrationAssessmentListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterMigrationAssessments(migrationAssessmentStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMigrationAssessment, len(items))
	for i, a := range items {
		out[i] = toAPIMigrationAssessment(a)
	}
	apiOK(w, apiMigrationAssessmentList{Assessments: out, Total: len(out)})
}

type apiCreateMigrationAssessmentRequest struct {
	Name      string `json:"name"`
	SourceEnv string `json:"source_env"`
	TargetEnv string `json:"target_env"`
}

func apiWSMulticloudMigrationAssessmentCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMigrationAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	srcEnv := strings.TrimSpace(req.SourceEnv)
	if srcEnv == "" {
		srcEnv = "On-premises"
	}
	tgtEnv := strings.TrimSpace(req.TargetEnv)
	if tgtEnv == "" {
		tgtEnv = "AWS"
	}
	assessment := MigrationAssessment{
		ID: newID(), Name: name, SourceEnv: srcEnv,
		TargetEnv: tgtEnv, DiscoveredAssets: 0, ReadyToMigrate: 0,
		NeedsRemediation: 0, Unsupported: 0, EstimatedCost: "—",
		Complexity: "Pending", LastScan: "—",
		Status: "active", Updated: "2026-04-17",
	}
	migrationAssessmentStore.add(slug, assessment)
	apiCreated(w, toAPIMigrationAssessment(assessment))
}

func apiWSMulticloudMigrationAssessmentGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	a, found := migrationAssessmentStore.getByID(slug, r.PathValue("aid"))
	if !found {
		apiError(w, http.StatusNotFound, "Migration assessment not found.")
		return
	}
	apiOK(w, toAPIMigrationAssessment(a))
}

func apiWSMulticloudMigrationAssessmentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := migrationAssessmentStore.getByID(slug, r.PathValue("aid")); !found {
		apiError(w, http.StatusNotFound, "Migration assessment not found.")
		return
	}
	migrationAssessmentStore.delete(slug, r.PathValue("aid"))
	apiOK(w, map[string]any{"ok": true, "message": "Migration assessment deleted."})
}
