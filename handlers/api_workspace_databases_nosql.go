package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — NoSQL
// ════════════════════════════════════════════════════════════════

// ═══════════════════════════════════════════════════════════════
// 1. NoSQL — Document   databases/nosql/document
// ═══════════════════════════════════════════════════════════════

type DBDocument struct {
	ID          string
	Name        string
	Engine      string
	Version     string
	Region      string
	Status      string
	Size        string
	Storage     string
	Collections int
	Documents   string
	Endpoint    string
	Encrypted   bool
}

type dbDocumentStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBDocument
}

var dbDocumentStore = &dbDocumentStoreType{
	instances: map[string][]DBDocument{},
}

func (s *dbDocumentStoreType) list(slug string) []DBDocument {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBDocument, len(src))
	copy(out, src)
	return out
}

func (s *dbDocumentStoreType) getByID(slug, id string) (DBDocument, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBDocument{}, false
}

func (s *dbDocumentStoreType) add(slug string, d DBDocument) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbDocumentStoreType) delete(slug, id string) {
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

type apiDBDocument struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Engine      string `json:"engine"`
	Version     string `json:"version"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	Size        string `json:"size"`
	Storage     string `json:"storage"`
	Collections int    `json:"collections"`
	Documents   string `json:"documents"`
	Endpoint    string `json:"endpoint"`
	Encrypted   bool   `json:"encrypted"`
}

type apiDBDocumentList struct {
	Instances []apiDBDocument `json:"instances"`
	Total     int             `json:"total"`
}

func toAPIDBDocument(d DBDocument) apiDBDocument {
	return apiDBDocument{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Collections: d.Collections,
		Documents: d.Documents, Endpoint: d.Endpoint,
		Encrypted: d.Encrypted,
	}
}

func filterDBDocument(all []DBDocument, search, status string) []DBDocument {
	var out []DBDocument
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

func apiWSDBNoSQLDocumentListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBDocument(dbDocumentStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBDocument, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBDocument(d)
	}
	apiOK(w, apiDBDocumentList{Instances: out, Total: len(out)})
}

type apiCreateDBDocumentRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBNoSQLDocumentCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBDocumentRequest
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
		engine = "mongodb"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		if engine == "couchdb" {
			version = "3.3"
		} else {
			version = "7.0"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "M10"
	}
	port := "27017"
	if engine == "couchdb" {
		port = "5984"
	}
	inst := DBDocument{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size,
		Storage: "10 GB", Collections: 0, Documents: "0",
		Endpoint:  name + "." + engine[:5] + "." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbDocumentStore.add(slug, inst)
	apiCreated(w, toAPIDBDocument(inst))
}

func apiWSDBNoSQLDocumentGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbDocumentStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "Document database not found.")
		return
	}
	apiOK(w, toAPIDBDocument(inst))
}

func apiWSDBNoSQLDocumentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbDocumentStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "Document database not found.")
		return
	}
	dbDocumentStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "Document database deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 2. NoSQL — Key-Value   databases/nosql/keyvalue
// ═══════════════════════════════════════════════════════════════

type DBKeyValue struct {
	ID        string
	Name      string
	Engine    string
	Version   string
	Region    string
	Status    string
	Size      string
	Storage   string
	Keys      string
	MaxMemory string
	Endpoint  string
	Encrypted bool
}

type dbKeyValueStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBKeyValue
}

var dbKeyValueStore = &dbKeyValueStoreType{
	instances: map[string][]DBKeyValue{},
}

func (s *dbKeyValueStoreType) list(slug string) []DBKeyValue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBKeyValue, len(src))
	copy(out, src)
	return out
}

func (s *dbKeyValueStoreType) getByID(slug, id string) (DBKeyValue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBKeyValue{}, false
}

func (s *dbKeyValueStoreType) add(slug string, d DBKeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbKeyValueStoreType) delete(slug, id string) {
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

type apiDBKeyValue struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	Size      string `json:"size"`
	Storage   string `json:"storage"`
	Keys      string `json:"keys"`
	MaxMemory string `json:"max_memory"`
	Endpoint  string `json:"endpoint"`
	Encrypted bool   `json:"encrypted"`
}

type apiDBKeyValueList struct {
	Instances []apiDBKeyValue `json:"instances"`
	Total     int             `json:"total"`
}

func toAPIDBKeyValue(d DBKeyValue) apiDBKeyValue {
	return apiDBKeyValue{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Keys: d.Keys,
		MaxMemory: d.MaxMemory, Endpoint: d.Endpoint,
		Encrypted: d.Encrypted,
	}
}

func filterDBKeyValue(all []DBKeyValue, search, status string) []DBKeyValue {
	var out []DBKeyValue
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

func apiWSDBNoSQLKeyValueListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBKeyValue(dbKeyValueStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBKeyValue, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBKeyValue(d)
	}
	apiOK(w, apiDBKeyValueList{Instances: out, Total: len(out)})
}

type apiCreateDBKeyValueRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	MaxMemory string `json:"max_memory"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBNoSQLKeyValueCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBKeyValueRequest
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
		switch engine {
		case "etcd":
			version = "3.5"
		case "consul":
			version = "1.18"
		default:
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
	maxMem := strings.TrimSpace(req.MaxMemory)
	if maxMem == "" {
		maxMem = "4 GB"
	}
	port := "6379"
	switch engine {
	case "etcd":
		port = "2379"
	case "consul":
		port = "8500"
	}
	inst := DBKeyValue{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size,
		Storage: "—", Keys: "0", MaxMemory: maxMem,
		Endpoint:  name + ".kv." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbKeyValueStore.add(slug, inst)
	apiCreated(w, toAPIDBKeyValue(inst))
}

func apiWSDBNoSQLKeyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbKeyValueStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "Key-value store not found.")
		return
	}
	apiOK(w, toAPIDBKeyValue(inst))
}

func apiWSDBNoSQLKeyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbKeyValueStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "Key-value store not found.")
		return
	}
	dbKeyValueStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "Key-value store deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 3. NoSQL — Wide-Column   databases/nosql/wide-column
// ═══════════════════════════════════════════════════════════════

type DBWideColumn struct {
	ID         string
	Name       string
	Engine     string
	Version    string
	Region     string
	Status     string
	Nodes      int
	Storage    string
	Keyspaces  int
	Throughput string
	Endpoint   string
	Encrypted  bool
}

type dbWideColumnStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBWideColumn
}

var dbWideColumnStore = &dbWideColumnStoreType{
	instances: map[string][]DBWideColumn{},
}

func (s *dbWideColumnStoreType) list(slug string) []DBWideColumn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBWideColumn, len(src))
	copy(out, src)
	return out
}

func (s *dbWideColumnStoreType) getByID(slug, id string) (DBWideColumn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBWideColumn{}, false
}

func (s *dbWideColumnStoreType) add(slug string, d DBWideColumn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbWideColumnStoreType) delete(slug, id string) {
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

type apiDBWideColumn struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Engine     string `json:"engine"`
	Version    string `json:"version"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Nodes      int    `json:"nodes"`
	Storage    string `json:"storage"`
	Keyspaces  int    `json:"keyspaces"`
	Throughput string `json:"throughput"`
	Endpoint   string `json:"endpoint"`
	Encrypted  bool   `json:"encrypted"`
}

type apiDBWideColumnList struct {
	Instances []apiDBWideColumn `json:"instances"`
	Total     int               `json:"total"`
}

func toAPIDBWideColumn(d DBWideColumn) apiDBWideColumn {
	return apiDBWideColumn{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Nodes: d.Nodes,
		Storage: d.Storage, Keyspaces: d.Keyspaces,
		Throughput: d.Throughput, Endpoint: d.Endpoint,
		Encrypted: d.Encrypted,
	}
}

func filterDBWideColumn(all []DBWideColumn, search, status string) []DBWideColumn {
	var out []DBWideColumn
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

func apiWSDBNoSQLWideColumnListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBWideColumn(dbWideColumnStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBWideColumn, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBWideColumn(d)
	}
	apiOK(w, apiDBWideColumnList{Instances: out, Total: len(out)})
}

type apiCreateDBWideColumnRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Nodes     int    `json:"nodes"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBNoSQLWideColumnCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBWideColumnRequest
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
		engine = "cassandra"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		switch engine {
		case "scylladb":
			version = "5.4"
		case "hbase":
			version = "2.5"
		default:
			version = "4.1"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	nodes := req.Nodes
	if nodes <= 0 {
		nodes = 3
	}
	port := "9042"
	if engine == "hbase" {
		port = "16010"
	}
	inst := DBWideColumn{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Nodes: nodes,
		Storage: "0 B", Keyspaces: 0, Throughput: "0 ops/s",
		Endpoint:  name + ".wc." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbWideColumnStore.add(slug, inst)
	apiCreated(w, toAPIDBWideColumn(inst))
}

func apiWSDBNoSQLWideColumnGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbWideColumnStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "Wide-column database not found.")
		return
	}
	apiOK(w, toAPIDBWideColumn(inst))
}

func apiWSDBNoSQLWideColumnDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbWideColumnStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "Wide-column database not found.")
		return
	}
	dbWideColumnStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "Wide-column database deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 4. NoSQL — Graph   databases/nosql/graph
// ═══════════════════════════════════════════════════════════════

type DBGraph struct {
	ID        string
	Name      string
	Engine    string
	Version   string
	Region    string
	Status    string
	Size      string
	Storage   string
	Nodes     string
	Edges     string
	Endpoint  string
	Encrypted bool
}

type dbGraphStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBGraph
}

var dbGraphStore = &dbGraphStoreType{
	instances: map[string][]DBGraph{},
}

func (s *dbGraphStoreType) list(slug string) []DBGraph {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBGraph, len(src))
	copy(out, src)
	return out
}

func (s *dbGraphStoreType) getByID(slug, id string) (DBGraph, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBGraph{}, false
}

func (s *dbGraphStoreType) add(slug string, d DBGraph) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbGraphStoreType) delete(slug, id string) {
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

type apiDBGraph struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	Size      string `json:"size"`
	Storage   string `json:"storage"`
	Nodes     string `json:"nodes"`
	Edges     string `json:"edges"`
	Endpoint  string `json:"endpoint"`
	Encrypted bool   `json:"encrypted"`
}

type apiDBGraphList struct {
	Instances []apiDBGraph `json:"instances"`
	Total     int          `json:"total"`
}

func toAPIDBGraph(d DBGraph) apiDBGraph {
	return apiDBGraph{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Nodes: d.Nodes,
		Edges: d.Edges, Endpoint: d.Endpoint,
		Encrypted: d.Encrypted,
	}
}

func filterDBGraph(all []DBGraph, search, status string) []DBGraph {
	var out []DBGraph
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

func apiWSDBNoSQLGraphListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBGraph(dbGraphStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBGraph, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBGraph(d)
	}
	apiOK(w, apiDBGraphList{Instances: out, Total: len(out)})
}

type apiCreateDBGraphRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBNoSQLGraphCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBGraphRequest
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
		engine = "neo4j"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		if engine == "neptune" {
			version = "1.3"
		} else {
			version = "5.18"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		if engine == "neptune" {
			size = "db.r6g.large"
		} else {
			size = "community.small"
		}
	}
	port := "7687"
	if engine == "neptune" {
		port = "8182"
	}
	inst := DBGraph{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size,
		Storage: "0 B", Nodes: "0", Edges: "0",
		Endpoint:  name + ".graph." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbGraphStore.add(slug, inst)
	apiCreated(w, toAPIDBGraph(inst))
}

func apiWSDBNoSQLGraphGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbGraphStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "Graph database not found.")
		return
	}
	apiOK(w, toAPIDBGraph(inst))
}

func apiWSDBNoSQLGraphDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbGraphStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "Graph database not found.")
		return
	}
	dbGraphStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "Graph database deleted."})
}
