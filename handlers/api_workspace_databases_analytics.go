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
	instances: map[string][]DBWarehouse{
		"buyvm": {
			{ID: newID(), Name: "analytics-main", Engine: "clickhouse", Version: "24.3", Region: "eu-central-1", Status: "running", Size: "dw.x2large", Storage: "12 TB", Tables: 248, Queries24h: "18,400", ScanBytes: "4.2 TB", Endpoint: "analytics-main.warehouse.eu-central-1.example.com:8443", Encrypted: true},
			{ID: newID(), Name: "bi-reporting", Engine: "snowflake", Version: "8.x", Region: "eu-central-1", Status: "running", Size: "X-Large", Storage: "8 TB", Tables: 142, Queries24h: "6,200", ScanBytes: "1.8 TB", Endpoint: "bi-reporting.warehouse.eu-central-1.example.com:443", Encrypted: true},
			{ID: newID(), Name: "events-warehouse", Engine: "clickhouse", Version: "24.3", Region: "us-east-1", Status: "running", Size: "dw.xlarge", Storage: "24 TB", Tables: 86, Queries24h: "42,000", ScanBytes: "9.6 TB", Endpoint: "events-warehouse.warehouse.us-east-1.example.com:8443", Encrypted: true},
			{ID: newID(), Name: "finance-dwh", Engine: "redshift", Version: "1.0", Region: "eu-central-1", Status: "running", Size: "ra3.xlplus", Storage: "4 TB", Tables: 320, Queries24h: "2,800", ScanBytes: "800 GB", Endpoint: "finance-dwh.warehouse.eu-central-1.example.com:5439", Encrypted: true},
			{ID: newID(), Name: "staging-dwh", Engine: "clickhouse", Version: "24.3", Region: "eu-central-1", Status: "stopped", Size: "dw.medium", Storage: "500 GB", Tables: 42, Queries24h: "0", ScanBytes: "0 B", Endpoint: "staging-dwh.warehouse.eu-central-1.example.com:8443", Encrypted: false},
		},
		"hosting": {
			{ID: newID(), Name: "product-analytics", Engine: "clickhouse", Version: "24.3", Region: "ap-south-1", Status: "running", Size: "dw.xlarge", Storage: "6 TB", Tables: 94, Queries24h: "8,400", ScanBytes: "2.1 TB", Endpoint: "product-analytics.warehouse.ap-south-1.example.com:8443", Encrypted: true},
			{ID: newID(), Name: "customer-insights", Engine: "snowflake", Version: "8.x", Region: "ap-south-1", Status: "running", Size: "Large", Storage: "3 TB", Tables: 68, Queries24h: "3,200", ScanBytes: "900 GB", Endpoint: "customer-insights.warehouse.ap-south-1.example.com:443", Encrypted: true},
			{ID: newID(), Name: "log-analytics", Engine: "clickhouse", Version: "24.3", Region: "ap-south-1", Status: "running", Size: "dw.large", Storage: "18 TB", Tables: 32, Queries24h: "24,000", ScanBytes: "5.4 TB", Endpoint: "log-analytics.warehouse.ap-south-1.example.com:8443", Encrypted: true},
		},
	},
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
	instances: map[string][]DBTimeSeries{
		"buyvm": {
			{ID: newID(), Name: "infra-metrics", Engine: "influxdb", Version: "3.0", Region: "eu-central-1", Status: "running", Size: "ts.xlarge", Storage: "2 TB", Metrics: "48,000", DataPoints: "12.4 B", Retention: "90d", Endpoint: "infra-metrics.ts.eu-central-1.example.com:8086", Encrypted: true},
			{ID: newID(), Name: "app-telemetry", Engine: "timescaledb", Version: "2.14", Region: "eu-central-1", Status: "running", Size: "ts.large", Storage: "800 GB", Metrics: "12,000", DataPoints: "4.8 B", Retention: "30d", Endpoint: "app-telemetry.ts.eu-central-1.example.com:5432", Encrypted: true},
			{ID: newID(), Name: "iot-sensor-data", Engine: "influxdb", Version: "3.0", Region: "us-east-1", Status: "running", Size: "ts.x2large", Storage: "8 TB", Metrics: "120,000", DataPoints: "48.2 B", Retention: "365d", Endpoint: "iot-sensor-data.ts.us-east-1.example.com:8086", Encrypted: true},
			{ID: newID(), Name: "business-kpis", Engine: "prometheus", Version: "2.51", Region: "eu-central-1", Status: "running", Size: "ts.medium", Storage: "200 GB", Metrics: "8,400", DataPoints: "2.1 B", Retention: "15d", Endpoint: "business-kpis.ts.eu-central-1.example.com:9090", Encrypted: false},
			{ID: newID(), Name: "dev-metrics", Engine: "influxdb", Version: "3.0", Region: "eu-central-1", Status: "stopped", Size: "ts.small", Storage: "50 GB", Metrics: "2,400", DataPoints: "180 M", Retention: "7d", Endpoint: "dev-metrics.ts.eu-central-1.example.com:8086", Encrypted: false},
		},
		"hosting": {
			{ID: newID(), Name: "platform-metrics", Engine: "influxdb", Version: "3.0", Region: "ap-south-1", Status: "running", Size: "ts.large", Storage: "1.5 TB", Metrics: "32,000", DataPoints: "8.6 B", Retention: "60d", Endpoint: "platform-metrics.ts.ap-south-1.example.com:8086", Encrypted: true},
			{ID: newID(), Name: "cdn-analytics", Engine: "timescaledb", Version: "2.14", Region: "ap-south-1", Status: "running", Size: "ts.large", Storage: "600 GB", Metrics: "6,800", DataPoints: "3.2 B", Retention: "30d", Endpoint: "cdn-analytics.ts.ap-south-1.example.com:5432", Encrypted: true},
			{ID: newID(), Name: "network-telemetry", Engine: "prometheus", Version: "2.51", Region: "ap-south-1", Status: "running", Size: "ts.medium", Storage: "300 GB", Metrics: "14,200", DataPoints: "4.4 B", Retention: "15d", Endpoint: "network-telemetry.ts.ap-south-1.example.com:9090", Encrypted: true},
		},
	},
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
	instances: map[string][]DBSearch{
		"buyvm": {
			{ID: newID(), Name: "product-search", Engine: "elasticsearch", Version: "8.13", Region: "eu-central-1", Status: "running", Size: "search.xlarge", Nodes: 5, Storage: "2 TB", Indices: 12, Documents: "24,000,000", Queries24h: "840,000", Endpoint: "product-search.search.eu-central-1.example.com:9200", Encrypted: true},
			{ID: newID(), Name: "log-search", Engine: "opensearch", Version: "2.13", Region: "eu-central-1", Status: "running", Size: "search.x2large", Nodes: 8, Storage: "12 TB", Indices: 48, Documents: "180,000,000", Queries24h: "120,000", Endpoint: "log-search.search.eu-central-1.example.com:9200", Encrypted: true},
			{ID: newID(), Name: "content-index", Engine: "elasticsearch", Version: "8.13", Region: "us-east-1", Status: "running", Size: "search.large", Nodes: 3, Storage: "500 GB", Indices: 8, Documents: "6,200,000", Queries24h: "280,000", Endpoint: "content-index.search.us-east-1.example.com:9200", Encrypted: true},
			{ID: newID(), Name: "autocomplete", Engine: "meilisearch", Version: "1.7", Region: "eu-central-1", Status: "running", Size: "search.medium", Nodes: 2, Storage: "80 GB", Indices: 4, Documents: "1,200,000", Queries24h: "2,400,000", Endpoint: "autocomplete.search.eu-central-1.example.com:7700", Encrypted: true},
			{ID: newID(), Name: "dev-search", Engine: "elasticsearch", Version: "8.13", Region: "eu-central-1", Status: "stopped", Size: "search.small", Nodes: 1, Storage: "20 GB", Indices: 3, Documents: "45,000", Queries24h: "0", Endpoint: "dev-search.search.eu-central-1.example.com:9200", Encrypted: false},
		},
		"hosting": {
			{ID: newID(), Name: "app-search", Engine: "elasticsearch", Version: "8.13", Region: "ap-south-1", Status: "running", Size: "search.large", Nodes: 3, Storage: "800 GB", Indices: 10, Documents: "14,000,000", Queries24h: "520,000", Endpoint: "app-search.search.ap-south-1.example.com:9200", Encrypted: true},
			{ID: newID(), Name: "docs-search", Engine: "meilisearch", Version: "1.7", Region: "ap-south-1", Status: "running", Size: "search.medium", Nodes: 1, Storage: "40 GB", Indices: 6, Documents: "480,000", Queries24h: "86,000", Endpoint: "docs-search.search.ap-south-1.example.com:7700", Encrypted: true},
			{ID: newID(), Name: "audit-search", Engine: "opensearch", Version: "2.13", Region: "ap-south-1", Status: "running", Size: "search.large", Nodes: 3, Storage: "4 TB", Indices: 24, Documents: "62,000,000", Queries24h: "18,000", Endpoint: "audit-search.search.ap-south-1.example.com:9200", Encrypted: true},
		},
	},
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
