package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Multi-Cloud Migration: Data Migration ─────────────────────────────────────

type MulticloudDataMigrationJobDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	SourceStore string
	TargetStore string
	DataType    string
	TotalSize   string
	Transferred string
	ProgressPct string
	Throughput  string
	Validation  string
	StartedAt   string
	JobStatus   string
	CreatedAt   time.Time
}

func ListMulticloudDataMigrationJobs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudDataMigrationJobDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       j.source_store, j.target_store, j.data_type,
		       j.total_size, j.transferred, j.progress_pct,
		       j.throughput, j.validation, j.started_at, j.job_status,
		       r.created_at
		FROM workspace_multicloud_data_migration_jobs j
		JOIN resources r ON r.id = j.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudDataMigrationJobs: %w", err)
	}
	defer rows.Close()
	var out []MulticloudDataMigrationJobDB
	for rows.Next() {
		var v MulticloudDataMigrationJobDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.SourceStore, &v.TargetStore, &v.DataType,
			&v.TotalSize, &v.Transferred, &v.ProgressPct,
			&v.Throughput, &v.Validation, &v.StartedAt, &v.JobStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudDataMigrationJob(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, sourceStore, targetStore, dataType, totalSize string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_data_migration_job',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudDataMigrationJob resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_data_migration_jobs
		       (resource_id, source_store, target_store, data_type, total_size)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, sourceStore, targetStore, dataType, totalSize)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudDataMigrationJob row: %w", err)
	}
	return resourceID, nil
}

// ── Multi-Cloud Migration: Application Migration ──────────────────────────────

type MulticloudAppMigrationDB struct {
	ResourceID      int64
	Name            string
	Status          string
	Region          string
	AppType         string
	SourcePlatform  string
	TargetPlatform  string
	Strategy        string
	Services        int
	Dependencies    int
	MigratedPct     string
	Rollback        string
	StartedAt       string
	MigrationStatus string
	CreatedAt       time.Time
}

func ListMulticloudAppMigrations(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudAppMigrationDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       m.app_type, m.source_platform, m.target_platform, m.strategy,
		       m.services, m.dependencies, m.migrated_pct,
		       m.rollback, m.started_at, m.migration_status,
		       r.created_at
		FROM workspace_multicloud_app_migrations m
		JOIN resources r ON r.id = m.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudAppMigrations: %w", err)
	}
	defer rows.Close()
	var out []MulticloudAppMigrationDB
	for rows.Next() {
		var v MulticloudAppMigrationDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.AppType, &v.SourcePlatform, &v.TargetPlatform, &v.Strategy,
			&v.Services, &v.Dependencies, &v.MigratedPct,
			&v.Rollback, &v.StartedAt, &v.MigrationStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudAppMigration(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, appType, sourcePlatform, targetPlatform, strategy string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_app_migration',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudAppMigration resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_app_migrations
		       (resource_id, app_type, source_platform, target_platform, strategy)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, appType, sourcePlatform, targetPlatform, strategy)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudAppMigration row: %w", err)
	}
	return resourceID, nil
}

// ── Pro: App Hosting — Static Sites ───────────────────────────────────────────

type ProStaticSiteDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Domain       string
	Framework    string
	Repository   string
	Branch       string
	BuildCmd     string
	OutputDir    string
	CDNRegions   int
	LastDeploy   string
	DeployStatus string
	SSL          string
	SiteStatus   string
	CreatedAt    time.Time
}

func ListProStaticSites(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ProStaticSiteDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.domain, s.framework, s.repository, s.branch,
		       s.build_cmd, s.output_dir, s.cdn_regions,
		       s.last_deploy, s.deploy_status, s.ssl, s.site_status,
		       r.created_at
		FROM workspace_static_sites s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListProStaticSites: %w", err)
	}
	defer rows.Close()
	var out []ProStaticSiteDB
	for rows.Next() {
		var v ProStaticSiteDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Domain, &v.Framework, &v.Repository, &v.Branch,
			&v.BuildCmd, &v.OutputDir, &v.CDNRegions,
			&v.LastDeploy, &v.DeployStatus, &v.SSL, &v.SiteStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateProStaticSite(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, domain, framework, repository, branch string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'pro_static_site',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateProStaticSite resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_static_sites (resource_id, domain, framework, repository, branch)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, domain, framework, repository, branch)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateProStaticSite row: %w", err)
	}
	return resourceID, nil
}

// ── Pro: Compute — Linux VMs ──────────────────────────────────────────────────

type LinuxVMDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Distro      string
	Version     string
	VCPUs       int
	MemoryGB    int
	DiskGB      int
	PublicIP    string
	PrivateIP   string
	SSHKey      string
	Uptime      string
	MonthlyCost string
	VMStatus    string
	CreatedAt   time.Time
}

func ListLinuxVMs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]LinuxVMDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       v.distro, v.version, v.vcpus, v.memory_gb, v.disk_gb,
		       v.public_ip, v.private_ip, v.ssh_key,
		       v.uptime, v.monthly_cost, v.vm_status,
		       r.created_at
		FROM workspace_linux_vms v
		JOIN resources r ON r.id = v.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListLinuxVMs: %w", err)
	}
	defer rows.Close()
	var out []LinuxVMDB
	for rows.Next() {
		var v LinuxVMDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Distro, &v.Version, &v.VCPUs, &v.MemoryGB, &v.DiskGB,
			&v.PublicIP, &v.PrivateIP, &v.SSHKey,
			&v.Uptime, &v.MonthlyCost, &v.VMStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateLinuxVM(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, distro, version string, vcpus, memoryGB, diskGB int, region string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'linux_vm',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateLinuxVM resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_linux_vms (resource_id, distro, version, vcpus, memory_gb, disk_gb)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, distro, version, vcpus, memoryGB, diskGB)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateLinuxVM row: %w", err)
	}
	return resourceID, nil
}

// ── Pro: Databases — Managed PostgreSQL ───────────────────────────────────────

type ManagedPostgresDB struct {
	ResourceID      int64
	Name            string
	Status          string
	Region          string
	PGVersion       string
	Plan            string
	VCPUs           int
	MemoryGB        int
	StorageGB       int
	Host            string
	Port            int
	Replicas        int
	BackupSchedule  string
	SSLEnforced     bool
	ConnectionLimit int
	DBStatus        string
	CreatedAt       time.Time
}

func ListManagedPostgres(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]ManagedPostgresDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       p.pg_version, p.plan, p.vcpus, p.memory_gb, p.storage_gb,
		       p.host, p.port, p.replicas, p.backup_schedule,
		       p.ssl_enforced, p.connection_limit, p.db_status,
		       r.created_at
		FROM workspace_managed_postgres p
		JOIN resources r ON r.id = p.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListManagedPostgres: %w", err)
	}
	defer rows.Close()
	var out []ManagedPostgresDB
	for rows.Next() {
		var v ManagedPostgresDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.PGVersion, &v.Plan, &v.VCPUs, &v.MemoryGB, &v.StorageGB,
			&v.Host, &v.Port, &v.Replicas, &v.BackupSchedule,
			&v.SSLEnforced, &v.ConnectionLimit, &v.DBStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateManagedPostgres(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, pgVersion, plan string, vcpus, memoryGB, storageGB int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'managed_postgres',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateManagedPostgres resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_managed_postgres
		       (resource_id, pg_version, plan, vcpus, memory_gb, storage_gb)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, pgVersion, plan, vcpus, memoryGB, storageGB)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateManagedPostgres row: %w", err)
	}
	return resourceID, nil
}
