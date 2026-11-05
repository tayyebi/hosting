package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — Relational: PostgreSQL
// databases/relational/postgresql
// ════════════════════════════════════════════════════════════════

type ManagedPostgres struct {
	ID              string
	Name            string
	Version         string
	Plan            string
	VCPUs           int
	MemoryGB        int
	StorageGB       int
	Region          string
	Host            string
	Port            int
	Replicas        int
	BackupSchedule  string
	SSLEnforced     bool
	ConnectionLimit int
	Status          string
	Updated         string
}

type managedPostgresStoreType struct {
	mu    sync.RWMutex
	items map[string][]ManagedPostgres
}

var managedPostgresStore = &managedPostgresStoreType{
	items: map[string][]ManagedPostgres{},
}

func (s *managedPostgresStoreType) list(slug string) []ManagedPostgres {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ManagedPostgres, len(src))
	copy(out, src)
	return out
}

func (s *managedPostgresStoreType) getByID(slug, id string) (ManagedPostgres, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, db := range s.items[slug] {
		if db.ID == id {
			return db, true
		}
	}
	return ManagedPostgres{}, false
}

func (s *managedPostgresStoreType) add(slug string, db ManagedPostgres) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], db)
}

func (s *managedPostgresStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, db := range list {
		if db.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiManagedPostgres struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	Plan            string `json:"plan"`
	VCPUs           int    `json:"vcpus"`
	MemoryGB        int    `json:"memory_gb"`
	StorageGB       int    `json:"storage_gb"`
	Region          string `json:"region"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Replicas        int    `json:"replicas"`
	BackupSchedule  string `json:"backup_schedule"`
	SSLEnforced     bool   `json:"ssl_enforced"`
	ConnectionLimit int    `json:"connection_limit"`
	Status          string `json:"status"`
	Updated         string `json:"updated"`
}

type apiManagedPostgresList struct {
	Databases []apiManagedPostgres `json:"databases"`
	Total     int                  `json:"total"`
}

func toAPIManagedPostgres(d ManagedPostgres) apiManagedPostgres {
	return apiManagedPostgres{
		ID: d.ID, Name: d.Name, Version: d.Version,
		Plan: d.Plan, VCPUs: d.VCPUs, MemoryGB: d.MemoryGB,
		StorageGB: d.StorageGB, Region: d.Region, Host: d.Host,
		Port: d.Port, Replicas: d.Replicas, BackupSchedule: d.BackupSchedule,
		SSLEnforced: d.SSLEnforced, ConnectionLimit: d.ConnectionLimit,
		Status: d.Status, Updated: d.Updated,
	}
}

func filterManagedPostgres(all []ManagedPostgres, search, status string) []ManagedPostgres {
	var out []ManagedPostgres
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Version + " " + d.Plan + " " + d.Region + " " + d.Host)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSDBRelationalPostgresListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterManagedPostgres(managedPostgresStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiManagedPostgres, len(items))
	for i, d := range items {
		out[i] = toAPIManagedPostgres(d)
	}
	apiOK(w, apiManagedPostgresList{Databases: out, Total: len(out)})
}

type apiCreateManagedPostgresRequest struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Plan      string `json:"plan"`
	VCPUs     int    `json:"vcpus"`
	MemoryGB  int    `json:"memory_gb"`
	StorageGB int    `json:"storage_gb"`
	Region    string `json:"region"`
}

func apiWSDBRelationalPostgresCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateManagedPostgresRequest
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
		version = "16.2"
	}
	plan := strings.TrimSpace(req.Plan)
	if plan == "" {
		plan = "Starter"
	}
	vcpus := req.VCPUs
	if vcpus <= 0 {
		vcpus = 1
	}
	memGB := req.MemoryGB
	if memGB <= 0 {
		memGB = 2
	}
	storageGB := req.StorageGB
	if storageGB <= 0 {
		storageGB = 20
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	db := ManagedPostgres{
		ID: newID(), Name: name, Version: version,
		Plan: plan, VCPUs: vcpus, MemoryGB: memGB,
		StorageGB: storageGB, Region: region,
		Host: "pg-" + name + ".internal", Port: 5432,
		Replicas: 0, BackupSchedule: "Daily 02:00 UTC",
		SSLEnforced: true, ConnectionLimit: 25,
		Status: "provisioning", Updated: "2026-04-17",
	}
	managedPostgresStore.add(slug, db)
	apiCreated(w, toAPIManagedPostgres(db))
}

func apiWSDBRelationalPostgresGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	db, found := managedPostgresStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "PostgreSQL database not found.")
		return
	}
	apiOK(w, toAPIManagedPostgres(db))
}

func apiWSDBRelationalPostgresDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := managedPostgresStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "PostgreSQL database not found.")
		return
	}
	managedPostgresStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "PostgreSQL database deleted."})
}
