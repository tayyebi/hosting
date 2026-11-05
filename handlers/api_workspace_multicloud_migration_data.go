package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Multi-Cloud & Hybrid API — Migration: Data Migration
// multicloud/migration/data
// ════════════════════════════════════════════════════════════════

type DataMigrationJob struct {
	ID             string
	Name           string
	SourceStore    string
	TargetStore    string
	DataType       string
	TotalSize      string
	Transferred    string
	ProgressPct    string
	Throughput     string
	Validation     string
	StartedAt      string
	Status         string
	Updated        string
}

type dataMigrationJobStoreType struct {
	mu    sync.RWMutex
	items map[string][]DataMigrationJob
}

var dataMigrationJobStore = &dataMigrationJobStoreType{
	items: map[string][]DataMigrationJob{
		"buyvm": {
			{ID: newID(), Name: "prod-db-to-rds", SourceStore: "On-prem PostgreSQL", TargetStore: "AWS RDS PostgreSQL", DataType: "Relational", TotalSize: "2.4 TB", Transferred: "2.4 TB", ProgressPct: "100%", Throughput: "—", Validation: "Checksums passed", StartedAt: "2026-04-14 02:00", Status: "completed", Updated: "2026-04-15"},
			{ID: newID(), Name: "blob-storage-sync", SourceStore: "Azure Blob Storage", TargetStore: "AWS S3", DataType: "Object", TotalSize: "18.7 TB", Transferred: "12.3 TB", ProgressPct: "66%", Throughput: "850 MB/s", Validation: "In progress", StartedAt: "2026-04-16 08:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "analytics-warehouse", SourceStore: "On-prem Hadoop", TargetStore: "GCP BigQuery", DataType: "Analytics", TotalSize: "45.2 TB", Transferred: "8.1 TB", ProgressPct: "18%", Throughput: "1.2 GB/s", Validation: "Pending", StartedAt: "2026-04-17 00:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "redis-cache-migrate", SourceStore: "Self-hosted Redis", TargetStore: "AWS ElastiCache", DataType: "Key-value", TotalSize: "128 GB", Transferred: "128 GB", ProgressPct: "100%", Throughput: "—", Validation: "Key count matched", StartedAt: "2026-04-16 14:00", Status: "completed", Updated: "2026-04-16"},
			{ID: newID(), Name: "document-db-sync", SourceStore: "MongoDB Atlas", TargetStore: "AWS DocumentDB", DataType: "Document", TotalSize: "890 GB", Transferred: "0 GB", ProgressPct: "0%", Throughput: "—", Validation: "Not started", StartedAt: "—", Status: "pending", Updated: "2026-04-17"},
		},
		"hosting": {
			{ID: newID(), Name: "multi-region-s3-sync", SourceStore: "AWS S3 us-east-1", TargetStore: "AWS S3 eu-west-1 + ap-south-1", DataType: "Object", TotalSize: "120 TB", Transferred: "95 TB", ProgressPct: "79%", Throughput: "2.1 GB/s", Validation: "Incremental check", StartedAt: "2026-04-10 00:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "gcp-to-aws-bigdata", SourceStore: "GCP Cloud Storage", TargetStore: "AWS S3 + Glue", DataType: "Data lake", TotalSize: "68 TB", Transferred: "68 TB", ProgressPct: "100%", Throughput: "—", Validation: "Row counts verified", StartedAt: "2026-04-05 06:00", Status: "completed", Updated: "2026-04-12"},
			{ID: newID(), Name: "alibaba-oss-migration", SourceStore: "Alibaba OSS", TargetStore: "AWS S3", DataType: "Object", TotalSize: "8.5 TB", Transferred: "3.2 TB", ProgressPct: "38%", Throughput: "420 MB/s", Validation: "In progress", StartedAt: "2026-04-16 20:00", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "timeseries-influx-to-ts", SourceStore: "InfluxDB", TargetStore: "AWS Timestream", DataType: "Time-series", TotalSize: "1.8 TB", Transferred: "0 TB", ProgressPct: "0%", Throughput: "—", Validation: "Not started", StartedAt: "—", Status: "pending", Updated: "2026-04-17"},
		},
	},
}

func (s *dataMigrationJobStoreType) list(slug string) []DataMigrationJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]DataMigrationJob, len(src))
	copy(out, src)
	return out
}

