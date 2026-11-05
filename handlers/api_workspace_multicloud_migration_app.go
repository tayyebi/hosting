package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Migration: Application Migration
// multicloud/migration/app
// ════════════════════════════════════════════════════════════════

type AppMigration struct {
	ID             string
	Name           string
	AppType        string
	SourcePlatform string
	TargetPlatform string
	Strategy       string
	Services       int
	Dependencies   int
	MigratedPct    string
	Rollback       string
	StartedAt      string
	Status         string
	Updated        string
}

type appMigrationStoreType struct {
	mu    sync.RWMutex
	items map[string][]AppMigration
}

var appMigrationStore = &appMigrationStoreType{
	items: map[string][]AppMigration{},
}

func (s *appMigrationStoreType) list(slug string) []AppMigration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]AppMigration, len(src))
	copy(out, src)
	return out
}

func (s *appMigrationStoreType) getByID(slug, id string) (AppMigration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.items[slug] {
		if a.ID == id {
			return a, true
		}
	}
	return AppMigration{}, false
}

func (s *appMigrationStoreType) add(slug string, a AppMigration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], a)
}

func (s *appMigrationStoreType) delete(slug, id string) {
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

type apiAppMigration struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	AppType        string `json:"app_type"`
	SourcePlatform string `json:"source_platform"`
	TargetPlatform string `json:"target_platform"`
	Strategy       string `json:"strategy"`
	Services       int    `json:"services"`
	Dependencies   int    `json:"dependencies"`
	MigratedPct    string `json:"migrated_pct"`
	Rollback       string `json:"rollback"`
	StartedAt      string `json:"started_at"`
	Status         string `json:"status"`
	Updated        string `json:"updated"`
}

type apiAppMigrationList struct {
	Migrations []apiAppMigration `json:"migrations"`
	Total      int               `json:"total"`
}

func toAPIAppMigration(a AppMigration) apiAppMigration {
	return apiAppMigration{
		ID: a.ID, Name: a.Name, AppType: a.AppType,
		SourcePlatform: a.SourcePlatform, TargetPlatform: a.TargetPlatform,
		Strategy: a.Strategy, Services: a.Services,
		Dependencies: a.Dependencies, MigratedPct: a.MigratedPct,
		Rollback: a.Rollback, StartedAt: a.StartedAt,
		Status: a.Status, Updated: a.Updated,
	}
}

func filterAppMigrations(all []AppMigration, search, status string) []AppMigration {
	var out []AppMigration
	q := strings.ToLower(search)
	for _, a := range all {
		if status != "" && a.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.Name + " " + a.AppType + " " + a.SourcePlatform + " " + a.TargetPlatform + " " + a.Strategy)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func apiWSMulticloudMigrationAppListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterAppMigrations(appMigrationStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiAppMigration, len(items))
	for i, a := range items {
		out[i] = toAPIAppMigration(a)
	}
	apiOK(w, apiAppMigrationList{Migrations: out, Total: len(out)})
}

type apiCreateAppMigrationRequest struct {
	Name           string `json:"name"`
	AppType        string `json:"app_type"`
	SourcePlatform string `json:"source_platform"`
	TargetPlatform string `json:"target_platform"`
	Strategy       string `json:"strategy"`
}

func apiWSMulticloudMigrationAppCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateAppMigrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	appType := strings.TrimSpace(req.AppType)
	if appType == "" {
		appType = "Web application"
	}
	srcPlat := strings.TrimSpace(req.SourcePlatform)
	if srcPlat == "" {
		srcPlat = "On-premises"
	}
	tgtPlat := strings.TrimSpace(req.TargetPlatform)
	if tgtPlat == "" {
		tgtPlat = "AWS"
	}
	strategy := strings.TrimSpace(req.Strategy)
	if strategy == "" {
		strategy = "Replatform"
	}
	migration := AppMigration{
		ID: newID(), Name: name, AppType: appType,
		SourcePlatform: srcPlat, TargetPlatform: tgtPlat,
		Strategy: strategy, Services: 0, Dependencies: 0,
		MigratedPct: "0%", Rollback: "N/A",
		StartedAt: "—", Status: "pending", Updated: "2026-04-17",
	}
	appMigrationStore.add(slug, migration)
	apiCreated(w, toAPIAppMigration(migration))
}

func apiWSMulticloudMigrationAppGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	a, found := appMigrationStore.getByID(slug, r.PathValue("mid"))
	if !found {
		apiError(w, http.StatusNotFound, "Application migration not found.")
		return
	}
	apiOK(w, toAPIAppMigration(a))
}

func apiWSMulticloudMigrationAppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := appMigrationStore.getByID(slug, r.PathValue("mid")); !found {
		apiError(w, http.StatusNotFound, "Application migration not found.")
		return
	}
	appMigrationStore.delete(slug, r.PathValue("mid"))
	apiOK(w, map[string]any{"ok": true, "message": "Application migration deleted."})
}
