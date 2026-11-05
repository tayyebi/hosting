package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Databases API — Analytics
// ════════════════════════════════════════════════════════════════

// ═══════════════════════════════════════════════════════════════
// 1. Analytics — Data Warehouse   databases/analytics/warehouse
// ═══════════════════════════════════════════════════════════════

type DBWarehouse struct {
	ID         string
	Name       string
	Engine     string
	Version    string
	Region     string
	Status     string
	Size       string
	Storage    string
	Tables     int
	Queries24h string
	ScanBytes  string
	Endpoint   string
	Encrypted  bool
}

type dbWarehouseStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBWarehouse
}

var dbWarehouseStore = &dbWarehouseStoreType{
	instances: map[string][]DBWarehouse{},
}

func (s *dbWarehouseStoreType) list(slug string) []DBWarehouse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBWarehouse, len(src))
	copy(out, src)
	return out
}

func (s *dbWarehouseStoreType) getByID(slug, id string) (DBWarehouse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBWarehouse{}, false
}

func (s *dbWarehouseStoreType) add(slug string, d DBWarehouse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbWarehouseStoreType) delete(slug, id string) {
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

type apiDBWarehouse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Engine     string `json:"engine"`
	Version    string `json:"version"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Size       string `json:"size"`
	Storage    string `json:"storage"`
	Tables     int    `json:"tables"`
	Queries24h string `json:"queries_24h"`
	ScanBytes  string `json:"scan_bytes"`
	Endpoint   string `json:"endpoint"`
	Encrypted  bool   `json:"encrypted"`
}

type apiDBWarehouseList struct {
	Instances []apiDBWarehouse `json:"instances"`
	Total     int              `json:"total"`
}

func toAPIDBWarehouse(d DBWarehouse) apiDBWarehouse {
	return apiDBWarehouse{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Tables: d.Tables,
		Queries24h: d.Queries24h, ScanBytes: d.ScanBytes,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBWarehouse(all []DBWarehouse, search, status string) []DBWarehouse {
	var out []DBWarehouse
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

func apiWSDBAnalyticsWarehouseListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBWarehouse(dbWarehouseStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBWarehouse, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBWarehouse(d)
	}
	apiOK(w, apiDBWarehouseList{Instances: out, Total: len(out)})
}

type apiCreateDBWarehouseRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBAnalyticsWarehouseCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBWarehouseRequest
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
		engine = "clickhouse"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		switch engine {
		case "snowflake":
			version = "8.x"
		case "redshift":
			version = "1.0"
		default:
			version = "24.3"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		switch engine {
		case "snowflake":
			size = "Medium"
		case "redshift":
			size = "ra3.xlplus"
		default:
			size = "dw.medium"
		}
	}
	port := "8443"
	switch engine {
	case "snowflake":
		port = "443"
	case "redshift":
		port = "5439"
	}
	inst := DBWarehouse{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size,
		Storage: "0 B", Tables: 0,
		Queries24h: "0", ScanBytes: "0 B",
		Endpoint:  name + ".warehouse." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbWarehouseStore.add(slug, inst)
	apiCreated(w, toAPIDBWarehouse(inst))
}

func apiWSDBAnalyticsWarehouseGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbWarehouseStore.getByID(slug, r.PathValue("dwid"))
	if !found {
		apiError(w, http.StatusNotFound, "Data warehouse not found.")
		return
	}
	apiOK(w, toAPIDBWarehouse(inst))
}

func apiWSDBAnalyticsWarehouseDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbWarehouseStore.getByID(slug, r.PathValue("dwid")); !found {
		apiError(w, http.StatusNotFound, "Data warehouse not found.")
		return
	}
	dbWarehouseStore.delete(slug, r.PathValue("dwid"))
	apiOK(w, map[string]any{"ok": true, "message": "Data warehouse deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 2. Analytics — Time-Series   databases/analytics/timeseries
// ═══════════════════════════════════════════════════════════════

type DBTimeSeries struct {
	ID         string
	Name       string
	Engine     string
	Version    string
	Region     string
	Status     string
	Size       string
	Storage    string
	Metrics    string
	DataPoints string
	Retention  string
	Endpoint   string
	Encrypted  bool
}

type dbTimeSeriesStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBTimeSeries
}

var dbTimeSeriesStore = &dbTimeSeriesStoreType{
	instances: map[string][]DBTimeSeries{},
}

func (s *dbTimeSeriesStoreType) list(slug string) []DBTimeSeries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBTimeSeries, len(src))
	copy(out, src)
	return out
}

func (s *dbTimeSeriesStoreType) getByID(slug, id string) (DBTimeSeries, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBTimeSeries{}, false
}

func (s *dbTimeSeriesStoreType) add(slug string, d DBTimeSeries) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbTimeSeriesStoreType) delete(slug, id string) {
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

type apiDBTimeSeries struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Engine     string `json:"engine"`
	Version    string `json:"version"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Size       string `json:"size"`
	Storage    string `json:"storage"`
	Metrics    string `json:"metrics"`
	DataPoints string `json:"data_points"`
	Retention  string `json:"retention"`
	Endpoint   string `json:"endpoint"`
	Encrypted  bool   `json:"encrypted"`
}

