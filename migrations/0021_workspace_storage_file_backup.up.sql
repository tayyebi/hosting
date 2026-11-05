-- 0021: workspace_storage_file_backup — file shares (NFS/SMB, distributed) and
-- backup resources (snapshots, DR plans, long-term archives).

-- ── File shares (NFS / SMB) ────────────────────────────────────────────────
-- Covers storage/file/nfs.
CREATE TABLE IF NOT EXISTS workspace_file_shares (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    protocol       TEXT NOT NULL DEFAULT 'NFSv4',   -- 'NFSv4' | 'SMB3' | 'SMB2'
    size_gb        INTEGER NOT NULL DEFAULT 0,
    used_gb        INTEGER NOT NULL DEFAULT 0,
    mount_point    TEXT NOT NULL DEFAULT '',
    access_cidr    TEXT NOT NULL DEFAULT '',
    active_mounts  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_file_shares_resource_idx
    ON workspace_file_shares (resource_id);

-- ── Distributed file systems ───────────────────────────────────────────────
-- Covers storage/file/distributed.
-- fs_type: 'Lustre' | 'CephFS' | 'GlusterFS' | 'HDFS' | 'MinIO'
CREATE TABLE IF NOT EXISTS workspace_distributed_filesystems (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    fs_type        TEXT NOT NULL DEFAULT 'CephFS',
    capacity_tb    NUMERIC(10,2) NOT NULL DEFAULT 0,
    used_tb        NUMERIC(10,2) NOT NULL DEFAULT 0,
    node_count     INTEGER NOT NULL DEFAULT 0,
    replication    INTEGER NOT NULL DEFAULT 2,
    mount_point    TEXT NOT NULL DEFAULT '',
    active_clients INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_distributed_fs_resource_idx
    ON workspace_distributed_filesystems (resource_id);

-- ── Backup snapshots ───────────────────────────────────────────────────────
-- Covers storage/backup/snapshots.
-- snapshot_type: 'auto' | 'manual'
-- source_kind: 'block-ssd' | 'block-hdd' | 'vm' | 'nfs' | 'database'
CREATE TABLE IF NOT EXISTS workspace_snapshots (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    snapshot_type  TEXT NOT NULL DEFAULT 'auto',
    source_kind    TEXT NOT NULL DEFAULT 'block-ssd',
    source_name    TEXT NOT NULL DEFAULT '',
    size_gb        NUMERIC(14,2) NOT NULL DEFAULT 0,
    retention_days INTEGER NOT NULL DEFAULT 30,
    encrypted      BOOLEAN NOT NULL DEFAULT false,
    verified_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS workspace_snapshots_resource_idx
    ON workspace_snapshots (resource_id);

-- ── Disaster recovery plans ────────────────────────────────────────────────
-- Covers storage/backup/dr.
-- replication_type: 'continuous' | 'scheduled' | 'manual'
CREATE TABLE IF NOT EXISTS workspace_dr_plans (
    resource_id        BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    source_region      TEXT NOT NULL DEFAULT '',
    target_region      TEXT NOT NULL DEFAULT '',
    rpo_minutes        INTEGER NOT NULL DEFAULT 60,
    rto_minutes        INTEGER NOT NULL DEFAULT 240,
    replication_type   TEXT NOT NULL DEFAULT 'scheduled',
    protected_count    INTEGER NOT NULL DEFAULT 0,
    replication_lag_s  INTEGER NOT NULL DEFAULT 0,
    encrypted          BOOLEAN NOT NULL DEFAULT false,
    last_tested_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS workspace_dr_plans_resource_idx
    ON workspace_dr_plans (resource_id);

-- ── Long-term archives ─────────────────────────────────────────────────────
-- Covers storage/backup/archive.
-- archive_class: 'Glacier Deep' | 'Glacier' | 'Standard-IA' | 'Cold'
-- archive_type:  'auto' | 'manual'
CREATE TABLE IF NOT EXISTS workspace_archive_jobs (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    archive_class  TEXT NOT NULL DEFAULT 'Glacier Deep',
    archive_type   TEXT NOT NULL DEFAULT 'auto',
    source_name    TEXT NOT NULL DEFAULT '',
    size_gb        NUMERIC(14,2) NOT NULL DEFAULT 0,
    retention_days INTEGER NOT NULL DEFAULT 2555,  -- 7 years
    encrypted      BOOLEAN NOT NULL DEFAULT false,
    verified_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS workspace_archive_jobs_resource_idx
    ON workspace_archive_jobs (resource_id);
