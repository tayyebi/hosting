package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── File Shares (NFS / SMB) ────────────────────────────────────────────────

type FileShare struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Protocol     string
	SizeGB       int
	UsedGB       int
	MountPoint   string
	AccessCIDR   string
	ActiveMounts int
	CreatedAt    time.Time
}

func ListFileShares(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]FileShare, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       fs.protocol, fs.size_gb, fs.used_gb,
		       fs.mount_point, fs.access_cidr, fs.active_mounts,
		       r.created_at
		FROM workspace_file_shares fs
		JOIN resources r ON r.id = fs.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListFileShares: %w", err)
	}
	defer rows.Close()
	var out []FileShare
	for rows.Next() {
		var v FileShare
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Protocol, &v.SizeGB, &v.UsedGB,
			&v.MountPoint, &v.AccessCIDR, &v.ActiveMounts,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateFileShare(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, protocol, region string,
	sizeGB int, accessCIDR, mountPoint string,
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
		VALUES ($1,$2,'file_share',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateFileShare resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_file_shares
		    (resource_id, protocol, size_gb, mount_point, access_cidr)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, protocol, sizeGB, mountPoint, accessCIDR)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateFileShare share row: %w", err)
	}
	return resourceID, nil
}

// ── Distributed File Systems ───────────────────────────────────────────────

type DistributedFilesystem struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	FSType        string
	CapacityTB    float64
	UsedTB        float64
	NodeCount     int
	Replication   int
	MountPoint    string
	ActiveClients int
	CreatedAt     time.Time
}

func ListDistributedFilesystems(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DistributedFilesystem, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       dfs.fs_type, dfs.capacity_tb, dfs.used_tb,
		       dfs.node_count, dfs.replication,
		       dfs.mount_point, dfs.active_clients,
		       r.created_at
		FROM workspace_distributed_filesystems dfs
		JOIN resources r ON r.id = dfs.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDistributedFilesystems: %w", err)
	}
	defer rows.Close()
	var out []DistributedFilesystem
	for rows.Next() {
		var v DistributedFilesystem
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.FSType, &v.CapacityTB, &v.UsedTB,
			&v.NodeCount, &v.Replication,
			&v.MountPoint, &v.ActiveClients,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDistributedFilesystem(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, fsType, region string,
	capacityTB float64, nodeCount, replication int,
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
		VALUES ($1,$2,'distributed_fs',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDistributedFilesystem resource: %w", err)
	}
	mountPoint := "/mnt/" + name
	if fsType == "HDFS" {
		mountPoint = "hdfs://namenode:8020/" + name
	} else if fsType == "MinIO" {
		mountPoint = "s3://minio.internal/" + name
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_distributed_filesystems
		    (resource_id, fs_type, capacity_tb, node_count, replication, mount_point)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, fsType, capacityTB, nodeCount, replication, mountPoint)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDistributedFilesystem fs row: %w", err)
	}
	return resourceID, nil
}

// ── Backup Snapshots ───────────────────────────────────────────────────────

type BackupSnapshot struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	SnapshotType  string
	SourceKind    string
	SourceName    string
	SizeGB        float64
	RetentionDays int
	Encrypted     bool
	VerifiedAt    *time.Time
	CreatedAt     time.Time
}

func ListBackupSnapshots(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]BackupSnapshot, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       sn.snapshot_type, sn.source_kind, sn.source_name,
		       sn.size_gb, sn.retention_days, sn.encrypted, sn.verified_at,
		       r.created_at
		FROM workspace_snapshots sn
		JOIN resources r ON r.id = sn.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListBackupSnapshots: %w", err)
	}
	defer rows.Close()
	var out []BackupSnapshot
	for rows.Next() {
		var v BackupSnapshot
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.SnapshotType, &v.SourceKind, &v.SourceName,
			&v.SizeGB, &v.RetentionDays, &v.Encrypted, &v.VerifiedAt,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateBackupSnapshot(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, snapshotType, sourceKind, sourceName, region string,
	retentionDays int, encrypted bool,
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
		VALUES ($1,$2,'snapshot',$3,'in-progress',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateBackupSnapshot resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_snapshots
		    (resource_id, snapshot_type, source_kind, source_name, retention_days, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, snapshotType, sourceKind, sourceName, retentionDays, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateBackupSnapshot snapshot row: %w", err)
	}
	return resourceID, nil
}

// ── Disaster Recovery Plans ────────────────────────────────────────────────

type DRPlanDB struct {
	ResourceID        int64
	Name              string
	Status            string
	Region            string
	SourceRegion      string
	TargetRegion      string
	RPOMinutes        int
	RTOMinutes        int
	ReplicationType   string
	ProtectedCount    int
	ReplicationLagSec int
	Encrypted         bool
	LastTestedAt      *time.Time
	CreatedAt         time.Time
}

func ListDRPlans(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DRPlanDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       dr.source_region, dr.target_region,
		       dr.rpo_minutes, dr.rto_minutes,
		       dr.replication_type, dr.protected_count,
		       dr.replication_lag_s, dr.encrypted,
		       dr.last_tested_at,
		       r.created_at
		FROM workspace_dr_plans dr
		JOIN resources r ON r.id = dr.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDRPlans: %w", err)
	}
	defer rows.Close()
	var out []DRPlanDB
	for rows.Next() {
		var v DRPlanDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.SourceRegion, &v.TargetRegion,
			&v.RPOMinutes, &v.RTOMinutes,
			&v.ReplicationType, &v.ProtectedCount,
			&v.ReplicationLagSec, &v.Encrypted,
			&v.LastTestedAt,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDRPlan(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, sourceRegion, targetRegion, replicationType string,
	rpoMinutes, rtoMinutes int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if sourceRegion != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, sourceRegion).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'dr_plan',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDRPlan resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_dr_plans
		    (resource_id, source_region, target_region, rpo_minutes, rto_minutes,
		     replication_type, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, sourceRegion, targetRegion, rpoMinutes, rtoMinutes,
		replicationType, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDRPlan dr row: %w", err)
	}
	return resourceID, nil
}

// ── Long-term Archive Jobs ─────────────────────────────────────────────────

type ArchiveJob struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	ArchiveClass  string
	ArchiveType   string
	SourceName    string
	SizeGB        float64
	RetentionDays int
	Encrypted     bool
	VerifiedAt    *time.Time
	CreatedAt     time.Time
}

func ListArchiveJobs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ArchiveJob, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       aj.archive_class, aj.archive_type, aj.source_name,
		       aj.size_gb, aj.retention_days, aj.encrypted, aj.verified_at,
		       r.created_at
		FROM workspace_archive_jobs aj
		JOIN resources r ON r.id = aj.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListArchiveJobs: %w", err)
	}
	defer rows.Close()
	var out []ArchiveJob
	for rows.Next() {
		var v ArchiveJob
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.ArchiveClass, &v.ArchiveType, &v.SourceName,
			&v.SizeGB, &v.RetentionDays, &v.Encrypted, &v.VerifiedAt,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateArchiveJob(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, archiveClass, archiveType, sourceName, region string,
	sizeGB float64, retentionDays int, encrypted bool,
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
		VALUES ($1,$2,'archive_job',$3,'archiving',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateArchiveJob resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_archive_jobs
		    (resource_id, archive_class, archive_type, source_name,
		     size_gb, retention_days, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, archiveClass, archiveType, sourceName,
		sizeGB, retentionDays, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateArchiveJob archive row: %w", err)
	}
	return resourceID, nil
}
