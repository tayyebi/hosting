package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ContainerRegistry is a row from workspace_compute_registry joined with resources.
type ContainerRegistry struct {
	ResourceID   int64
	Name         string
	Status       string
	RepoName     string
	TagCount     int
	SizeBytes    int64
	LastScanAt   *time.Time
	VulnCritical int
	VulnHigh     int
	VulnMedium   int
	CreatedAt    time.Time
}

// ListContainerRegistries returns all container registry repos in a workspace.
func ListContainerRegistries(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ContainerRegistry, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       cr.repo_name, cr.tag_count, cr.size_bytes, cr.last_scan_at,
		       cr.vuln_critical, cr.vuln_high, cr.vuln_medium,
		       r.created_at
		FROM workspace_compute_registry cr
		JOIN resources r ON r.id = cr.resource_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListContainerRegistries: %w", err)
	}
	defer rows.Close()
	var out []ContainerRegistry
	for rows.Next() {
		var c ContainerRegistry
		if err := rows.Scan(&c.ResourceID, &c.Name, &c.Status,
			&c.RepoName, &c.TagCount, &c.SizeBytes, &c.LastScanAt,
			&c.VulnCritical, &c.VulnHigh, &c.VulnMedium,
			&c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateContainerRegistry inserts a new container registry repository.
func CreateContainerRegistry(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, repoName string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'container_registry',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateContainerRegistry resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_compute_registry (resource_id, repo_name)
		VALUES ($1,$2)`,
		resourceID, repoName)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateContainerRegistry registry row: %w", err)
	}

	return resourceID, nil
}
