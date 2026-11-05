package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — Relational: MySQL / PostgreSQL
// databases/relational/mysql-pg
// ════════════════════════════════════════════════════════════════

type DBMySQLPG struct {
	ID        string
	Name      string
	Engine    string
	Version   string
	Region    string
	Status    string
	Size      string
	Storage   string
	Replicas  int
	Endpoint  string
	Encrypted bool
}

type dbMySQLPGStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBMySQLPG
}

var dbMySQLPGStore = &dbMySQLPGStoreType{
	instances: map[string][]DBMySQLPG{},
}

func (s *dbMySQLPGStoreType) list(slug string) []DBMySQLPG {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBMySQLPG, len(src))
	copy(out, src)
	return out
}

func (s *dbMySQLPGStoreType) getByID(slug, id string) (DBMySQLPG, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBMySQLPG{}, false
}

func (s *dbMySQLPGStoreType) add(slug string, d DBMySQLPG) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbMySQLPGStoreType) delete(slug, id string) {
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

type apiDBMySQLPG struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	Size      string `json:"size"`
	Storage   string `json:"storage"`
	Replicas  int    `json:"replicas"`
	Endpoint  string `json:"endpoint"`
	Encrypted bool   `json:"encrypted"`
}

type apiDBMySQLPGList struct {
	Instances []apiDBMySQLPG `json:"instances"`
	Total     int            `json:"total"`
}

func toAPIDBMySQLPG(d DBMySQLPG) apiDBMySQLPG {
	return apiDBMySQLPG{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Replicas: d.Replicas,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBMySQLPG(all []DBMySQLPG, search, status string) []DBMySQLPG {
	var out []DBMySQLPG
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Engine + " " + d.Version + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSDBRelationalMySQLPGListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBMySQLPG(dbMySQLPGStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBMySQLPG, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBMySQLPG(d)
	}
	apiOK(w, apiDBMySQLPGList{Instances: out, Total: len(out)})
}

type apiCreateDBMySQLPGRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Storage   string `json:"storage"`
	Replicas  int    `json:"replicas"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBRelationalMySQLPGCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBMySQLPGRequest
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
		engine = "postgresql"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		if engine == "mysql" {
			version = "8.0.36"
		} else {
			version = "16.2"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "db.m6g.large"
	}
	storage := strings.TrimSpace(req.Storage)
	if storage == "" {
		storage = "100 GB SSD"
	}
	port := "5432"
	if engine == "mysql" {
		port = "3306"
	}
	inst := DBMySQLPG{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "creating", Size: size,
		Storage: storage, Replicas: req.Replicas,
		Endpoint:  name + ".db." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbMySQLPGStore.add(slug, inst)
	apiCreated(w, toAPIDBMySQLPG(inst))
}

func apiWSDBRelationalMySQLPGGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbMySQLPGStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "Database instance not found.")
		return
	}
	apiOK(w, toAPIDBMySQLPG(inst))
}

func apiWSDBRelationalMySQLPGDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbMySQLPGStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "Database instance not found.")
		return
	}
	dbMySQLPGStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "Database instance deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — Relational: SQL Server
// databases/relational/sql-server
// ════════════════════════════════════════════════════════════════

type DBSQLServer struct {
	ID        string
	Name      string
	Edition   string
	Version   string
	Region    string
	Status    string
	Size      string
	Storage   string
	Replicas  int
	Endpoint  string
	Encrypted bool
}

type dbSQLServerStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBSQLServer
}

var dbSQLServerStore = &dbSQLServerStoreType{
	instances: map[string][]DBSQLServer{},
}

func (s *dbSQLServerStoreType) list(slug string) []DBSQLServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBSQLServer, len(src))
	copy(out, src)
	return out
}

func (s *dbSQLServerStoreType) getByID(slug, id string) (DBSQLServer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBSQLServer{}, false
}

func (s *dbSQLServerStoreType) add(slug string, d DBSQLServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbSQLServerStoreType) delete(slug, id string) {
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

type apiDBSQLServer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Edition   string `json:"edition"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	Size      string `json:"size"`
	Storage   string `json:"storage"`
	Replicas  int    `json:"replicas"`
	Endpoint  string `json:"endpoint"`
	Encrypted bool   `json:"encrypted"`
}

type apiDBSQLServerList struct {
	Instances []apiDBSQLServer `json:"instances"`
	Total     int              `json:"total"`
}

func toAPIDBSQLServer(d DBSQLServer) apiDBSQLServer {
	return apiDBSQLServer{
		ID: d.ID, Name: d.Name, Edition: d.Edition,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Replicas: d.Replicas,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBSQLServer(all []DBSQLServer, search, status string) []DBSQLServer {
	var out []DBSQLServer
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Edition + " " + d.Version + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSDBRelationalSQLServerListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBSQLServer(dbSQLServerStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBSQLServer, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBSQLServer(d)
	}
	apiOK(w, apiDBSQLServerList{Instances: out, Total: len(out)})
}

type apiCreateDBSQLServerRequest struct {
	Name      string `json:"name"`
	Edition   string `json:"edition"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Storage   string `json:"storage"`
	Replicas  int    `json:"replicas"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBRelationalSQLServerCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBSQLServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	edition := strings.TrimSpace(req.Edition)
	if edition == "" {
		edition = "standard"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "2022"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "db.m6i.large"
	}
	storage := strings.TrimSpace(req.Storage)
	if storage == "" {
		storage = "100 GB SSD"
	}
	inst := DBSQLServer{
		ID: newID(), Name: name, Edition: edition,
		Version: version, Region: region,
		Status: "creating", Size: size,
		Storage: storage, Replicas: req.Replicas,
		Endpoint:  name + ".db." + region + ".example.com:1433",
		Encrypted: req.Encrypted,
	}
	dbSQLServerStore.add(slug, inst)
	apiCreated(w, toAPIDBSQLServer(inst))
}

func apiWSDBRelationalSQLServerGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbSQLServerStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "SQL Server instance not found.")
		return
	}
	apiOK(w, toAPIDBSQLServer(inst))
}

func apiWSDBRelationalSQLServerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbSQLServerStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "SQL Server instance not found.")
		return
	}
	dbSQLServerStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "SQL Server instance deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — Relational: Cloud-native
// databases/relational/cloud-native
// ════════════════════════════════════════════════════════════════

type DBCloudNative struct {
	ID              string
	Name            string
	Engine          string
	Version         string
	Region          string
	Status          string
	Size            string
	Storage         string
	Replicas        int
	AutoScaling     bool
	ServerlessUnits string
	Endpoint        string
	Encrypted       bool
}

type dbCloudNativeStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBCloudNative
}

var dbCloudNativeStore = &dbCloudNativeStoreType{
	instances: map[string][]DBCloudNative{},
}

func (s *dbCloudNativeStoreType) list(slug string) []DBCloudNative {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBCloudNative, len(src))
	copy(out, src)
	return out
}

func (s *dbCloudNativeStoreType) getByID(slug, id string) (DBCloudNative, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBCloudNative{}, false
}

func (s *dbCloudNativeStoreType) add(slug string, d DBCloudNative) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbCloudNativeStoreType) delete(slug, id string) {
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

type apiDBCloudNative struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Engine          string `json:"engine"`
	Version         string `json:"version"`
	Region          string `json:"region"`
	Status          string `json:"status"`
	Size            string `json:"size"`
	Storage         string `json:"storage"`
	Replicas        int    `json:"replicas"`
	AutoScaling     bool   `json:"auto_scaling"`
	ServerlessUnits string `json:"serverless_units"`
	Endpoint        string `json:"endpoint"`
	Encrypted       bool   `json:"encrypted"`
}

type apiDBCloudNativeList struct {
	Instances []apiDBCloudNative `json:"instances"`
	Total     int                `json:"total"`
}

func toAPIDBCloudNative(d DBCloudNative) apiDBCloudNative {
	return apiDBCloudNative{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Replicas: d.Replicas,
		AutoScaling: d.AutoScaling, ServerlessUnits: d.ServerlessUnits,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBCloudNative(all []DBCloudNative, search, status string) []DBCloudNative {
	var out []DBCloudNative
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.Engine + " " + d.Version + " " + d.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSDBRelationalCloudNativeListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBCloudNative(dbCloudNativeStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBCloudNative, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBCloudNative(d)
	}
	apiOK(w, apiDBCloudNativeList{Instances: out, Total: len(out)})
}

type apiCreateDBCloudNativeRequest struct {
	Name            string `json:"name"`
	Engine          string `json:"engine"`
	Version         string `json:"version"`
	Region          string `json:"region"`
	Size            string `json:"size"`
	Replicas        int    `json:"replicas"`
	AutoScaling     bool   `json:"auto_scaling"`
	ServerlessUnits string `json:"serverless_units"`
	Encrypted       bool   `json:"encrypted"`
}

func apiWSDBRelationalCloudNativeCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBCloudNativeRequest
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
		engine = "aurora-postgresql"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "latest"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "db.r6g.large"
	}
	serverlessUnits := strings.TrimSpace(req.ServerlessUnits)
	if serverlessUnits == "" {
		serverlessUnits = "—"
	}
	port := "5432"
	if strings.Contains(engine, "mysql") {
		port = "3306"
	} else if strings.Contains(engine, "spanner") {
		port = "443"
	} else if strings.Contains(engine, "cockroach") {
		port = "26257"
	}
	inst := DBCloudNative{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "creating", Size: size,
		Storage: "auto (0 GB used)", Replicas: req.Replicas,
		AutoScaling: req.AutoScaling, ServerlessUnits: serverlessUnits,
		Endpoint:  name + ".cluster." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbCloudNativeStore.add(slug, inst)
	apiCreated(w, toAPIDBCloudNative(inst))
}

func apiWSDBRelationalCloudNativeGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbCloudNativeStore.getByID(slug, r.PathValue("dbid"))
	if !found {
		apiError(w, http.StatusNotFound, "Cloud-native database instance not found.")
		return
	}
	apiOK(w, toAPIDBCloudNative(inst))
}

func apiWSDBRelationalCloudNativeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbCloudNativeStore.getByID(slug, r.PathValue("dbid")); !found {
		apiError(w, http.StatusNotFound, "Cloud-native database instance not found.")
		return
	}
	dbCloudNativeStore.delete(slug, r.PathValue("dbid"))
	apiOK(w, map[string]any{"ok": true, "message": "Cloud-native database instance deleted."})
}
