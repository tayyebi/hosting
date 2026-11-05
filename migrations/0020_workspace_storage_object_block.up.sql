-- 0020: workspace_storage_object_block — object buckets (standard, archive, datalake)
-- and block volumes (SSD gp3/io2, HDD st1/sc1).

-- ── Object buckets ─────────────────────────────────────────────────────────────
-- Covers storage/object/standard, storage/object/archive, storage/object/datalake.
-- storage_class discriminates: 'standard' | 'archive' | 'datalake'
CREATE TABLE IF NOT EXISTS workspace_object_buckets (
    resource_id         BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    storage_class       TEXT NOT NULL DEFAULT 'standard',
    size_gb             NUMERIC(14,2) NOT NULL DEFAULT 0,
    object_count        BIGINT NOT NULL DEFAULT 0,
    -- standard
    versioning_enabled  BOOLEAN NOT NULL DEFAULT false,
    access_level        TEXT NOT NULL DEFAULT 'private',
    -- archive
    archive_tier        TEXT NOT NULL DEFAULT '',
    retention_policy    TEXT NOT NULL DEFAULT '',
    last_accessed_at    TIMESTAMPTZ,
    -- datalake
    data_format         TEXT NOT NULL DEFAULT '',
    partition_key       TEXT NOT NULL DEFAULT '',
    record_count        BIGINT NOT NULL DEFAULT 0,
    last_write_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS workspace_object_buckets_resource_idx
    ON workspace_object_buckets (resource_id);

CREATE INDEX IF NOT EXISTS workspace_object_buckets_class_idx
    ON workspace_object_buckets (storage_class);

-- ── Block volumes ──────────────────────────────────────────────────────────────
-- Covers storage/block/ssd (gp3, io2) and storage/block/hdd (st1, sc1).
-- volume_category discriminates: 'ssd' | 'hdd'
CREATE TABLE IF NOT EXISTS workspace_block_volumes (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    volume_category   TEXT NOT NULL DEFAULT 'ssd',
    volume_type       TEXT NOT NULL DEFAULT '',
    size_gb           INTEGER NOT NULL DEFAULT 0,
    iops              INTEGER NOT NULL DEFAULT 0,
    throughput_mbps   INTEGER NOT NULL DEFAULT 0,
    utilization_pct   INTEGER NOT NULL DEFAULT 0,
    attached_to       TEXT NOT NULL DEFAULT '',
    encrypted         BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS workspace_block_volumes_resource_idx
    ON workspace_block_volumes (resource_id);

CREATE INDEX IF NOT EXISTS workspace_block_volumes_category_idx
    ON workspace_block_volumes (volume_category);
