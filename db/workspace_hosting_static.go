package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StaticSite is a row from workspace_hosting_static joined with resources.
type StaticSite struct {
	ResourceID      int64
	Name            string
	Status          string
	Region          string
	Framework       string
	Branch          string
	LastDeployAt    *time.Time
	BuildTimeSec    int
	CDNPoPs         int
	BandwidthGBMTD  float64
	CreatedAt       time.Time
}

// ListStaticSites returns all static site hosting resources in a workspace.
func ListStaticSites(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]StaticSite, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       h.framework, h.branch, h.last_deploy_at,
		       h.build_time_sec, h.cdn_pops, h.bandwidth_gb_mtd,
		       r.created_at
		FROM workspace_hosting_static h
		JOIN resources r ON r.id = h.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListStaticSites: %w", err)
	}
	defer rows.Close()
	var out []StaticSite
	for rows.Next() {
		var s StaticSite
		if err := rows.Scan(&s.ResourceID, &s.Name, &s.Status, &s.Region,
			&s.Framework, &s.Branch, &s.LastDeployAt,
			&s.BuildTimeSec, &s.CDNPoPs, &s.BandwidthGBMTD,
			&s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CreateStaticSite inserts a new static site hosting resource.
func CreateStaticSite(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, framework, branch, region string,
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
		VALUES ($1,$2,'hosting_static',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateStaticSite resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_hosting_static (resource_id, framework, branch)
		VALUES ($1,$2,$3)`,
		resourceID, framework, branch)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateStaticSite static row: %w", err)
	}

	return resourceID, nil
}
