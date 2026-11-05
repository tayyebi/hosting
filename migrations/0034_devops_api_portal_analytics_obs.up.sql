-- 0034: workspace resources — DevOps API Portal + API Analytics + Observability (Logging, Metrics, Tracing)
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── API Developer Portal ──────────────────────────────────────────────────────
-- Covers devops/api/portal
CREATE TABLE IF NOT EXISTS workspace_devops_api_portal (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    owner         TEXT NOT NULL DEFAULT '',
    plan          TEXT NOT NULL DEFAULT 'Public',
    api_keys      TEXT NOT NULL DEFAULT '1 key',
    calls_7d      TEXT NOT NULL DEFAULT '—',
    last_used     TEXT NOT NULL DEFAULT '—',
    quota_usage   TEXT NOT NULL DEFAULT '—',
    portal_status TEXT NOT NULL DEFAULT 'Active'
);
CREATE INDEX IF NOT EXISTS workspace_devops_api_portal_resource_idx ON workspace_devops_api_portal (resource_id);

-- ── API Analytics ────────────────────────────────────────────────────────────
-- Covers devops/api/analytics
CREATE TABLE IF NOT EXISTS workspace_devops_api_analytics (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    api_name     TEXT NOT NULL DEFAULT '',
    method       TEXT NOT NULL DEFAULT 'GET',
    calls_24h    TEXT NOT NULL DEFAULT '—',
    latency_p50  TEXT NOT NULL DEFAULT '—',
    latency_p95  TEXT NOT NULL DEFAULT '—',
    latency_p99  TEXT NOT NULL DEFAULT '—',
    error_rate   TEXT NOT NULL DEFAULT '—',
    ep_status    TEXT NOT NULL DEFAULT 'Healthy'
);
CREATE INDEX IF NOT EXISTS workspace_devops_api_analytics_resource_idx ON workspace_devops_api_analytics (resource_id);

-- ── Observability: Logging ────────────────────────────────────────────────────
-- Covers devops/obs/logging
CREATE TABLE IF NOT EXISTS workspace_devops_obs_logging (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    source       TEXT NOT NULL DEFAULT '',
    format       TEXT NOT NULL DEFAULT 'JSON',
    volume_24h   TEXT NOT NULL DEFAULT '—',
    lines_24h    TEXT NOT NULL DEFAULT '—',
    retention    TEXT NOT NULL DEFAULT '30d hot',
    lag          TEXT NOT NULL DEFAULT '< 2s',
    log_status   TEXT NOT NULL DEFAULT 'Ingesting'
);
CREATE INDEX IF NOT EXISTS workspace_devops_obs_logging_resource_idx ON workspace_devops_obs_logging (resource_id);

-- ── Observability: Metrics & Dashboards ──────────────────────────────────────
-- Covers devops/obs/metrics
CREATE TABLE IF NOT EXISTS workspace_devops_obs_metrics (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    owner         TEXT NOT NULL DEFAULT '',
    panels        INTEGER NOT NULL DEFAULT 0,
    data_source   TEXT NOT NULL DEFAULT '',
    refresh_rate  TEXT NOT NULL DEFAULT '30s',
    last_viewed   TEXT NOT NULL DEFAULT '—',
    alerts_linked TEXT NOT NULL DEFAULT '0 alerts',
    dash_status   TEXT NOT NULL DEFAULT 'Live'
);
CREATE INDEX IF NOT EXISTS workspace_devops_obs_metrics_resource_idx ON workspace_devops_obs_metrics (resource_id);

-- ── Observability: Distributed Tracing ───────────────────────────────────────
-- Covers devops/obs/tracing
CREATE TABLE IF NOT EXISTS workspace_devops_obs_tracing (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    spans_24h    TEXT NOT NULL DEFAULT '—',
    latency_p50  TEXT NOT NULL DEFAULT '—',
    latency_p95  TEXT NOT NULL DEFAULT '—',
    latency_p99  TEXT NOT NULL DEFAULT '—',
    error_rate   TEXT NOT NULL DEFAULT '—',
    sampling     TEXT NOT NULL DEFAULT '5% head',
    svc_status   TEXT NOT NULL DEFAULT 'Healthy'
);
CREATE INDEX IF NOT EXISTS workspace_devops_obs_tracing_resource_idx ON workspace_devops_obs_tracing (resource_id);
