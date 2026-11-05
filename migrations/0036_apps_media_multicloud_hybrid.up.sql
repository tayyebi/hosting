-- 0036: workspace resources — Application Services: Media (Encoding, Livestream, Storage)
--        + Multi-Cloud & Hybrid: On-Premises Extension, Unified Control Plane
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Media: Video Encoding ─────────────────────────────────────────────────────
-- Covers apps/media/encoding
CREATE TABLE IF NOT EXISTS workspace_apps_media_encoding (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    input_format   TEXT    NOT NULL DEFAULT '',
    profile        TEXT    NOT NULL DEFAULT '',
    duration       TEXT    NOT NULL DEFAULT '—',
    progress       INTEGER NOT NULL DEFAULT 0,
    output_size    TEXT    NOT NULL DEFAULT '—',
    eta            TEXT    NOT NULL DEFAULT '—',
    encode_status  TEXT    NOT NULL DEFAULT 'queued'
);
CREATE INDEX IF NOT EXISTS workspace_apps_media_encoding_resource_idx ON workspace_apps_media_encoding (resource_id);

-- ── Media: Live Streaming ─────────────────────────────────────────────────────
-- Covers apps/media/livestream
CREATE TABLE IF NOT EXISTS workspace_apps_media_livestream (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    source_protocol  TEXT    NOT NULL DEFAULT 'RTMP',
    output_protocol  TEXT    NOT NULL DEFAULT 'HLS',
    profile          TEXT    NOT NULL DEFAULT '',
    viewers          INTEGER NOT NULL DEFAULT 0,
    bitrate          TEXT    NOT NULL DEFAULT '—',
    latency          TEXT    NOT NULL DEFAULT '—',
    stream_status    TEXT    NOT NULL DEFAULT 'draft'
);
CREATE INDEX IF NOT EXISTS workspace_apps_media_livestream_resource_idx ON workspace_apps_media_livestream (resource_id);

-- ── Media: Media Storage ──────────────────────────────────────────────────────
-- Covers apps/media/storage
CREATE TABLE IF NOT EXISTS workspace_apps_media_storage (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    asset_type    TEXT    NOT NULL DEFAULT '',
    size          TEXT    NOT NULL DEFAULT '—',
    codec         TEXT    NOT NULL DEFAULT '',
    plays_7d      INTEGER NOT NULL DEFAULT 0,
    bandwidth     TEXT    NOT NULL DEFAULT '—',
    cdn_hit       TEXT    NOT NULL DEFAULT '—',
    asset_status  TEXT    NOT NULL DEFAULT 'pending'
);
CREATE INDEX IF NOT EXISTS workspace_apps_media_storage_resource_idx ON workspace_apps_media_storage (resource_id);

-- ── Multi-Cloud: On-Premises Extension ───────────────────────────────────────
-- Covers multicloud/hybrid/onprem
CREATE TABLE IF NOT EXISTS workspace_multicloud_onprem_sites (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    location     TEXT    NOT NULL DEFAULT '',
    hosts        INTEGER NOT NULL DEFAULT 0,
    vcpus        INTEGER NOT NULL DEFAULT 0,
    workloads    INTEGER NOT NULL DEFAULT 0,
    link_type    TEXT    NOT NULL DEFAULT '',
    bandwidth    TEXT    NOT NULL DEFAULT '—',
    latency      TEXT    NOT NULL DEFAULT '—',
    site_status  TEXT    NOT NULL DEFAULT 'connected'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_onprem_sites_resource_idx ON workspace_multicloud_onprem_sites (resource_id);

-- ── Multi-Cloud: Unified Control Plane ───────────────────────────────────────
-- Covers multicloud/hybrid/control
CREATE TABLE IF NOT EXISTS workspace_multicloud_control_clusters (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    provider         TEXT    NOT NULL DEFAULT '',
    region           TEXT    NOT NULL DEFAULT '',
    nodes            INTEGER NOT NULL DEFAULT 0,
    pods             INTEGER NOT NULL DEFAULT 0,
    policies         TEXT    NOT NULL DEFAULT '—',
    config_drift     TEXT    NOT NULL DEFAULT '0 drift',
    cluster_status   TEXT    NOT NULL DEFAULT 'healthy'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_control_clusters_resource_idx ON workspace_multicloud_control_clusters (resource_id);
