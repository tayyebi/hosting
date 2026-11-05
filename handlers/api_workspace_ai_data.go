package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Data: ETL / ELT Pipelines
// ai/data/etl
// ════════════════════════════════════════════════════════════════

type ETLPipeline struct {
	ID          string
	Name        string
	Status      string
	Type        string
	Source      string
	Destination string
	Schedule    string
	LastRun     string
	Duration    string
	Records     string
	Region      string
}

type etlPipelineStoreType struct {
	mu        sync.RWMutex
	pipelines map[string][]ETLPipeline
}

var etlPipelineStore = &etlPipelineStoreType{
	pipelines: map[string][]ETLPipeline{
		"buyvm": {
			{ID: newID(), Name: "user-events-pipeline", Status: "active", Type: "etl", Source: "Kafka: user-events", Destination: "S3: datalake/user-events/", Schedule: "*/15 * * * *", LastRun: "2026-04-17 08:45 UTC", Duration: "3m 22s", Records: "2.4M rows", Region: "eu-central-1"},
			{ID: newID(), Name: "order-sync", Status: "active", Type: "elt", Source: "PostgreSQL: orders", Destination: "BigQuery: analytics.orders", Schedule: "0 */2 * * *", LastRun: "2026-04-17 08:00 UTC", Duration: "8m 14s", Records: "180K rows", Region: "eu-central-1"},
			{ID: newID(), Name: "log-aggregation", Status: "active", Type: "etl", Source: "Fluentd: app-logs", Destination: "Elasticsearch: logs-*", Schedule: "*/5 * * * *", LastRun: "2026-04-17 09:10 UTC", Duration: "1m 48s", Records: "12M rows", Region: "us-east-1"},
			{ID: newID(), Name: "crm-data-sync", Status: "paused", Type: "elt", Source: "Salesforce API", Destination: "Snowflake: raw.crm", Schedule: "0 6 * * *", LastRun: "2026-04-15 06:00 UTC", Duration: "22m 10s", Records: "45K rows", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "payment-events", Status: "active", Type: "etl", Source: "Kafka: payments", Destination: "S3: datalake/payments/", Schedule: "*/10 * * * *", LastRun: "2026-04-17 09:00 UTC", Duration: "2m 05s", Records: "680K rows", Region: "ap-south-1"},
			{ID: newID(), Name: "inventory-sync", Status: "active", Type: "elt", Source: "MySQL: inventory", Destination: "Redshift: warehouse.inventory", Schedule: "0 */4 * * *", LastRun: "2026-04-17 08:00 UTC", Duration: "5m 30s", Records: "120K rows", Region: "ap-south-1"},
			{ID: newID(), Name: "clickstream-etl", Status: "failed", Type: "etl", Source: "Kinesis: clicks", Destination: "S3: datalake/clicks/", Schedule: "*/5 * * * *", LastRun: "2026-04-16 14:22 UTC", Duration: "\u2014", Records: "0 rows", Region: "ap-south-1"},
		},
	},
}

func (s *etlPipelineStoreType) list(slug string) []ETLPipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.pipelines[slug]
	out := make([]ETLPipeline, len(src))
	copy(out, src)
	return out
}

