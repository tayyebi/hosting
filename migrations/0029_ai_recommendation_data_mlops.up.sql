-- 0029: workspace resources — AI Recommendation API + Data (ETL, Streaming, Catalog) + MLOps Experiments
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── AI Recommendation APIs ────────────────────────────────────────────────────
-- Covers ai/apis/recommendation
-- algorithm: 'collaborative' | 'content-based' | 'hybrid' | 'contextual'
CREATE TABLE IF NOT EXISTS workspace_ai_recommendation_apis (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model        TEXT NOT NULL DEFAULT '',
    algorithm    TEXT NOT NULL DEFAULT 'hybrid',
    version      TEXT NOT NULL DEFAULT 'v1.0.0',
    requests     TEXT NOT NULL DEFAULT '0/mo',
    latency      TEXT NOT NULL DEFAULT '—',
    catalog      TEXT NOT NULL DEFAULT '0 items',
    endpoint     TEXT NOT NULL DEFAULT '',
    api_key      TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_ai_recommendation_apis_resource_idx
    ON workspace_ai_recommendation_apis (resource_id);

-- ── AI ETL / ELT Pipelines ────────────────────────────────────────────────────
-- Covers ai/data/etl
-- pipeline_type: 'etl' | 'elt' | 'reverse-etl'
CREATE TABLE IF NOT EXISTS workspace_ai_etl_pipelines (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    pipeline_type TEXT NOT NULL DEFAULT 'etl',
    source        TEXT NOT NULL DEFAULT '',
    destination   TEXT NOT NULL DEFAULT '',
    schedule      TEXT NOT NULL DEFAULT 'daily',
    last_run      TEXT NOT NULL DEFAULT '—',
    duration      TEXT NOT NULL DEFAULT '—',
    records       TEXT NOT NULL DEFAULT '0 rows'
);

CREATE INDEX IF NOT EXISTS workspace_ai_etl_pipelines_resource_idx
    ON workspace_ai_etl_pipelines (resource_id);

-- ── AI Data Streams ───────────────────────────────────────────────────────────
-- Covers ai/data/streaming
-- engine: 'kafka' | 'kinesis' | 'pulsar'
CREATE TABLE IF NOT EXISTS workspace_ai_data_streams (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine        TEXT NOT NULL DEFAULT 'kafka',
    source        TEXT NOT NULL DEFAULT '',
    destination   TEXT NOT NULL DEFAULT '',
    throughput    TEXT NOT NULL DEFAULT '0 msg/s',
    partitions    INTEGER NOT NULL DEFAULT 1,
    retention     TEXT NOT NULL DEFAULT '24 hours',
    lag           TEXT NOT NULL DEFAULT '—'
);

CREATE INDEX IF NOT EXISTS workspace_ai_data_streams_resource_idx
    ON workspace_ai_data_streams (resource_id);

-- ── AI Data Catalog Assets ────────────────────────────────────────────────────
-- Covers ai/data/catalog
-- asset_type: 'table' | 'view' | 'feature_group' | 'topic'
-- pii: 'pii_present' | 'pseudonymised' | 'no_pii' | 'unknown'
CREATE TABLE IF NOT EXISTS workspace_ai_data_catalog_assets (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    source         TEXT NOT NULL DEFAULT '',
    owner          TEXT NOT NULL DEFAULT '',
    asset_type     TEXT NOT NULL DEFAULT 'table',
    pii            TEXT NOT NULL DEFAULT 'unknown',
    tags           TEXT NOT NULL DEFAULT '',
    asset_updated  TEXT NOT NULL DEFAULT '—',
    columns        INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_ai_data_catalog_assets_resource_idx
    ON workspace_ai_data_catalog_assets (resource_id);

-- ── AI ML Experiments ─────────────────────────────────────────────────────────
-- Covers ai/mlops/experiments
-- framework: 'PyTorch' | 'TensorFlow' | 'XGBoost' | 'LightGBM' | 'sklearn'
CREATE TABLE IF NOT EXISTS workspace_ai_ml_experiments (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    best_run      TEXT NOT NULL DEFAULT '—',
    metric        TEXT NOT NULL DEFAULT '—',
    runs          INTEGER NOT NULL DEFAULT 0,
    framework     TEXT NOT NULL DEFAULT 'PyTorch',
    owner         TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_ai_ml_experiments_resource_idx
    ON workspace_ai_ml_experiments (resource_id);
