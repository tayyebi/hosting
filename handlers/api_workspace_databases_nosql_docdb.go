package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — NoSQL: Document DB (Pro Feature)
// databases/nosql/document-db
// ════════════════════════════════════════════════════════════════

type ProDocumentDB struct {
	ID              string
	Name            string
	Engine          string
	EngineVersion   string
	ClusterNodes    int
	StorageGB       int
	DocumentCount   int64
	CollectionCount int
	ReadsPerSec     int
	WritesPerSec    int
	AvgLatencyMs    float64
	Encrypted       bool
	BackupEnabled   bool
	BackupRetention int
	Region          string
	MultiAZ         bool
	CostPerMonth    float64
	Status          string
	Updated         string
}

type proDocumentDBStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProDocumentDB
}

var proDocumentDBStore = &proDocumentDBStoreType{
	items: map[string][]ProDocumentDB{
		"buyvm": {
			{ID: newID(), Name: "app-data-prod", Engine: "MongoDB", EngineVersion: "7.0.12", ClusterNodes: 3, StorageGB: 240, DocumentCount: 18400000, CollectionCount: 42, ReadsPerSec: 12400, WritesPerSec: 3200, AvgLatencyMs: 1.8, Encrypted: true, BackupEnabled: true, BackupRetention: 30, Region: "us-east-1", MultiAZ: true, CostPerMonth: 420.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "user-profiles", Engine: "MongoDB", EngineVersion: "7.0.12", ClusterNodes: 3, StorageGB: 85, DocumentCount: 4200000, CollectionCount: 8, ReadsPerSec: 8600, WritesPerSec: 640, AvgLatencyMs: 0.9, Encrypted: true, BackupEnabled: true, BackupRetention: 14, Region: "us-east-1", MultiAZ: true, CostPerMonth: 280.00, Status: "running", Updated: "2026-04-17"},
			{ID: newID(), Name: "event-store", Engine: "MongoDB", EngineVersion: "6.0.18", ClusterNodes: 1, StorageGB: 420, DocumentCount: 124000000, CollectionCount: 6, ReadsPerSec: 2200, WritesPerSec: 8400, AvgLatencyMs: 2.4, Encrypted: true, BackupEnabled: true, BackupRetention: 7, Region: "us-east-1", MultiAZ: false, CostPerMonth: 180.00, Status: "running", Updated: "2026-04-17"},
			{ID: newID(), Name: "staging-docs", Engine: "MongoDB", EngineVersion: "7.0.12", ClusterNodes: 1, StorageGB: 12, DocumentCount: 86000, CollectionCount: 42, ReadsPerSec: 120, WritesPerSec: 80, AvgLatencyMs: 3.2, Encrypted: false, BackupEnabled: false, BackupRetention: 0, Region: "us-east-1", MultiAZ: false, CostPerMonth: 45.00, Status: "running", Updated: "2026-04-15"},
		},
		"hosting": {
			{ID: newID(), Name: "platform-core", Engine: "MongoDB", EngineVersion: "7.0.12", ClusterNodes: 5, StorageGB: 1200, DocumentCount: 890000000, CollectionCount: 86, ReadsPerSec: 42000, WritesPerSec: 12000, AvgLatencyMs: 1.2, Encrypted: true, BackupEnabled: true, BackupRetention: 30, Region: "eu-central-1", MultiAZ: true, CostPerMonth: 1800.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "search-index", Engine: "MongoDB", EngineVersion: "7.0.12", ClusterNodes: 3, StorageGB: 320, DocumentCount: 42000000, CollectionCount: 12, ReadsPerSec: 18000, WritesPerSec: 4200, AvgLatencyMs: 1.5, Encrypted: true, BackupEnabled: true, BackupRetention: 14, Region: "eu-central-1", MultiAZ: true, CostPerMonth: 640.00, Status: "running", Updated: "2026-04-17"},
			{ID: newID(), Name: "cms-content", Engine: "MongoDB", EngineVersion: "7.0.12", ClusterNodes: 2, StorageGB: 64, DocumentCount: 1800000, CollectionCount: 24, ReadsPerSec: 6400, WritesPerSec: 420, AvgLatencyMs: 1.1, Encrypted: true, BackupEnabled: true, BackupRetention: 14, Region: "us-west-2", MultiAZ: true, CostPerMonth: 220.00, Status: "running", Updated: "2026-04-16"},
			{ID: newID(), Name: "analytics-staging", Engine: "MongoDB", EngineVersion: "6.0.18", ClusterNodes: 1, StorageGB: 28, DocumentCount: 2400000, CollectionCount: 8, ReadsPerSec: 800, WritesPerSec: 200, AvgLatencyMs: 4.8, Encrypted: false, BackupEnabled: false, BackupRetention: 0, Region: "eu-central-1", MultiAZ: false, CostPerMonth: 65.00, Status: "stopped", Updated: "2026-04-12"},
		},
	},
}