func (s *etlPipelineStoreType) getByID(slug, id string) (ETLPipeline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.pipelines[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return ETLPipeline{}, false
}

func (s *etlPipelineStoreType) add(slug string, p ETLPipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelines[slug] = append(s.pipelines[slug], p)
}

func (s *etlPipelineStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.pipelines[slug]
	for i, p := range list {
		if p.ID == id {
			s.pipelines[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiETLPipeline struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Schedule    string `json:"schedule"`
	LastRun     string `json:"last_run"`
	Duration    string `json:"duration"`
	Records     string `json:"records"`
	Region      string `json:"region"`
}

type apiETLPipelineList struct {
	Pipelines []apiETLPipeline `json:"pipelines"`
	Total     int              `json:"total"`
}

func toAPIETLPipeline(p ETLPipeline) apiETLPipeline {
	return apiETLPipeline{
		ID: p.ID, Name: p.Name, Status: p.Status,
		Type: p.Type, Source: p.Source,
		Destination: p.Destination, Schedule: p.Schedule,
		LastRun: p.LastRun, Duration: p.Duration,
		Records: p.Records, Region: p.Region,
	}
}

func filterETLPipelines(all []ETLPipeline, search, status string) []ETLPipeline {
	var out []ETLPipeline
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.Type + " " + p.Source + " " + p.Destination + " " + p.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSAIDataETLListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipelines := filterETLPipelines(etlPipelineStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiETLPipeline, len(pipelines))
	for i, p := range pipelines {
		out[i] = toAPIETLPipeline(p)
	}
	apiOK(w, apiETLPipelineList{Pipelines: out, Total: len(out)})
}

type apiCreateETLPipelineRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Schedule    string `json:"schedule"`
	Region      string `json:"region"`
}

func apiWSAIDataETLCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateETLPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		apiError(w, http.StatusBadRequest, "Source is required.")
		return
	}
	destination := strings.TrimSpace(req.Destination)
	if destination == "" {
		apiError(w, http.StatusBadRequest, "Destination is required.")
		return
	}
	pType := strings.TrimSpace(req.Type)
	if pType == "" {
		pType = "etl"
	}
	schedule := strings.TrimSpace(req.Schedule)
	if schedule == "" {
		schedule = "daily"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	pipeline := ETLPipeline{
		ID: newID(), Name: name, Status: "creating",
		Type: pType, Source: source,
		Destination: destination, Schedule: schedule,
		LastRun: "\u2014", Duration: "\u2014",
		Records: "0 rows", Region: region,
	}
	etlPipelineStore.add(slug, pipeline)
	apiCreated(w, toAPIETLPipeline(pipeline))
}

func apiWSAIDataETLGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	pipeline, found := etlPipelineStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "ETL pipeline not found.")
		return
	}
	apiOK(w, toAPIETLPipeline(pipeline))
}

func apiWSAIDataETLDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := etlPipelineStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "ETL pipeline not found.")
		return
	}
	etlPipelineStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "ETL pipeline deleted."})
}

// ════════════════════════════════════════════════════════════════
// Workspace AI API — Data: Streaming Ingestion
// ai/data/streaming
// ════════════════════════════════════════════════════════════════

type StreamIngestion struct {
	ID          string
	Name        string
	Status      string
	Engine      string
	Source      string
	Destination string
	Throughput  string
	Partitions  int
	Retention   string
	Lag         string
	Region      string
}

type streamIngestionStoreType struct {
	mu      sync.RWMutex
	streams map[string][]StreamIngestion
}

var streamIngestionStore = &streamIngestionStoreType{
	streams: map[string][]StreamIngestion{
		"buyvm": {
			{ID: newID(), Name: "real-time-events", Status: "active", Engine: "kafka", Source: "App: event-producer", Destination: "S3 + Elasticsearch", Throughput: "48K msg/s", Partitions: 12, Retention: "7 days", Lag: "120ms", Region: "eu-central-1"},
			{ID: newID(), Name: "iot-telemetry", Status: "active", Engine: "kinesis", Source: "IoT Gateway", Destination: "Timestream + S3", Throughput: "8.2K msg/s", Partitions: 8, Retention: "24 hours", Lag: "340ms", Region: "us-east-1"},
			{ID: newID(), Name: "change-data-capture", Status: "active", Engine: "kafka", Source: "PostgreSQL CDC", Destination: "Kafka Topics \u2192 Consumers", Throughput: "2.4K msg/s", Partitions: 6, Retention: "3 days", Lag: "45ms", Region: "eu-central-1"},
			{ID: newID(), Name: "click-stream", Status: "active", Engine: "pulsar", Source: "Web SDK", Destination: "ClickHouse", Throughput: "120K msg/s", Partitions: 24, Retention: "48 hours", Lag: "80ms", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), Name: "order-stream", Status: "active", Engine: "kafka", Source: "API Gateway", Destination: "DynamoDB + S3", Throughput: "12K msg/s", Partitions: 8, Retention: "7 days", Lag: "95ms", Region: "ap-south-1"},
			{ID: newID(), Name: "sensor-ingest", Status: "active", Engine: "kinesis", Source: "Edge Devices", Destination: "TimescaleDB", Throughput: "3.8K msg/s", Partitions: 4, Retention: "24 hours", Lag: "210ms", Region: "ap-south-1"},
		},
	},
}

