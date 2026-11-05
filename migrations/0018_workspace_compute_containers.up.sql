-- 0018: workspace_compute_containers — Kubernetes clusters, container registries,
-- container instances, and web app hosting. All extend the resources CTI base table.

-- ── Managed Kubernetes clusters ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_compute_k8s (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    cluster_id     TEXT NOT NULL DEFAULT '',
    k8s_version    TEXT NOT NULL DEFAULT 'v1.29.4',
    node_count     INTEGER NOT NULL DEFAULT 0,
    node_config    TEXT NOT NULL DEFAULT '',
    pods_running   INTEGER NOT NULL DEFAULT 0,
    cni_plugin     TEXT NOT NULL DEFAULT '',
    ingress_ctrl   TEXT NOT NULL DEFAULT '',
    autoscale      BOOLEAN NOT NULL DEFAULT false,
    ha_control     BOOLEAN NOT NULL DEFAULT false
);

ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS cluster_id TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS k8s_version TEXT NOT NULL DEFAULT 'v1.29.4';
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS node_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS node_config TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS pods_running INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS cni_plugin TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS ingress_ctrl TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS autoscale BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE workspace_compute_k8s ADD COLUMN IF NOT EXISTS ha_control BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS workspace_compute_k8s_resource_idx
    ON workspace_compute_k8s (resource_id);

-- ── Container registries / repositories ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_compute_registry (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    repo_name      TEXT NOT NULL DEFAULT '',
    tag_count      INTEGER NOT NULL DEFAULT 0,
    size_bytes     BIGINT NOT NULL DEFAULT 0,
    last_scan_at   TIMESTAMPTZ,
    vuln_critical  INTEGER NOT NULL DEFAULT 0,
    vuln_high      INTEGER NOT NULL DEFAULT 0,
    vuln_medium    INTEGER NOT NULL DEFAULT 0
);

ALTER TABLE workspace_compute_registry ADD COLUMN IF NOT EXISTS repo_name TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_registry ADD COLUMN IF NOT EXISTS tag_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_registry ADD COLUMN IF NOT EXISTS size_bytes BIGINT NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_registry ADD COLUMN IF NOT EXISTS last_scan_at TIMESTAMPTZ;
ALTER TABLE workspace_compute_registry ADD COLUMN IF NOT EXISTS vuln_critical INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_registry ADD COLUMN IF NOT EXISTS vuln_high INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_registry ADD COLUMN IF NOT EXISTS vuln_medium INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS workspace_compute_registry_resource_idx
    ON workspace_compute_registry (resource_id);

-- ── Container instances ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_compute_container_instances (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    image          TEXT NOT NULL DEFAULT '',
    cpu_millicores INTEGER NOT NULL DEFAULT 250,
    memory_mb      INTEGER NOT NULL DEFAULT 512,
    ip_address     TEXT NOT NULL DEFAULT '',
    restart_count  INTEGER NOT NULL DEFAULT 0
);

ALTER TABLE workspace_compute_container_instances ADD COLUMN IF NOT EXISTS image TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_container_instances ADD COLUMN IF NOT EXISTS cpu_millicores INTEGER NOT NULL DEFAULT 250;
ALTER TABLE workspace_compute_container_instances ADD COLUMN IF NOT EXISTS memory_mb INTEGER NOT NULL DEFAULT 512;
ALTER TABLE workspace_compute_container_instances ADD COLUMN IF NOT EXISTS ip_address TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_container_instances ADD COLUMN IF NOT EXISTS restart_count INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS workspace_compute_container_instances_resource_idx
    ON workspace_compute_container_instances (resource_id);

-- ── Web app hosting ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_hosting_web (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    domain         TEXT NOT NULL DEFAULT '',
    runtime        TEXT NOT NULL DEFAULT '',
    stack          TEXT NOT NULL DEFAULT '',
    replicas       INTEGER NOT NULL DEFAULT 1,
    memory_mb      INTEGER NOT NULL DEFAULT 256,
    cpu_pct        INTEGER NOT NULL DEFAULT 0
);

ALTER TABLE workspace_hosting_web ADD COLUMN IF NOT EXISTS domain TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_hosting_web ADD COLUMN IF NOT EXISTS runtime TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_hosting_web ADD COLUMN IF NOT EXISTS stack TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_hosting_web ADD COLUMN IF NOT EXISTS replicas INTEGER NOT NULL DEFAULT 1;
ALTER TABLE workspace_hosting_web ADD COLUMN IF NOT EXISTS memory_mb INTEGER NOT NULL DEFAULT 256;
ALTER TABLE workspace_hosting_web ADD COLUMN IF NOT EXISTS cpu_pct INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS workspace_hosting_web_resource_idx
    ON workspace_hosting_web (resource_id);
