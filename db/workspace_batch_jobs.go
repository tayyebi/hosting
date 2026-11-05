package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BatchJob is a row from workspace_batch_jobs joined with resources.
type BatchJob struct {
	ResourceID      int64
	Name            string
	Status          string
	Schedule        string
	LastRunAt       *time.Time
	LastDurationSec int
	NextRunAt       *time.Time
	CreatedAt       time.Time
}

// ListBatchJobs returns all batch jobs in a workspace.
func ListBatchJobs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]BatchJob, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       j.schedule, j.last_run_at, j.last_duration_sec, j.next_run_at,
		       r.created_at
		FROM workspace_batch_jobs j
		JOIN resources r ON r.id = j.resource_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListBatchJobs: %w", err)
	}
	defer rows.Close()
	var out []BatchJob
	for rows.Next() {
		var j BatchJob
		if err := rows.Scan(&j.ResourceID, &j.Name, &j.Status,
			&j.Schedule, &j.LastRunAt, &j.LastDurationSec, &j.NextRunAt,
			&j.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// CreateBatchJob inserts a new batch job resource.
func CreateBatchJob(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, schedule string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'batch_job',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateBatchJob resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_batch_jobs (resource_id, schedule)
		VALUES ($1,$2)`,
		resourceID, schedule)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateBatchJob job row: %w", err)
	}

	return resourceID, nil
}