func (s *streamIngestionStoreType) list(slug string) []StreamIngestion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.streams[slug]
	out := make([]StreamIngestion, len(src))
	copy(out, src)
	return out
}

func (s *streamIngestionStoreType) getByID(slug, id string) (StreamIngestion, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, st := range s.streams[slug] {
		if st.ID == id {
			return st, true
		}
	}
	return StreamIngestion{}, false
}

func (s *streamIngestionStoreType) add(slug string, st StreamIngestion) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams[slug] = append(s.streams[slug], st)
}

func (s *streamIngestionStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.streams[slug]
	for i, st := range list {
		if st.ID == id {
			s.streams[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiStreamIngestion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Engine      string `json:"engine"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Throughput  string `json:"throughput"`
	Partitions  int    `json:"partitions"`
	Retention   string `json:"retention"`
	Lag         string `json:"lag"`
	Region      string `json:"region"`
}

type apiStreamIngestionList struct {
	Streams []apiStreamIngestion `json:"streams"`
	Total   int                  `json:"total"`
}

func toAPIStreamIngestion(st StreamIngestion) apiStreamIngestion {
	return apiStreamIngestion{
		ID: st.ID, Name: st.Name, Status: st.Status,
		Engine: st.Engine, Source: st.Source,
		Destination: st.Destination, Throughput: st.Throughput,
		Partitions: st.Partitions, Retention: st.Retention,
		Lag: st.Lag, Region: st.Region,
	}
}

func filterStreamIngestions(all []StreamIngestion, search, status string) []StreamIngestion {
	var out []StreamIngestion
	q := strings.ToLower(search)
	for _, st := range all {
		if status != "" && st.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(st.Name + " " + st.Engine + " " + st.Source + " " + st.Destination + " " + st.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, st)
	}
	return out
}

func apiWSAIDataStreamListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	streams := filterStreamIngestions(streamIngestionStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiStreamIngestion, len(streams))
	for i, st := range streams {
		out[i] = toAPIStreamIngestion(st)
	}
	apiOK(w, apiStreamIngestionList{Streams: out, Total: len(out)})
}

type apiCreateStreamIngestionRequest struct {
	Name        string `json:"name"`
	Engine      string `json:"engine"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Region      string `json:"region"`
}

func apiWSAIDataStreamCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateStreamIngestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		apiError(w, http.StatusBadRequest, "Source is required.")
		return
	}
	destination := strings.TrimSpace(req.Destination)
	if destination == "" {
		apiError(w, http.StatusBadRequest, "Destination is required.")
		return
	}
	engine := strings.TrimSpace(req.Engine)
	if engine == "" {
		engine = "kafka"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	stream := StreamIngestion{
		ID: newID(), Name: name, Status: "provisioning",
		Engine: engine, Source: source,
		Destination: destination, Throughput: "0 msg/s",
		Partitions: 1, Retention: "24 hours",
		Lag: "\u2014", Region: region,
	}
	streamIngestionStore.add(slug, stream)
	apiCreated(w, toAPIStreamIngestion(stream))
}

func apiWSAIDataStreamGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	stream, found := streamIngestionStore.getByID(slug, r.PathValue("sid"))
	if !found {
		apiError(w, http.StatusNotFound, "Stream ingestion not found.")
		return
	}
	apiOK(w, toAPIStreamIngestion(stream))
}

func apiWSAIDataStreamDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := streamIngestionStore.getByID(slug, r.PathValue("sid")); !found {
		apiError(w, http.StatusNotFound, "Stream ingestion not found.")
		return
	}
	streamIngestionStore.delete(slug, r.PathValue("sid"))
	apiOK(w, map[string]any{"ok": true, "message": "Stream ingestion deleted."})
}
