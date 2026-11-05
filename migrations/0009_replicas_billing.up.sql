-- 0009_replicas_billing: platform replicas registry and external billing
-- integration tables used by /s/replicas/, /s/current-replica/, and
-- /s/accounting/.

-- Self-healing: an earlier (now-removed) definition of `replicas` in 0004
-- used a tenant-scoped / merchant_code schema. If a DB already has that
-- shape, drop it so the canonical platform-wide schema below can be
-- created. Detection is by the presence of the `merchant_code` column.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name   = 'replicas'
          AND column_name  = 'merchant_code'
    ) THEN
        DROP TABLE replicas CASCADE;
    END IF;
END $$;

-- ── replicas ──────────────────────────────────────────────────
-- Stores every replica instance known to this platform installation.
-- The "current" replica is identified at runtime by matching hostname.
CREATE TABLE IF NOT EXISTS replicas (
    id              BIGSERIAL PRIMARY KEY,
    replica_id      TEXT NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    host            TEXT NOT NULL UNIQUE,
    region          TEXT NOT NULL DEFAULT '',
    provider        TEXT NOT NULL DEFAULT 'Platform Internal',
    datacenter      TEXT NOT NULL DEFAULT '',
    health          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (health IN ('healthy','sync-delayed','overloaded','out-of-service','pending')),
    max_workspaces  INTEGER NOT NULL DEFAULT 50,
    max_connections INTEGER NOT NULL DEFAULT 500,
    added_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_heartbeat  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS replicas_health_idx ON replicas (health);

-- ── billing_settings ──────────────────────────────────────────
-- Single-row config for the external billing API integration.
CREATE TABLE IF NOT EXISTS billing_settings (
    id          BIGSERIAL PRIMARY KEY,
    url         TEXT NOT NULL DEFAULT '',
    timeout_s   INTEGER NOT NULL DEFAULT 10,
    retry       BOOLEAN NOT NULL DEFAULT true,
    log_all     BOOLEAN NOT NULL DEFAULT true,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── billing_api_logs ──────────────────────────────────────────
-- Log every HTTP call made to the external billing API so sysadmins
-- can debug integration issues without grepping application logs.
CREATE TABLE IF NOT EXISTS billing_api_logs (
    id            BIGSERIAL PRIMARY KEY,
    method        TEXT NOT NULL,
    path          TEXT NOT NULL,
    tenant_ref    TEXT NOT NULL DEFAULT '',
    status_code   INTEGER NOT NULL,
    latency_ms    INTEGER NOT NULL,
    request_body  TEXT,
    response_body TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS billing_api_logs_created_idx ON billing_api_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS billing_api_logs_status_idx  ON billing_api_logs (status_code, created_at DESC);

-- ── billing_webhooks ──────────────────────────────────────────
-- Webhook endpoints that receive billing events (invoice.created, etc.).
CREATE TABLE IF NOT EXISTS billing_webhooks (
    id             BIGSERIAL PRIMARY KEY,
    endpoint       TEXT NOT NULL UNIQUE,
    events         TEXT NOT NULL DEFAULT 'all',
    status         TEXT NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active','paused','disabled')),
    delivery_ok    INTEGER NOT NULL DEFAULT 0,
    delivery_total INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
