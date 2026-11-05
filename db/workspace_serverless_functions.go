package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ServerlessFunction is a row from workspace_serverless_functions joined with resources.
type ServerlessFunction struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Runtime        string
	TriggerType    string
	MemoryMB       int
	TimeoutSec     int
	Invocations1D  int
	Errors1D       int
	AvgDurationMS  int
	CreatedAt      time.Time
}

// ListServerlessFunctions returns all serverless functions in a workspace.
func ListServerlessFunctions(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ServerlessFunction, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       f.runtime, f.trigger_type, f.memory_mb, f.timeout_sec,
		       f.invocations_1d, f.errors_1d, f.avg_duration_ms,
		       r.created_at
		FROM workspace_serverless_functions f
		JOIN resources r ON r.id = f.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListServerlessFunctions: %w", err)
	}
	defer rows.Close()
	var out []ServerlessFunction
	for rows.Next() {
		var f ServerlessFunction
		if err := rows.Scan(&f.ResourceID, &f.Name, &f.Status, &f.Region,
			&f.Runtime, &f.TriggerType, &f.MemoryMB, &f.TimeoutSec,
			&f.Invocations1D, &f.Errors1D, &f.AvgDurationMS,
			&f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// CreateServerlessFunction inserts a new serverless function resource.
func CreateServerlessFunction(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, runtime, triggerType, region string,
	memoryMB, timeoutSec int,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'serverless_function',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateServerlessFunction resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_serverless_functions
		    (resource_id, runtime, trigger_type, memory_mb, timeout_sec)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, runtime, triggerType, memoryMB, timeoutSec)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateServerlessFunction function row: %w", err)
	}

	return resourceID, nil
}
