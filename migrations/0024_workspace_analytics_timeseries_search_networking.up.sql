-- 0024: workspace resources — analytics (time-series, search) and
--        networking virtual-network (vpc, peering, nsg).
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Time-series databases ──────────────────────────────────────────────────────
-- Covers databases/analytics/timeseries
-- engine: 'influxdb' | 'timescaledb' | 'prometheus' | 'victoriametrics'
CREATE TABLE IF NOT EXISTS workspace_timeseries_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'influxdb',
    version          TEXT NOT NULL DEFAULT '3.0',
    size             TEXT NOT NULL DEFAULT 'ts.medium',
    retention_days   INTEGER NOT NULL DEFAULT 30,
    storage_gb       INTEGER NOT NULL DEFAULT 100,
    metrics_count    BIGINT NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_timeseries_dbs_resource_idx
    ON workspace_timeseries_dbs (resource_id);

-- ── Search & full-text clusters ────────────────────────────────────────────────
-- Covers databases/analytics/search
-- engine: 'elasticsearch' | 'opensearch' | 'meilisearch' | 'typesense'
CREATE TABLE IF NOT EXISTS workspace_search_clusters (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'elasticsearch',
    version          TEXT NOT NULL DEFAULT '8.13',
    size             TEXT NOT NULL DEFAULT 'search.medium',
    nodes            INTEGER NOT NULL DEFAULT 1,
    storage_gb       INTEGER NOT NULL DEFAULT 50,
    indices          INTEGER NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_search_clusters_resource_idx
    ON workspace_search_clusters (resource_id);

-- ── Virtual Private Clouds (VPCs) ─────────────────────────────────────────────
-- Covers networking/vnet/vpc
CREATE TABLE IF NOT EXISTS workspace_vpcs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    cidr             TEXT NOT NULL DEFAULT '10.0.0.0/16',
    dns              TEXT NOT NULL DEFAULT 'enabled',
    subnets          INTEGER NOT NULL DEFAULT 0,
    peerings         INTEGER NOT NULL DEFAULT 0,
    gateway          TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_vpcs_resource_idx
    ON workspace_vpcs (resource_id);

-- ── VPC Peerings ───────────────────────────────────────────────────────────────
-- Covers networking/vnet/peering
-- direction: 'bidirectional' | 'requester' | 'accepter'
CREATE TABLE IF NOT EXISTS workspace_vpc_peerings (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    local_vpc        TEXT NOT NULL DEFAULT '',
    remote_vpc       TEXT NOT NULL DEFAULT '',
    remote_owner     TEXT NOT NULL DEFAULT 'self',
    direction        TEXT NOT NULL DEFAULT 'bidirectional',
    bandwidth        TEXT NOT NULL DEFAULT '1 Gbps',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_vpc_peerings_resource_idx
    ON workspace_vpc_peerings (resource_id);

-- ── Network Security Groups ────────────────────────────────────────────────────
-- Covers networking/vnet/nsg
-- default_action: 'allow' | 'deny'
CREATE TABLE IF NOT EXISTS workspace_network_security_groups (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    vpc_name         TEXT NOT NULL DEFAULT '',
    inbound_rules    INTEGER NOT NULL DEFAULT 0,
    outbound_rules   INTEGER NOT NULL DEFAULT 0,
    associated_nics  INTEGER NOT NULL DEFAULT 0,
    default_action   TEXT NOT NULL DEFAULT 'deny',
    logging          BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_network_security_groups_resource_idx
    ON workspace_network_security_groups (resource_id);
