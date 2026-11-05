package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BlockVolume is a row from workspace_block_volumes joined with resources.
type BlockVolume struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	VolumeCategory string
	VolumeType     string
	SizeGB         int
	IOPS           int
	ThroughputMBPS int
	UtilizationPct int
	AttachedTo     string
	Encrypted      bool
	CreatedAt      time.Time
}

// ListBlockVolumes returns all block volume resources in a workspace filtered by category ('ssd' or 'hdd').
func ListBlockVolumes(ctx context.Context, pool *pgxpool.Pool, workspaceID int64, category string) ([]BlockVolume, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       bv.volume_category, bv.volume_type,
		       bv.size_gb, bv.iops, bv.throughput_mbps,
		       bv.utilization_pct, bv.attached_to, bv.encrypted,
		       r.created_at
		FROM workspace_block_volumes bv
		JOIN resources r ON r.id = bv.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND bv.volume_category = $2
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID, category)
	if err != nil {
		return nil, fmt.Errorf("db.ListBlockVolumes: %w", err)
	}
	defer rows.Close()
	var out []BlockVolume
	for rows.Next() {
		var v BlockVolume
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.VolumeCategory, &v.VolumeType,
			&v.SizeGB, &v.IOPS, &v.ThroughputMBPS,
			&v.UtilizationPct, &v.AttachedTo, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// CreateBlockVolume inserts a new block volume resource.
func CreateBlockVolume(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, volumeCategory, volumeType, region string,
	sizeGB, iops, throughputMBPS int,
	encrypted bool,
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
		VALUES ($1,$2,'block_volume',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateBlockVolume resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_block_volumes
		    (resource_id, volume_category, volume_type, size_gb, iops, throughput_mbps, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, volumeCategory, volumeType, sizeGB, iops, throughputMBPS, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateBlockVolume volume row: %w", err)
	}

	return resourceID, nil
}
