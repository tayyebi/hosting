package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HPCJob is a row from workspace_batch_hpc joined with resources.
type HPCJob struct {
	ResourceID    int64
	Name          string
	Status        string
	UserName      string
	Framework     string
	Cores         int
	NodeCount     int
	QueueName     string
	ElapsedSec    int
	WalltimeSec   int
	CreatedAt     time.Time
}

// ListHPCJobs returns all HPC jobs in a workspace.
func ListHPCJobs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]HPCJob, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       h.user_name, h.framework, h.cores, h.node_count,
		       h.queue_name, h.elapsed_sec, h.walltime_sec,
		       r.created_at
		FROM workspace_batch_hpc h
		JOIN resources r ON r.id = h.resource_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListHPCJobs: %w", err)
	}
	defer rows.Close()
	var out []HPCJob
	for rows.Next() {
		var h HPCJob
		if err := rows.Scan(&h.ResourceID, &h.Name, &h.Status,
			&h.UserName, &h.Framework, &h.Cores, &h.NodeCount,
			&h.QueueName, &h.ElapsedSec, &h.WalltimeSec,
			&h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// CreateHPCJob inserts a new HPC job resource.
func CreateHPCJob(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, userName, framework, queueName string,
	cores, nodeCount int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'hpc_job',$3,'pending',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateHPCJob resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_batch_hpc
		    (resource_id, user_name, framework, queue_name, cores, node_count)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, userName, framework, queueName, cores, nodeCount)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateHPCJob hpc row: %w", err)
	}

	return resourceID, nil
}
