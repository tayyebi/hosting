package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — In-Memory: Cache
// databases/memory/cache
// ════════════════════════════════════════════════════════════════

type DBMemoryCache struct {
	ID          string
	Name        string
	Engine      string
	Version     string
	Region      string
	Status      string
	Size        string
	Nodes       int
	MaxMemory   string
	Eviction    string
	HitRate     string
	Connections int
	Endpoint    string
	Encrypted   bool
}

type dbMemoryCacheStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBMemoryCache
}

var dbMemoryCacheStore = &dbMemoryCacheStoreType{
	instances: map[string][]DBMemoryCache{
		"buyvm": {
			{ID: newID(), Name: "api-response-cache", Engine: "redis", Version: "7.2", Region: "eu-central-1", Status: "running", Size: "cache.r7g.xlarge", Nodes: 3, MaxMemory: "32 GB", Eviction: "allkeys-lru", HitRate: "94.2%", Connections: 842, Endpoint: "api-response-cache.cache.eu-central-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "session-store", Engine: "redis", Version: "7.2", Region: "eu-central-1", Status: "running", Size: "cache.r7g.large", Nodes: 2, MaxMemory: "16 GB", Eviction: "volatile-lru", HitRate: "98.8%", Connections: 1240, Endpoint: "session-store.cache.eu-central-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "object-cache", Engine: "memcached", Version: "1.6", Region: "eu-central-1", Status: "running", Size: "cache.m7g.xlarge", Nodes: 4, MaxMemory: "64 GB", Eviction: "lru", HitRate: "91.5%", Connections: 320, Endpoint: "object-cache.cache.eu-central-1.example.com:11211", Encrypted: false},
			{ID: newID(), Name: "query-cache-us", Engine: "redis", Version: "7.2", Region: "us-east-1", Status: "running", Size: "cache.r7g.large", Nodes: 2, MaxMemory: "16 GB", Eviction: "allkeys-lfu", HitRate: "88.4%", Connections: 560, Endpoint: "query-cache-us.cache.us-east-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "dev-cache", Engine: "redis", Version: "7.2", Region: "eu-central-1", Status: "stopped", Size: "cache.t3.micro", Nodes: 1, MaxMemory: "512 MB", Eviction: "noeviction", HitRate: "—", Connections: 0, Endpoint: "dev-cache.cache.eu-central-1.example.com:6379", Encrypted: false},
		},
		"hosting": {
			{ID: newID(), Name: "app-cache-primary", Engine: "redis", Version: "7.2", Region: "ap-south-1", Status: "running", Size: "cache.r7g.large", Nodes: 2, MaxMemory: "16 GB", Eviction: "allkeys-lru", HitRate: "92.1%", Connections: 680, Endpoint: "app-cache-primary.cache.ap-south-1.example.com:6379", Encrypted: true},
			{ID: newID(), Name: "page-cache", Engine: "memcached", Version: "1.6", Region: "ap-south-1", Status: "running", Size: "cache.m7g.large", Nodes: 2, MaxMemory: "16 GB", Eviction: "lru", HitRate: "89.7%", Connections: 180, Endpoint: "page-cache.cache.ap-south-1.example.com:11211", Encrypted: false},
			{ID: newID(), Name: "token-cache", Engine: "redis", Version: "7.2", Region: "ap-south-1", Status: "running", Size: "cache.m7g.medium", Nodes: 1, MaxMemory: "4 GB", Eviction: "volatile-ttl", HitRate: "99.1%", Connections: 420, Endpoint: "token-cache.cache.ap-south-1.example.com:6379", Encrypted: true},
		},
	},
}

func (s *dbMemoryCacheStoreType) list(slug string) []DBMemoryCache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBMemoryCache, len(src))
	copy(out, src)
	return out
}

func (s *dbMemoryCacheStoreType) getByID(slug, id string) (DBMemoryCache, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBMemoryCache{}, false
}

func (s *dbMemoryCacheStoreType) add(slug string, d DBMemoryCache) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbMemoryCacheStoreType) delete(slug, id string) {
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

type apiDBMemoryCache struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Engine      string `json:"engine"`
	Version     string `json:"version"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Size        string `json:"size"`
	Nodes       int    `json:"nodes"`
	MaxMemory   string `json:"max_memory"`
	Eviction    string `json:"eviction"`
	HitRate     string `json:"hit_rate"`
	Connections int    `json:"connections"`
	Endpoint    string `json:"endpoint"`
	Encrypted   bool   `json:"encrypted"`
}

type apiDBMemoryCacheList struct {
	Instances []apiDBMemoryCache `json:"instances"`
	Total     int                `json:"total"`
}

func toAPIDBMemoryCache(d DBMemoryCache) apiDBMemoryCache {
	return apiDBMemoryCache{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size, Nodes: d.Nodes,
		MaxMemory: d.MaxMemory, Eviction: d.Eviction,
		HitRate: d.HitRate, Connections: d.Connections,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBMemoryCache(all []DBMemoryCache, search, status string) []DBMemoryCache {
	var out []DBMemoryCache
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

func apiWSDBMemoryCacheListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBMemoryCache(dbMemoryCacheStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBMemoryCache, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBMemoryCache(d)
	}
	apiOK(w, apiDBMemoryCacheList{Instances: out, Total: len(out)})
}

type apiCreateDBMemoryCacheRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Nodes     int    `json:"nodes"`
	MaxMemory string `json:"max_memory"`
	Eviction  string `json:"eviction"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBMemoryCacheCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBMemoryCacheRequest
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
	eviction := strings.TrimSpace(req.Eviction)
	if eviction == "" {
		if engine == "memcached" {
			eviction = "lru"
		} else {
			eviction = "allkeys-lru"
		}
	}
	port := "6379"
	if engine == "memcached" {
		port = "11211"
	}
	inst := DBMemoryCache{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size, Nodes: nodes,
		MaxMemory: maxMem, Eviction: eviction,
		HitRate: "—", Connections: 0,
		Endpoint: name + ".cache." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbMemoryCacheStore.add(slug, inst)
	apiCreated(w, toAPIDBMemoryCache(inst))
}

func apiWSDBMemoryCacheGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbMemoryCacheStore.getByID(slug, r.PathValue("cid"))
	if !found {
		apiError(w, http.StatusNotFound, "Cache instance not found.")
		return
	}
	apiOK(w, toAPIDBMemoryCache(inst))
}

func apiWSDBMemoryCacheDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbMemoryCacheStore.getByID(slug, r.PathValue("cid")); !found {
		apiError(w, http.StatusNotFound, "Cache instance not found.")
		return
	}
	dbMemoryCacheStore.delete(slug, r.PathValue("cid"))
	apiOK(w, map[string]any{"ok": true, "message": "Cache instance deleted."})
}
