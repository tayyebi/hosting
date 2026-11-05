package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — In-Memory: Session Store
// databases/memory/session
// ════════════════════════════════════════════════════════════════

type DBMemorySession struct {
	ID             string
	Name           string
	Engine         string
	Version        string
	Region         string
	Status         string
	Size           string
	Nodes          int
	MaxMemory      string
	TTL            string
	ActiveSessions string
	Throughput     string
	Endpoint       string
	Encrypted      bool
}

type dbMemorySessionStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBMemorySession
}

var dbMemorySessionStore = &dbMemorySessionStoreType{
	instances: map[string][]DBMemorySession{
		"buyvm": {
			{ID: newID(), Name: "web-sessions", Engine: "redis", Version: "7.2", Region: "eu-central-1", Status: "running", Size: "cache.r7g.xlarge", Nodes: 3, MaxMemory: "32 GB", TTL: "24h", ActiveSessions: "1,240,000", Throughput: "18,400 ops/s", Endpoint: "web-sessions.session.eu-central-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "api-tokens", Engine: "redis", Version: "7.2", Region: "eu-central-1", Status: "running", Size: "cache.r7g.large", Nodes: 2, MaxMemory: "16 GB", TTL: "1h", ActiveSessions: "320,000", Throughput: "8,200 ops/s", Endpoint: "api-tokens.session.eu-central-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "oauth-state", Engine: "redis", Version: "7.2", Region: "us-east-1", Status: "running", Size: "cache.m7g.large", Nodes: 2, MaxMemory: "8 GB", TTL: "10m", ActiveSessions: "45,000", Throughput: "3,600 ops/s", Endpoint: "oauth-state.session.us-east-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "cart-sessions", Engine: "memcached", Version: "1.6", Region: "eu-central-1", Status: "running", Size: "cache.m7g.xlarge", Nodes: 4, MaxMemory: "64 GB", TTL: "48h", ActiveSessions: "890,000", Throughput: "12,100 ops/s", Endpoint: "cart-sessions.session.eu-central-1.example.com:11211", Encrypted: false},
			{ID: newID(), Name: "dev-sessions", Engine: "redis", Version: "7.2", Region: "eu-central-1", Status: "stopped", Size: "cache.t3.micro", Nodes: 1, MaxMemory: "512 MB", TTL: "30m", ActiveSessions: "0", Throughput: "0 ops/s", Endpoint: "dev-sessions.session.eu-central-1.example.com:6379", Encrypted: false},
		},
		"hosting": {
			{ID: newID(), Name: "user-sessions", Engine: "redis", Version: "7.2", Region: "ap-south-1", Status: "running", Size: "cache.r7g.large", Nodes: 2, MaxMemory: "16 GB", TTL: "12h", ActiveSessions: "680,000", Throughput: "9,800 ops/s", Endpoint: "user-sessions.session.ap-south-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "admin-sessions", Engine: "redis", Version: "7.2", Region: "ap-south-1", Status: "running", Size: "cache.m7g.medium", Nodes: 1, MaxMemory: "4 GB", TTL: "2h", ActiveSessions: "1,200", Throughput: "240 ops/s", Endpoint: "admin-sessions.session.ap-south-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "websocket-state", Engine: "redis", Version: "7.2", Region: "ap-south-1", Status: "running", Size: "cache.m7g.large", Nodes: 2, MaxMemory: "8 GB", TTL: "6h", ActiveSessions: "42,000", Throughput: "5,400 ops/s", Endpoint: "websocket-state.session.ap-south-1.example.com:6379", Encrypted: true},
		},
	},
}

func (s *dbMemorySessionStoreType) list(slug string) []DBMemorySession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBMemorySession, len(src))
	copy(out, src)
	return out
}

func (s *dbMemorySessionStoreType) getByID(slug, id string) (DBMemorySession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBMemorySession{}, false
}

func (s *dbMemorySessionStoreType) add(slug string, d DBMemorySession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbMemorySessionStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.instances[slug]
	for i, d := range list {
		if d.ID == id {
			s.instances[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDBMemorySession struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Engine         string `json:"engine"`
	Version        string `json:"version"`
	Region         string `json:"region"`
	Status         string `json:"status"`
	Size           string `json:"size"`
	Nodes          int    `json:"nodes"`
	MaxMemory      string `json:"max_memory"`
	TTL            string `json:"ttl"`
	ActiveSessions string `json:"active_sessions"`
	Throughput     string `json:"throughput"`
	Endpoint       string `json:"endpoint"`
	Encrypted      bool   `json:"encrypted"`
}

type apiDBMemorySessionList struct {
	Instances []apiDBMemorySession `json:"instances"`
	Total     int                  `json:"total"`
}

func toAPIDBMemorySession(d DBMemorySession) apiDBMemorySession {
	return apiDBMemorySession{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size, Nodes: d.Nodes,
		MaxMemory: d.MaxMemory, TTL: d.TTL,
		ActiveSessions: d.ActiveSessions, Throughput: d.Throughput,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBMemorySession(all []DBMemorySession, search, status string) []DBMemorySession {
	var out []DBMemorySession
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Engine + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSDBMemorySessionListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBMemorySession(dbMemorySessionStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBMemorySession, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBMemorySession(d)
	}
	apiOK(w, apiDBMemorySessionList{Instances: out, Total: len(out)})
}

type apiCreateDBMemorySessionRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Nodes     int    `json:"nodes"`
	MaxMemory string `json:"max_memory"`
	TTL       string `json:"ttl"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBMemorySessionCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBMemorySessionRequest
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
		engine = "redis"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		if engine == "memcached" {
			version = "1.6"
		} else {
			version = "7.2"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "cache.m7g.medium"
	}
	nodes := req.Nodes
	if nodes <= 0 {
		nodes = 1
	}
	maxMem := strings.TrimSpace(req.MaxMemory)
	if maxMem == "" {
		maxMem = "4 GB"
	}
	ttl := strings.TrimSpace(req.TTL)
	if ttl == "" {
		ttl = "24h"
	}
	port := "6379"
	if engine == "memcached" {
		port = "11211"
	}
	inst := DBMemorySession{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size, Nodes: nodes,
		MaxMemory: maxMem, TTL: ttl,
		ActiveSessions: "0", Throughput: "0 ops/s",
		Endpoint:  name + ".session." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbMemorySessionStore.add(slug, inst)
	apiCreated(w, toAPIDBMemorySession(inst))
}

func apiWSDBMemorySessionGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbMemorySessionStore.getByID(slug, r.PathValue("sid"))
	if !found {
		apiError(w, http.StatusNotFound, "Session store not found.")
		return
	}
	apiOK(w, toAPIDBMemorySession(inst))
}

func apiWSDBMemorySessionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbMemorySessionStore.getByID(slug, r.PathValue("sid")); !found {
		apiError(w, http.StatusNotFound, "Session store not found.")
		return
	}
	dbMemorySessionStore.delete(slug, r.PathValue("sid"))
	apiOK(w, map[string]any{"ok": true, "message": "Session store deleted."})
}
