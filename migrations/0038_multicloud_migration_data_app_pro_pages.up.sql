-- 0038: workspace resources — Multi-Cloud Migration: Data + App
--        + Pro feature pages: Static Sites, Linux VMs, Managed PostgreSQL
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Multi-Cloud Migration: Data Migration ─────────────────────────────────────
-- Covers multicloud/migration/data
CREATE TABLE IF NOT EXISTS workspace_multicloud_data_migration_jobs (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    source_store  TEXT    NOT NULL DEFAULT '',
    target_store  TEXT    NOT NULL DEFAULT '',
    data_type     TEXT    NOT NULL DEFAULT '',
    total_size    TEXT    NOT NULL DEFAULT '—',
    transferred   TEXT    NOT NULL DEFAULT '0 GB',
    progress_pct  TEXT    NOT NULL DEFAULT '0%',
    throughput    TEXT    NOT NULL DEFAULT '—',
    validation    TEXT    NOT NULL DEFAULT 'Not started',
    started_at    TEXT    NOT NULL DEFAULT '—',
    job_status    TEXT    NOT NULL DEFAULT 'pending'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_data_migration_jobs_resource_idx
    ON workspace_multicloud_data_migration_jobs (resource_id);

-- ── Multi-Cloud Migration: Application Migration ──────────────────────────────
-- Covers multicloud/migration/app
CREATE TABLE IF NOT EXISTS workspace_multicloud_app_migrations (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    app_type          TEXT    NOT NULL DEFAULT '',
    source_platform   TEXT    NOT NULL DEFAULT '',
    target_platform   TEXT    NOT NULL DEFAULT '',
    strategy          TEXT    NOT NULL DEFAULT '',
    services          INTEGER NOT NULL DEFAULT 0,
    dependencies      INTEGER NOT NULL DEFAULT 0,
    migrated_pct      TEXT    NOT NULL DEFAULT '0%',
    rollback          TEXT    NOT NULL DEFAULT 'N/A',
    started_at        TEXT    NOT NULL DEFAULT '—',
    migration_status  TEXT    NOT NULL DEFAULT 'pending'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_app_migrations_resource_idx
    ON workspace_multicloud_app_migrations (resource_id);

-- ── Pro: App Hosting — Static Sites ───────────────────────────────────────────
-- Covers apps/app-hosting/static-sites
CREATE TABLE IF NOT EXISTS workspace_static_sites (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    domain        TEXT    NOT NULL DEFAULT '',
    framework     TEXT    NOT NULL DEFAULT '',
    repository    TEXT    NOT NULL DEFAULT '',
    branch        TEXT    NOT NULL DEFAULT 'main',
    build_cmd     TEXT    NOT NULL DEFAULT '',
    output_dir    TEXT    NOT NULL DEFAULT '',
    cdn_regions   INTEGER NOT NULL DEFAULT 0,
    last_deploy   TEXT    NOT NULL DEFAULT '—',
    deploy_status TEXT    NOT NULL DEFAULT 'Not deployed',
    ssl           TEXT    NOT NULL DEFAULT 'Pending',
    site_status   TEXT    NOT NULL DEFAULT 'draft'
);
CREATE INDEX IF NOT EXISTS workspace_static_sites_resource_idx
    ON workspace_static_sites (resource_id);

-- ── Pro: Compute — Linux VMs ──────────────────────────────────────────────────
-- Covers compute/virtual-machines/linux-vms
CREATE TABLE IF NOT EXISTS workspace_linux_vms (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    distro       TEXT    NOT NULL DEFAULT '',
    version      TEXT    NOT NULL DEFAULT '',
    vcpus        INTEGER NOT NULL DEFAULT 0,
    memory_gb    INTEGER NOT NULL DEFAULT 0,
    disk_gb      INTEGER NOT NULL DEFAULT 0,
    public_ip    TEXT    NOT NULL DEFAULT '—',
    private_ip   TEXT    NOT NULL DEFAULT '',
    ssh_key      TEXT    NOT NULL DEFAULT '',
    uptime       TEXT    NOT NULL DEFAULT '—',
    monthly_cost TEXT    NOT NULL DEFAULT '—',
    vm_status    TEXT    NOT NULL DEFAULT 'stopped'
);
CREATE INDEX IF NOT EXISTS workspace_linux_vms_resource_idx
    ON workspace_linux_vms (resource_id);

-- ── Pro: Databases — Managed PostgreSQL ───────────────────────────────────────
-- Covers databases/relational/postgresql
CREATE TABLE IF NOT EXISTS workspace_managed_postgres (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    pg_version       TEXT    NOT NULL DEFAULT '',
    plan             TEXT    NOT NULL DEFAULT '',
    vcpus            INTEGER NOT NULL DEFAULT 0,
    memory_gb        INTEGER NOT NULL DEFAULT 0,
    storage_gb       INTEGER NOT NULL DEFAULT 0,
    host             TEXT    NOT NULL DEFAULT '',
    port             INTEGER NOT NULL DEFAULT 5432,
    replicas         INTEGER NOT NULL DEFAULT 0,
    backup_schedule  TEXT    NOT NULL DEFAULT 'Daily',
    ssl_enforced     BOOLEAN NOT NULL DEFAULT true,
    connection_limit INTEGER NOT NULL DEFAULT 100,
    db_status        TEXT    NOT NULL DEFAULT 'running'
);
CREATE INDEX IF NOT EXISTS workspace_managed_postgres_resource_idx
    ON workspace_managed_postgres (resource_id);