type apiDBTimeSeriesList struct {
	Instances []apiDBTimeSeries `json:"instances"`
	Total     int               `json:"total"`
}

func toAPIDBTimeSeries(d DBTimeSeries) apiDBTimeSeries {
	return apiDBTimeSeries{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size,
		Storage: d.Storage, Metrics: d.Metrics,
		DataPoints: d.DataPoints, Retention: d.Retention,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBTimeSeries(all []DBTimeSeries, search, status string) []DBTimeSeries {
	var out []DBTimeSeries
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

func apiWSDBAnalyticsTimeSeriesListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBTimeSeries(dbTimeSeriesStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBTimeSeries, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBTimeSeries(d)
	}
	apiOK(w, apiDBTimeSeriesList{Instances: out, Total: len(out)})
}

type apiCreateDBTimeSeriesRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Retention string `json:"retention"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBAnalyticsTimeSeriesCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBTimeSeriesRequest
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
		engine = "influxdb"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		switch engine {
		case "timescaledb":
			version = "2.14"
		case "prometheus":
			version = "2.51"
		default:
			version = "3.0"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "ts.medium"
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "30d"
	}
	port := "8086"
	switch engine {
	case "timescaledb":
		port = "5432"
	case "prometheus":
		port = "9090"
	}
	inst := DBTimeSeries{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size,
		Storage: "0 B", Metrics: "0",
		DataPoints: "0", Retention: retention,
		Endpoint:  name + ".ts." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbTimeSeriesStore.add(slug, inst)
	apiCreated(w, toAPIDBTimeSeries(inst))
}

func apiWSDBAnalyticsTimeSeriesGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbTimeSeriesStore.getByID(slug, r.PathValue("tsid"))
	if !found {
		apiError(w, http.StatusNotFound, "Time-series database not found.")
		return
	}
	apiOK(w, toAPIDBTimeSeries(inst))
}

func apiWSDBAnalyticsTimeSeriesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbTimeSeriesStore.getByID(slug, r.PathValue("tsid")); !found {
		apiError(w, http.StatusNotFound, "Time-series database not found.")
		return
	}
	dbTimeSeriesStore.delete(slug, r.PathValue("tsid"))
	apiOK(w, map[string]any{"ok": true, "message": "Time-series database deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 3. Analytics — Search & Full-text   databases/analytics/search
// ═══════════════════════════════════════════════════════════════

type DBSearch struct {
	ID         string
	Name       string
	Engine     string
	Version    string
	Region     string
	Status     string
	Size       string
	Nodes      int
	Storage    string
	Indices    int
	Documents  string
	Queries24h string
	Endpoint   string
	Encrypted  bool
}

type dbSearchStoreType struct {
	mu        sync.RWMutex
	instances map[string][]DBSearch
}

var dbSearchStore = &dbSearchStoreType{
	instances: map[string][]DBSearch{},
}

func (s *dbSearchStoreType) list(slug string) []DBSearch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.instances[slug]
	out := make([]DBSearch, len(src))
	copy(out, src)
	return out
}

func (s *dbSearchStoreType) getByID(slug, id string) (DBSearch, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.instances[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DBSearch{}, false
}

func (s *dbSearchStoreType) add(slug string, d DBSearch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[slug] = append(s.instances[slug], d)
}

func (s *dbSearchStoreType) delete(slug, id string) {
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

type apiDBSearch struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Engine     string `json:"engine"`
	Version    string `json:"version"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	Size       string `json:"size"`
	Nodes      int    `json:"nodes"`
	Storage    string `json:"storage"`
	Indices    int    `json:"indices"`
	Documents  string `json:"documents"`
	Queries24h string `json:"queries_24h"`
	Endpoint   string `json:"endpoint"`
	Encrypted  bool   `json:"encrypted"`
}

type apiDBSearchList struct {
	Instances []apiDBSearch `json:"instances"`
	Total     int           `json:"total"`
}

func toAPIDBSearch(d DBSearch) apiDBSearch {
	return apiDBSearch{
		ID: d.ID, Name: d.Name, Engine: d.Engine,
		Version: d.Version, Region: d.Region,
		Status: d.Status, Size: d.Size, Nodes: d.Nodes,
		Storage: d.Storage, Indices: d.Indices,
		Documents: d.Documents, Queries24h: d.Queries24h,
		Endpoint: d.Endpoint, Encrypted: d.Encrypted,
	}
}

func filterDBSearch(all []DBSearch, search, status string) []DBSearch {
	var out []DBSearch
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

func apiWSDBAnalyticsSearchListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	instances := filterDBSearch(dbSearchStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDBSearch, len(instances))
	for i, d := range instances {
		out[i] = toAPIDBSearch(d)
	}
	apiOK(w, apiDBSearchList{Instances: out, Total: len(out)})
}

type apiCreateDBSearchRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Size      string `json:"size"`
	Nodes     int    `json:"nodes"`
	Encrypted bool   `json:"encrypted"`
}

func apiWSDBAnalyticsSearchCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDBSearchRequest
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
		engine = "elasticsearch"
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		switch engine {
		case "opensearch":
			version = "2.13"
		case "meilisearch":
			version = "1.7"
		default:
			version = "8.13"
		}
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = "search.medium"
	}
	nodes := req.Nodes
	if nodes <= 0 {
		nodes = 1
	}
	port := "9200"
	if engine == "meilisearch" {
		port = "7700"
	}
	inst := DBSearch{
		ID: newID(), Name: name, Engine: engine,
		Version: version, Region: region,
		Status: "provisioning", Size: size, Nodes: nodes,
		Storage: "0 B", Indices: 0,
		Documents: "0", Queries24h: "0",
		Endpoint:  name + ".search." + region + ".example.com:" + port,
		Encrypted: req.Encrypted,
	}
	dbSearchStore.add(slug, inst)
	apiCreated(w, toAPIDBSearch(inst))
}

func apiWSDBAnalyticsSearchGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	inst, found := dbSearchStore.getByID(slug, r.PathValue("seid"))
	if !found {
		apiError(w, http.StatusNotFound, "Search engine not found.")
		return
	}
	apiOK(w, toAPIDBSearch(inst))
}

func apiWSDBAnalyticsSearchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dbSearchStore.getByID(slug, r.PathValue("seid")); !found {
		apiError(w, http.StatusNotFound, "Search engine not found.")
		return
	}
	dbSearchStore.delete(slug, r.PathValue("seid"))
	apiOK(w, map[string]any{"ok": true, "message": "Search engine deleted."})
}
