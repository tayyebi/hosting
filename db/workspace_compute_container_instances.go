package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ContainerInstance is a row from workspace_compute_container_instances joined with resources.
type ContainerInstance struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	Image         string
	CPUMillicores int
	MemoryMB      int
	IPAddress     string
	RestartCount  int
	CreatedAt     time.Time
}

// ListContainerInstances returns all container instances in a workspace.
func ListContainerInstances(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ContainerInstance, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       ci.image, ci.cpu_millicores, ci.memory_mb,
		       ci.ip_address, ci.restart_count,
		       r.created_at
		FROM workspace_compute_container_instances ci
		JOIN resources r ON r.id = ci.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListContainerInstances: %w", err)
	}
	defer rows.Close()
	var out []ContainerInstance
	for rows.Next() {
		var c ContainerInstance
		if err := rows.Scan(&c.ResourceID, &c.Name, &c.Status, &c.Region,
			&c.Image, &c.CPUMillicores, &c.MemoryMB,
			&c.IPAddress, &c.RestartCount,
			&c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateContainerInstance inserts a new container instance resource.
func CreateContainerInstance(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, image, region string,
	cpuMillicores, memoryMB int,
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
		VALUES ($1,$2,'container_instance',$3,'pending',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateContainerInstance resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_compute_container_instances
		    (resource_id, image, cpu_millicores, memory_mb)
		VALUES ($1,$2,$3,$4)`,
		resourceID, image, cpuMillicores, memoryMB)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateContainerInstance instance row: %w", err)
	}

	return resourceID, nil
}