func (s *proDocumentDBStoreType) list(slug string) []ProDocumentDB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProDocumentDB, len(src))
	copy(out, src)
	return out
}

func (s *proDocumentDBStoreType) getByID(slug, id string) (ProDocumentDB, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.items[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return ProDocumentDB{}, false
}

func (s *proDocumentDBStoreType) add(slug string, d ProDocumentDB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], d)
}

func (s *proDocumentDBStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, d := range list {
		if d.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProDocumentDB struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Engine          string  `json:"engine"`
	EngineVersion   string  `json:"engine_version"`
	ClusterNodes    int     `json:"cluster_nodes"`
	StorageGB       int     `json:"storage_gb"`
	DocumentCount   int64   `json:"document_count"`
	CollectionCount int     `json:"collection_count"`
	ReadsPerSec     int     `json:"reads_per_sec"`
	WritesPerSec    int     `json:"writes_per_sec"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	Encrypted       bool    `json:"encrypted"`
	BackupEnabled   bool    `json:"backup_enabled"`
	BackupRetention int     `json:"backup_retention_days"`
	Region          string  `json:"region"`
	MultiAZ         bool    `json:"multi_az"`
	CostPerMonth    float64 `json:"cost_per_month"`
	Status          string  `json:"status"`
	Updated         string  `json:"updated"`
}

type apiProDocumentDBList struct {
	Databases []apiProDocumentDB `json:"databases"`
	Total     int                `json:"total"`
}

func toAPIProDocumentDB(d ProDocumentDB) apiProDocumentDB {
	return apiProDocumentDB{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		EngineVersion: d.EngineVersion, ClusterNodes: d.ClusterNodes,
		StorageGB: d.StorageGB, DocumentCount: d.DocumentCount,
		CollectionCount: d.CollectionCount,
		ReadsPerSec: d.ReadsPerSec, WritesPerSec: d.WritesPerSec,
		AvgLatencyMs: d.AvgLatencyMs, Encrypted: d.Encrypted,
		BackupEnabled: d.BackupEnabled, BackupRetention: d.BackupRetention,
		Region: d.Region, MultiAZ: d.MultiAZ,
		CostPerMonth: d.CostPerMonth, Status: d.Status,
		Updated: d.Updated,
	}
}

func filterProDocumentDBs(all []ProDocumentDB, search, status string) []ProDocumentDB {
	var out []ProDocumentDB
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Engine + " " + d.EngineVersion + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSDBNoSQLDocDBListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProDocumentDBs(proDocumentDBStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProDocumentDB, len(items))
	for i, d := range items {
		out[i] = toAPIProDocumentDB(d)
	}
	apiOK(w, apiProDocumentDBList{Databases: out, Total: len(out)})
}

type apiCreateProDocumentDBRequest struct {
	Name          string `json:"name"`
	Engine        string `json:"engine"`
	EngineVersion string `json:"engine_version"`
	ClusterNodes  int    `json:"cluster_nodes"`
	StorageGB     int    `json:"storage_gb"`
	Encrypted     bool   `json:"encrypted"`
	Region        string `json:"region"`
	MultiAZ       bool   `json:"multi_az"`
}

func apiWSDBNoSQLDocDBCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProDocumentDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	engine := strings.TrimSpace(req.Engine)
	if engine == "" {
		engine = "MongoDB"
	}
	version := strings.TrimSpace(req.EngineVersion)
	if version == "" {
		version = "7.0.12"
	}
	nodes := req.ClusterNodes
	if nodes <= 0 {
		nodes = 1
	}
	storageGB := req.StorageGB
	if storageGB <= 0 {
		storageGB = 50
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	db := ProDocumentDB{
		ID: newID(), Name: name, Engine: engine,
		EngineVersion: version, ClusterNodes: nodes,
		StorageGB: storageGB, DocumentCount: 0,
		CollectionCount: 0, ReadsPerSec: 0, WritesPerSec: 0,
		AvgLatencyMs: 0, Encrypted: req.Encrypted,
		BackupEnabled: false, BackupRetention: 0,
		Region: region, MultiAZ: req.MultiAZ,
		CostPerMonth: float64(nodes) * 95.00,
		Status: "provisioning", Updated: "2026-04-18",
	}
	proDocumentDBStore.add(slug, db)
	apiCreated(w, toAPIProDocumentDB(db))
}

func apiWSDBNoSQLDocDBGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	d, found := proDocumentDBStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "Document database not found.")
		return
	}
	apiOK(w, toAPIProDocumentDB(d))
}

func apiWSDBNoSQLDocDBDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proDocumentDBStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "Document database not found.")
		return
	}
	proDocumentDBStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "Document database deleted."})
}