func (s *dataMigrationJobStoreType) getByID(slug, id string) (DataMigrationJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.items[slug] {
		if d.ID == id {
			return d, true
		}
	}
	return DataMigrationJob{}, false
}

func (s *dataMigrationJobStoreType) add(slug string, d DataMigrationJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], d)
}

func (s *dataMigrationJobStoreType) delete(slug, id string) {
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

type apiDataMigrationJob struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SourceStore string `json:"source_store"`
	TargetStore string `json:"target_store"`
	DataType    string `json:"data_type"`
	TotalSize   string `json:"total_size"`
	Transferred string `json:"transferred"`
	ProgressPct string `json:"progress_pct"`
	Throughput  string `json:"throughput"`
	Validation  string `json:"validation"`
	StartedAt   string `json:"started_at"`
	Status      string `json:"status"`
	Updated     string `json:"updated"`
}

type apiDataMigrationJobList struct {
	Jobs  []apiDataMigrationJob `json:"jobs"`
	Total int                   `json:"total"`
}

func toAPIDataMigrationJob(d DataMigrationJob) apiDataMigrationJob {
	return apiDataMigrationJob{
		ID: d.ID, Name: d.Name, SourceStore: d.SourceStore,
		TargetStore: d.TargetStore, DataType: d.DataType,
		TotalSize: d.TotalSize, Transferred: d.Transferred,
		ProgressPct: d.ProgressPct, Throughput: d.Throughput,
		Validation: d.Validation, StartedAt: d.StartedAt,
		Status: d.Status, Updated: d.Updated,
	}
}

func filterDataMigrationJobs(all []DataMigrationJob, search, status string) []DataMigrationJob {
	var out []DataMigrationJob
	q := strings.ToLower(search)
	for _, d := range all {
		if status != "" && d.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(d.Name + " " + d.SourceStore + " " + d.TargetStore + " " + d.DataType)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, d)
	}
	return out
}

func apiWSMulticloudMigrationDataListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterDataMigrationJobs(dataMigrationJobStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDataMigrationJob, len(items))
	for i, d := range items {
		out[i] = toAPIDataMigrationJob(d)
	}
	apiOK(w, apiDataMigrationJobList{Jobs: out, Total: len(out)})
}

type apiCreateDataMigrationJobRequest struct {
	Name        string `json:"name"`
	SourceStore string `json:"source_store"`
	TargetStore string `json:"target_store"`
	DataType    string `json:"data_type"`
	TotalSize   string `json:"total_size"`
}

func apiWSMulticloudMigrationDataCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDataMigrationJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	srcStore := strings.TrimSpace(req.SourceStore)
	if srcStore == "" {
		srcStore = "On-premises"
	}
	tgtStore := strings.TrimSpace(req.TargetStore)
	if tgtStore == "" {
		tgtStore = "AWS S3"
	}
	dataType := strings.TrimSpace(req.DataType)
	if dataType == "" {
		dataType = "Object"
	}
	totalSize := strings.TrimSpace(req.TotalSize)
	if totalSize == "" {
		totalSize = "Unknown"
	}
	job := DataMigrationJob{
		ID: newID(), Name: name, SourceStore: srcStore,
		TargetStore: tgtStore, DataType: dataType,
		TotalSize: totalSize, Transferred: "0 GB",
		ProgressPct: "0%", Throughput: "—",
		Validation: "Not started", StartedAt: "—",
		Status: "pending", Updated: "2026-04-17",
	}
	dataMigrationJobStore.add(slug, job)
	apiCreated(w, toAPIDataMigrationJob(job))
}

func apiWSMulticloudMigrationDataGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	d, found := dataMigrationJobStore.getByID(slug, r.PathValue("did"))
	if !found {
		apiError(w, http.StatusNotFound, "Data migration job not found.")
		return
	}
	apiOK(w, toAPIDataMigrationJob(d))
}

func apiWSMulticloudMigrationDataDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := dataMigrationJobStore.getByID(slug, r.PathValue("did")); !found {
		apiError(w, http.StatusNotFound, "Data migration job not found.")
		return
	}
	dataMigrationJobStore.delete(slug, r.PathValue("did"))
	apiOK(w, map[string]any{"ok": true, "message": "Data migration job deleted."})
}
