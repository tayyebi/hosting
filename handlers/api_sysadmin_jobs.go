package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"platform/db"
)

// ── GET /api/sysadmin/jobs ──────────────────────────────────

type apiJob struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Queue      string `json:"queue"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	StartedAt  string `json:"started_at"`
	Duration   string `json:"duration"`
	Worker     string `json:"worker"`
	Tenant     string `json:"tenant"`
	Retries    int    `json:"retries"`
	MaxRetries int    `json:"max_retries"`
	Error      string `json:"error,omitempty"`
	Input      string `json:"input,omitempty"`
	Output     string `json:"output,omitempty"`
	TraceID    string `json:"trace_id"`
}

type apiQueueHealth struct {
	Name       string `json:"name"`
	Depth      int    `json:"depth"`
	Running    int    `json:"running"`
	Workers    int    `json:"workers"`
	Throughput int    `json:"throughput"`
	AvgWait    string `json:"avg_wait"`
	Healthy    bool   `json:"healthy"`
	LoadPct    int    `json:"load_pct"`
}

type apiWorkerInfo struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	CurrentJob string `json:"current_job,omitempty"`
	Queue      string `json:"queue"`
	JobsToday  int    `json:"jobs_today"`
	UptimePct  int    `json:"uptime_pct"`
}

type apiKPI struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	Unit     string `json:"unit"`
	Delta    string `json:"delta"`
	DeltaUp  bool   `json:"delta_up"`
	Positive bool   `json:"positive"`
}

type apiJobsResponse struct {
	Jobs           []apiJob         `json:"jobs"`
	RunningCount   int              `json:"running_count"`
	QueuedCount    int              `json:"queued_count"`
	FailedCount    int              `json:"failed_count"`
	CompletedCount int              `json:"completed_count"`
	CancelledCount int              `json:"cancelled_count"`
	TotalJobs      int              `json:"total_jobs"`
	QueueHealth    []apiQueueHealth `json:"queue_health"`
	Workers        []apiWorkerInfo  `json:"workers"`
	KPIs           []apiKPI         `json:"kpis"`
	Queues         []string         `json:"queues"`
	JobTypes       []string         `json:"job_types"`
}

func apiSysadminJobsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	q := r.URL.Query()
	filterStatus := q.Get("status")
	filterQueue := q.Get("queue")
	filterSearch := q.Get("q")

	resp := apiJobsResponse{}

	if Pools == nil {
		apiOK(w, resp)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	allDBJobs, _ := db.ListJobs(ctx, Pools.Admin, db.JobOpts{Limit: 500})
	allJobs := make([]Job, 0, len(allDBJobs))
	for _, j := range allDBJobs {
		allJobs = append(allJobs, dbJobToHandler(j))
	}

	var running, queued, failed, completed, cancelled int
	var runningJobs []Job
	for _, j := range allJobs {
		switch j.Status {
		case JobRunning:
			running++
			runningJobs = append(runningJobs, j)
		case JobQueued:
			queued++
		case JobFailed:
			failed++
		case JobCompleted:
			completed++
		case JobCancelled:
			cancelled++
		}
	}

	filtered := make([]Job, 0)
	for _, j := range allJobs {
		if filterStatus != "" && string(j.Status) != filterStatus {
			continue
		}
		if filterQueue != "" && j.Queue != filterQueue {
			continue
		}
		if filterSearch != "" {
			hay := strings.ToLower(j.Name + j.Queue + j.ID + j.TraceID)
			if !strings.Contains(hay, strings.ToLower(filterSearch)) {
				continue
			}
		}
		filtered = append(filtered, j)
	}

	dbQueues, _ := db.ListJobQueues(ctx, Pools.Admin)
	dbWorkers, _ := db.ListJobWorkers(ctx, Pools.Admin)

	queueNames := make([]string, 0, len(dbQueues))
	for _, q := range dbQueues {
		queueNames = append(queueNames, q.Name)
	}

	// Convert jobs
	apiJobs := make([]apiJob, len(filtered))
	for i, j := range filtered {
		apiJobs[i] = apiJob{
			ID:         j.ID,
			Name:       j.Name,
			Queue:      j.Queue,
			Status:     string(j.Status),
			Progress:   j.Progress,
			StartedAt:  j.StartedAt,
			Duration:   j.Duration,
			Worker:     j.Worker,
			Retries:    j.Retries,
			MaxRetries: j.MaxRetries,
			Error:      j.Error,
			Input:      j.Input,
			TraceID:    j.TraceID,
		}
	}

	// Queue health
	qh := queueHealthFromDB(allJobs, dbQueues, dbWorkers)
	apiQH := make([]apiQueueHealth, len(qh))
	for i, h := range qh {
		apiQH[i] = apiQueueHealth{h.Name, h.Depth, h.Running, h.Workers, h.Throughput, h.AvgWait, h.Healthy, h.LoadPct}
	}

	// Workers
	wk := workersFromDB(dbWorkers, runningJobs)
	apiWK := make([]apiWorkerInfo, len(wk))
	for i, w := range wk {
		apiWK[i] = apiWorkerInfo{
			ID:        w.ID,
			Status:    w.Status,
			Queue:     w.Queue,
			UptimePct: w.UptimePct,
		}
	}

	// KPIs
	kpis := kpisFromCounts(completed, failed, queued)
	apiKPIs := make([]apiKPI, len(kpis))
	for i, k := range kpis {
		apiKPIs[i] = apiKPI{k.Label, k.Value, k.Unit, k.Delta, k.DeltaUp, k.Positive}
	}

	// Job kinds
	kindSet := map[string]bool{}
	for _, j := range allJobs {
		kindSet[j.Name] = true
	}
	jobTypes := make([]string, 0, len(kindSet))
	for k := range kindSet {
		jobTypes = append(jobTypes, k)
	}

	resp = apiJobsResponse{
		Jobs:           apiJobs,
		RunningCount:   running,
		QueuedCount:    queued,
		FailedCount:    failed,
		CompletedCount: completed,
		CancelledCount: cancelled,
		TotalJobs:      len(allJobs),
		QueueHealth:    apiQH,
		Workers:        apiWK,
		KPIs:           apiKPIs,
		Queues:         queueNames,
		JobTypes:       jobTypes,
	}
	apiOK(w, resp)
}
