package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EdgeFunction is a row from workspace_serverless_edge joined with resources.
type EdgeFunction struct {
	ResourceID     int64
	Name           string
	Status         string
	PathPattern    string
	PoPs           int
	RequestsPerSec int
	P99LatencyMS   int
	CreatedAt      time.Time
}

// ListEdgeFunctions returns all edge functions in a workspace.
func ListEdgeFunctions(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]EdgeFunction, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       e.path_pattern, e.pop_count, e.requests_per_sec, e.p99_latency_ms,
		       r.created_at
		FROM workspace_serverless_edge e
		JOIN resources r ON r.id = e.resource_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListEdgeFunctions: %w", err)
	}
	defer rows.Close()
	var out []EdgeFunction
	for rows.Next() {
		var e EdgeFunction
		if err := rows.Scan(&e.ResourceID, &e.Name, &e.Status,
			&e.PathPattern, &e.PoPs, &e.RequestsPerSec, &e.P99LatencyMS,
			&e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CreateEdgeFunction inserts a new edge function resource.
func CreateEdgeFunction(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, pathPattern string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'serverless_edge',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateEdgeFunction resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_serverless_edge (resource_id, path_pattern)
		VALUES ($1,$2)`,
		resourceID, pathPattern)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateEdgeFunction edge row: %w", err)
	}

	return resourceID, nil
}
