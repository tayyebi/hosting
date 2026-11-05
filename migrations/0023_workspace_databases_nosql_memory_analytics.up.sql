-- 0023: workspace database resources — nosql (wide-column, graph),
--        memory (cache, session), and analytics (warehouse).
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Wide-column databases (NoSQL) ─────────────────────────────────────────────
-- Covers databases/nosql/wide-column
-- engine: 'cassandra' | 'scylladb' | 'hbase'
CREATE TABLE IF NOT EXISTS workspace_widecolumn_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'cassandra',
    version          TEXT NOT NULL DEFAULT '4.1',
    nodes            INTEGER NOT NULL DEFAULT 3,
    keyspaces        INTEGER NOT NULL DEFAULT 0,
    storage_gb       INTEGER NOT NULL DEFAULT 100,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_widecolumn_dbs_resource_idx
    ON workspace_widecolumn_dbs (resource_id);

-- ── Graph databases (NoSQL) ────────────────────────────────────────────────────
-- Covers databases/nosql/graph
-- engine: 'neo4j' | 'neptune' | 'tigergraph'
CREATE TABLE IF NOT EXISTS workspace_graph_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'neo4j',
    version          TEXT NOT NULL DEFAULT '5.18',
    size             TEXT NOT NULL DEFAULT 'enterprise.medium',
    storage_gb       INTEGER NOT NULL DEFAULT 50,
    node_count       BIGINT NOT NULL DEFAULT 0,
    edge_count       BIGINT NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_graph_dbs_resource_idx
    ON workspace_graph_dbs (resource_id);

-- ── In-memory cache ────────────────────────────────────────────────────────────
-- Covers databases/memory/cache
-- engine: 'redis' | 'memcached' | 'dragonflydb'
CREATE TABLE IF NOT EXISTS workspace_memory_caches (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'redis',
    version          TEXT NOT NULL DEFAULT '7.2',
    size             TEXT NOT NULL DEFAULT 'cache.m7g.medium',
    nodes            INTEGER NOT NULL DEFAULT 1,
    max_memory_gb    INTEGER NOT NULL DEFAULT 8,
    eviction_policy  TEXT NOT NULL DEFAULT 'allkeys-lru',
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_memory_caches_resource_idx
    ON workspace_memory_caches (resource_id);

-- ── In-memory session stores ──────────────────────────────────────────────────
-- Covers databases/memory/session
-- engine: 'redis' | 'memcached' | 'valkey'
CREATE TABLE IF NOT EXISTS workspace_memory_sessions (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'redis',
    version          TEXT NOT NULL DEFAULT '7.2',
    size             TEXT NOT NULL DEFAULT 'cache.m7g.medium',
    nodes            INTEGER NOT NULL DEFAULT 1,
    max_memory_gb    INTEGER NOT NULL DEFAULT 4,
    ttl_seconds      INTEGER NOT NULL DEFAULT 1800,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_memory_sessions_resource_idx
    ON workspace_memory_sessions (resource_id);

-- ── Data warehouses ────────────────────────────────────────────────────────────
-- Covers databases/analytics/warehouse
-- engine: 'clickhouse' | 'snowflake' | 'redshift' | 'bigquery'
CREATE TABLE IF NOT EXISTS workspace_warehouse_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'clickhouse',
    version          TEXT NOT NULL DEFAULT '24.3',
    size             TEXT NOT NULL DEFAULT 'dw.xlarge',
    storage_tb       INTEGER NOT NULL DEFAULT 1,
    tables           INTEGER NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_warehouse_dbs_resource_idx
    ON workspace_warehouse_dbs (resource_id);
