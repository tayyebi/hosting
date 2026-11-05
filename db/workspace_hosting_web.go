package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HostingWebApp is a row from workspace_hosting_web joined with resources.
type HostingWebApp struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Domain     string
	Runtime    string
	Stack      string
	Replicas   int
	MemoryMB   int
	CPUPct     int
	CreatedAt  time.Time
}

// ListHostingWebApps returns all web hosting apps in a workspace.
func ListHostingWebApps(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]HostingWebApp, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       h.domain, h.runtime, h.stack, h.replicas, h.memory_mb, h.cpu_pct,
		       r.created_at
		FROM workspace_hosting_web h
		JOIN resources r ON r.id = h.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListHostingWebApps: %w", err)
	}
	defer rows.Close()
	var out []HostingWebApp
	for rows.Next() {
		var h HostingWebApp
		if err := rows.Scan(&h.ResourceID, &h.Name, &h.Status, &h.Region,
			&h.Domain, &h.Runtime, &h.Stack, &h.Replicas, &h.MemoryMB, &h.CPUPct,
			&h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// CreateHostingWebApp inserts a new web hosting app resource.
func CreateHostingWebApp(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, domain, runtime, stack, region string,
	replicas, memoryMB int,
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
		VALUES ($1,$2,'hosting_web',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateHostingWebApp resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_hosting_web
		    (resource_id, domain, runtime, stack, replicas, memory_mb)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, domain, runtime, stack, replicas, memoryMB)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateHostingWebApp web row: %w", err)
	}

	return resourceID, nil
}
