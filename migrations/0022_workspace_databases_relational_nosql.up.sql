-- 0022: workspace database resources — relational (mysql/pg, sql server, cloud-native)
--        and nosql (document, key-value).
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── MySQL / PostgreSQL ─────────────────────────────────────────────────────
-- Covers databases/relational/mysql-pg
CREATE TABLE IF NOT EXISTS workspace_relational_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'postgresql',   -- 'postgresql' | 'mysql'
    version          TEXT NOT NULL DEFAULT '16.2',
    size             TEXT NOT NULL DEFAULT 'db.m6g.large',
    storage_gb       INTEGER NOT NULL DEFAULT 100,
    replicas         INTEGER NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_relational_dbs_resource_idx
    ON workspace_relational_dbs (resource_id);

-- ── SQL Server ─────────────────────────────────────────────────────────────
-- Covers databases/relational/sql-server
CREATE TABLE IF NOT EXISTS workspace_sqlserver_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    edition          TEXT NOT NULL DEFAULT 'standard',     -- 'developer' | 'standard' | 'enterprise'
    version          TEXT NOT NULL DEFAULT '2022',
    size             TEXT NOT NULL DEFAULT 'db.m6i.large',
    storage_gb       INTEGER NOT NULL DEFAULT 100,
    replicas         INTEGER NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_sqlserver_dbs_resource_idx
    ON workspace_sqlserver_dbs (resource_id);

-- ── Cloud-native databases ─────────────────────────────────────────────────
-- Covers databases/relational/cloud-native
-- engine: 'aurora-postgresql' | 'aurora-mysql-serverless' | 'cloud-spanner' | 'cockroachdb' | 'planetscale'
CREATE TABLE IF NOT EXISTS workspace_cloudnative_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'aurora-postgresql',
    version          TEXT NOT NULL DEFAULT 'latest',
    size             TEXT NOT NULL DEFAULT 'db.r6g.large',
    storage_used_gb  INTEGER NOT NULL DEFAULT 0,
    replicas         INTEGER NOT NULL DEFAULT 1,
    auto_scaling     BOOLEAN NOT NULL DEFAULT true,
    serverless_units TEXT NOT NULL DEFAULT '—',
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_cloudnative_dbs_resource_idx
    ON workspace_cloudnative_dbs (resource_id);

-- ── Document databases (NoSQL) ─────────────────────────────────────────────
-- Covers databases/nosql/document
-- engine: 'mongodb' | 'couchdb'
CREATE TABLE IF NOT EXISTS workspace_document_dbs (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'mongodb',
    version          TEXT NOT NULL DEFAULT '7.0',
    size             TEXT NOT NULL DEFAULT 'M10',
    storage_gb       INTEGER NOT NULL DEFAULT 10,
    collections      INTEGER NOT NULL DEFAULT 0,
    document_count   BIGINT NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_document_dbs_resource_idx
    ON workspace_document_dbs (resource_id);

-- ── Key-value stores (NoSQL) ───────────────────────────────────────────────
-- Covers databases/nosql/keyvalue
-- engine: 'redis' | 'etcd' | 'consul'
CREATE TABLE IF NOT EXISTS workspace_keyvalue_stores (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT NOT NULL DEFAULT 'redis',
    version          TEXT NOT NULL DEFAULT '7.2',
    size             TEXT NOT NULL DEFAULT 'cache.m7g.medium',
    max_memory_gb    INTEGER NOT NULL DEFAULT 4,
    key_count        BIGINT NOT NULL DEFAULT 0,
    endpoint         TEXT NOT NULL DEFAULT '',
    encrypted        BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS workspace_keyvalue_stores_resource_idx
    ON workspace_keyvalue_stores (resource_id);
