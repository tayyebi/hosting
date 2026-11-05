package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BackgroundJob is one row from the background_jobs table.
type BackgroundJob struct {
	ID          int64
	TenantID    *int64
	Queue       string
	Kind        string
	Status      string
	Priority    int
	Attempts    int
	MaxAttempts int
	LastError   string
	ScheduledAt time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	WorkerID    *int64
	Payload     map[string]any
}

// JobQueue is one row from background_job_queues.
type JobQueue struct {
	Name        string
	Description string
	Concurrency int
	Paused      bool
}

// JobWorker is one row from background_job_workers.
type JobWorker struct {
	ID            int64
	Hostname      string
	PID           *int
	Queues        []string
	Status        string
	LastHeartbeat time.Time
	StartedAt     time.Time
}

// JobOpts controls filtering for ListJobs.
type JobOpts struct {
	Status string
	Queue  string
	Search string
	Limit  int
}

// ListJobs returns background_jobs rows with optional filters, newest-first.
func ListJobs(ctx context.Context, pool *pgxpool.Pool, opts JobOpts) ([]BackgroundJob, error) {
	if opts.Limit <= 0 || opts.Limit > 500 {
		opts.Limit = 200
	}

	args := []any{}
	conds := []string{"1=1"}

	if opts.Status != "" {
		args = append(args, opts.Status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}
	if opts.Queue != "" {
		args = append(args, opts.Queue)
		conds = append(conds, fmt.Sprintf("queue = $%d", len(args)))
	}
	if opts.Search != "" {
		args = append(args, "%"+strings.ToLower(opts.Search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf("(lower(kind) LIKE $%d OR lower(queue) LIKE $%d)", n, n))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, queue, kind, status, priority,
		       attempts, max_attempts, COALESCE(last_error,''),
		       scheduled_at, started_at, finished_at, worker_id, payload
		FROM background_jobs
		WHERE %s
		ORDER BY scheduled_at DESC
		LIMIT %d`, strings.Join(conds, " AND "), opts.Limit)

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListJobs: %w", err)
	}
	defer rows.Close()

	var out []BackgroundJob
	for rows.Next() {
		var j BackgroundJob
		if err := rows.Scan(
			&j.ID, &j.TenantID, &j.Queue, &j.Kind, &j.Status, &j.Priority,
			&j.Attempts, &j.MaxAttempts, &j.LastError,
			&j.ScheduledAt, &j.StartedAt, &j.FinishedAt, &j.WorkerID, &j.Payload,
		); err != nil {
			return nil, fmt.Errorf("db.ListJobs scan: %w", err)
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// JobStatusCounts returns count per status across all background_jobs.
func JobStatusCounts(ctx context.Context, pool *pgxpool.Pool) (map[string]int, error) {
	rows, err := pool.Query(ctx, `
		SELECT status, count(*)::int FROM background_jobs GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("db.JobStatusCounts: %w", err)
	}
	defer rows.Close()
	m := map[string]int{}
	for rows.Next() {
		var s string
		var n int
		if err := rows.Scan(&s, &n); err != nil {
			return nil, err
		}
		m[s] = n
	}
	return m, rows.Err()
}

// ListJobQueues returns all rows from background_job_queues.
func ListJobQueues(ctx context.Context, pool *pgxpool.Pool) ([]JobQueue, error) {
	rows, err := pool.Query(ctx, `
		SELECT name, COALESCE(description,''), concurrency, paused
		FROM background_job_queues
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("db.ListJobQueues: %w", err)
	}
	defer rows.Close()
	var out []JobQueue
	for rows.Next() {
		var q JobQueue
		if err := rows.Scan(&q.Name, &q.Description, &q.Concurrency, &q.Paused); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// ListJobWorkers returns all rows from background_job_workers.
func ListJobWorkers(ctx context.Context, pool *pgxpool.Pool) ([]JobWorker, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, hostname, pid, queues, status, last_heartbeat, started_at
		FROM background_job_workers
		ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("db.ListJobWorkers: %w", err)
	}
	defer rows.Close()
	var out []JobWorker
	for rows.Next() {
		var w JobWorker
		if err := rows.Scan(&w.ID, &w.Hostname, &w.PID, &w.Queues, &w.Status, &w.LastHeartbeat, &w.StartedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
