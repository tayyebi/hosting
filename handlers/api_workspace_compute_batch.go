package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Compute — Batch: Job Scheduling & HPC / Grid Computing
// ════════════════════════════════════════════════════════════════

// ═══════════════════════════════════════════════════════════════
// 1. Job Scheduling   compute/batch/jobs
// ═══════════════════════════════════════════════════════════════

type BatchJob struct {
	ID       string
	Name     string
	Queue    string
	Priority string
	Schedule string
	Region   string
	Status   string
	Runtime  string
	ExitCode string
	Retries  int
}

type batchJobStoreType struct {
	mu   sync.RWMutex
	jobs map[string][]BatchJob
}

var batchJobStore = &batchJobStoreType{
	jobs: map[string][]BatchJob{},
}

func (s *batchJobStoreType) list(slug string) []BatchJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.jobs[slug]
	out := make([]BatchJob, len(src))
	copy(out, src)
	return out
}

func (s *batchJobStoreType) getByID(slug, id string) (BatchJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs[slug] {
		if j.ID == id {
			return j, true
		}
	}
	return BatchJob{}, false
}

func (s *batchJobStoreType) add(slug string, j BatchJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[slug] = append(s.jobs[slug], j)
}

func (s *batchJobStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.jobs[slug]
	for i, j := range list {
		if j.ID == id {
			s.jobs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiBatchJob struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Queue    string `json:"queue"`
	Priority string `json:"priority"`
	Schedule string `json:"schedule"`
	Region   string `json:"region"`
	Status   string `json:"status"`
	Runtime  string `json:"runtime"`
	ExitCode string `json:"exit_code"`
	Retries  int    `json:"retries"`
}

type apiBatchJobList struct {
	Jobs  []apiBatchJob `json:"jobs"`
	Total int           `json:"total"`
}

func toAPIBatchJob(j BatchJob) apiBatchJob {
	return apiBatchJob{
		ID: j.ID, Name: j.Name, Queue: j.Queue,
		Priority: j.Priority, Schedule: j.Schedule, Region: j.Region,
		Status: j.Status, Runtime: j.Runtime, ExitCode: j.ExitCode, Retries: j.Retries,
	}
}

func filterBatchJobs(all []BatchJob, search, status string) []BatchJob {
	var out []BatchJob
	q := strings.ToLower(search)
	for _, j := range all {
		if status != "" && j.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(j.Name + " " + j.Queue + " " + j.Priority + " " + j.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func apiWSComputeBatchJobsListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	jobs := filterBatchJobs(batchJobStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiBatchJob, len(jobs))
	for i, j := range jobs {
		out[i] = toAPIBatchJob(j)
	}
	apiOK(w, apiBatchJobList{Jobs: out, Total: len(out)})
}

type apiCreateBatchJobRequest struct {
	Name     string `json:"name"`
	Queue    string `json:"queue"`
	Priority string `json:"priority"`
	Schedule string `json:"schedule"`
	Region   string `json:"region"`
}

func apiWSComputeBatchJobsCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateBatchJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	queue := strings.TrimSpace(req.Queue)
	if queue == "" {
		queue = "default"
	}
	priority := strings.TrimSpace(req.Priority)
	if priority == "" {
		priority = "medium"
	}
	schedule := strings.TrimSpace(req.Schedule)
	if schedule == "" {
		schedule = "manual"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	job := BatchJob{
		ID: newID(), Name: name, Queue: queue,
		Priority: priority, Schedule: schedule, Region: region,
		Status: "queued", Runtime: "—", ExitCode: "—", Retries: 0,
	}
	batchJobStore.add(slug, job)
	apiCreated(w, toAPIBatchJob(job))
}

func apiWSComputeBatchJobsGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	job, found := batchJobStore.getByID(slug, r.PathValue("jid"))
	if !found {
		apiError(w, http.StatusNotFound, "Batch job not found.")
		return
	}
	apiOK(w, toAPIBatchJob(job))
}

func apiWSComputeBatchJobsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := batchJobStore.getByID(slug, r.PathValue("jid")); !found {
		apiError(w, http.StatusNotFound, "Batch job not found.")
		return
	}
	batchJobStore.delete(slug, r.PathValue("jid"))
	apiOK(w, map[string]any{"ok": true, "message": "Batch job deleted."})
}

// ═══════════════════════════════════════════════════════════════
// 2. HPC / Grid Computing   compute/batch/hpc
// ═══════════════════════════════════════════════════════════════

type HPCJob struct {
	ID        string
	Name      string
	Cluster   string
	Nodes     int
	CPUsNode  int
	Partition string
	Region    string
	Status    string
	WallTime  string
	Elapsed   string
}

type hpcJobStoreType struct {
	mu   sync.RWMutex
	jobs map[string][]HPCJob
}

var hpcJobStore = &hpcJobStoreType{
	jobs: map[string][]HPCJob{},
}

func (s *hpcJobStoreType) list(slug string) []HPCJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.jobs[slug]
	out := make([]HPCJob, len(src))
	copy(out, src)
	return out
}

func (s *hpcJobStoreType) getByID(slug, id string) (HPCJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, j := range s.jobs[slug] {
		if j.ID == id {
			return j, true
		}
	}
	return HPCJob{}, false
}

func (s *hpcJobStoreType) add(slug string, j HPCJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[slug] = append(s.jobs[slug], j)
}

func (s *hpcJobStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.jobs[slug]
	for i, j := range list {
		if j.ID == id {
			s.jobs[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiHPCJob struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	Nodes     int    `json:"nodes"`
	CPUsNode  int    `json:"cpus_per_node"`
	Partition string `json:"partition"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	WallTime  string `json:"wall_time"`
	Elapsed   string `json:"elapsed"`
}

type apiHPCJobList struct {
	Jobs  []apiHPCJob `json:"jobs"`
	Total int         `json:"total"`
}

func toAPIHPCJob(j HPCJob) apiHPCJob {
	return apiHPCJob{
		ID: j.ID, Name: j.Name, Cluster: j.Cluster,
		Nodes: j.Nodes, CPUsNode: j.CPUsNode, Partition: j.Partition,
		Region: j.Region, Status: j.Status, WallTime: j.WallTime, Elapsed: j.Elapsed,
	}
}

func filterHPCJobs(all []HPCJob, search, status string) []HPCJob {
	var out []HPCJob
	q := strings.ToLower(search)
	for _, j := range all {
		if status != "" && j.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(j.Name + " " + j.Cluster + " " + j.Partition + " " + j.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func apiWSComputeBatchHPCListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	jobs := filterHPCJobs(hpcJobStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiHPCJob, len(jobs))
	for i, j := range jobs {
		out[i] = toAPIHPCJob(j)
	}
	apiOK(w, apiHPCJobList{Jobs: out, Total: len(out)})
}

type apiCreateHPCJobRequest struct {
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	Nodes     int    `json:"nodes"`
	CPUsNode  int    `json:"cpus_per_node"`
	Partition string `json:"partition"`
	Region    string `json:"region"`
	WallTime  string `json:"wall_time"`
}

func apiWSComputeBatchHPCCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateHPCJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	cluster := strings.TrimSpace(req.Cluster)
	if cluster == "" {
		cluster = "hpc-alpha"
	}
	nodes := req.Nodes
	if nodes <= 0 {
		nodes = 1
	}
	cpusNode := req.CPUsNode
	if cpusNode <= 0 {
		cpusNode = 64
	}
	partition := strings.TrimSpace(req.Partition)
	if partition == "" {
		partition = "compute"
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	wallTime := strings.TrimSpace(req.WallTime)
	if wallTime == "" {
		wallTime = "24:00:00"
	}
	job := HPCJob{
		ID: newID(), Name: name, Cluster: cluster,
		Nodes: nodes, CPUsNode: cpusNode, Partition: partition,
		Region: region, Status: "queued", WallTime: wallTime, Elapsed: "—",
	}
	hpcJobStore.add(slug, job)
	apiCreated(w, toAPIHPCJob(job))
}

func apiWSComputeBatchHPCGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	job, found := hpcJobStore.getByID(slug, r.PathValue("hid"))
	if !found {
		apiError(w, http.StatusNotFound, "HPC job not found.")
		return
	}
	apiOK(w, toAPIHPCJob(job))
}

func apiWSComputeBatchHPCDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := hpcJobStore.getByID(slug, r.PathValue("hid")); !found {
		apiError(w, http.StatusNotFound, "HPC job not found.")
		return
	}
	hpcJobStore.delete(slug, r.PathValue("hid"))
	apiOK(w, map[string]any{"ok": true, "message": "HPC job deleted."})
}
