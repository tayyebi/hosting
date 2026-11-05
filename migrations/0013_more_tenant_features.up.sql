-- 0013_more_tenant_features: abuse_reports, refund_complaints, one_time_access, tenant_resources, monitoring_checks

-- ── abuse_reports ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS abuse_reports (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type        TEXT NOT NULL DEFAULT 'other' CHECK (type IN ('spam','phishing','malware','tos-violation','dmca','other')),
    source      TEXT NOT NULL DEFAULT '',
    target      TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new','investigating','resolved','dismissed')),
    summary     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE abuse_reports ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'new';
ALTER TABLE abuse_reports ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'other';
ALTER TABLE abuse_reports ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT '';
ALTER TABLE abuse_reports ADD COLUMN IF NOT EXISTS target TEXT NOT NULL DEFAULT '';
ALTER TABLE abuse_reports ADD COLUMN IF NOT EXISTS summary TEXT NOT NULL DEFAULT '';
ALTER TABLE abuse_reports ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS abuse_reports_tenant_idx ON abuse_reports (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS abuse_reports_status_idx ON abuse_reports (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS abuse_reports_type_idx   ON abuse_reports (tenant_id, type) WHERE deleted_at IS NULL;

-- ── refund_complaints ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS refund_complaints (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    client      TEXT NOT NULL DEFAULT '',
    amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency    TEXT NOT NULL DEFAULT 'USD',
    reason      TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','denied','escalated')),
    invoice_id  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE refund_complaints ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE refund_complaints ADD COLUMN IF NOT EXISTS client TEXT NOT NULL DEFAULT '';
ALTER TABLE refund_complaints ADD COLUMN IF NOT EXISTS reason TEXT NOT NULL DEFAULT '';
ALTER TABLE refund_complaints ADD COLUMN IF NOT EXISTS invoice_id TEXT NOT NULL DEFAULT '';
ALTER TABLE refund_complaints ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS refund_complaints_tenant_idx ON refund_complaints (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS refund_complaints_status_idx ON refund_complaints (tenant_id, status) WHERE deleted_at IS NULL;

-- ── one_time_access ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS one_time_access (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    label       TEXT NOT NULL DEFAULT '',
    token       TEXT NOT NULL,
    target      TEXT NOT NULL DEFAULT '',
    granted_to  TEXT NOT NULL DEFAULT '',
    granted_by  TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','used','expired','revoked')),
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '12 hours'),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (tenant_id, token)
);
ALTER TABLE one_time_access ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE one_time_access ADD COLUMN IF NOT EXISTS label TEXT NOT NULL DEFAULT '';
ALTER TABLE one_time_access ADD COLUMN IF NOT EXISTS target TEXT NOT NULL DEFAULT '';
ALTER TABLE one_time_access ADD COLUMN IF NOT EXISTS granted_to TEXT NOT NULL DEFAULT '';
ALTER TABLE one_time_access ADD COLUMN IF NOT EXISTS granted_by TEXT NOT NULL DEFAULT '';
ALTER TABLE one_time_access ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS one_time_access_tenant_idx ON one_time_access (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS one_time_access_status_idx ON one_time_access (tenant_id, status) WHERE deleted_at IS NULL;

-- ── tenant_resources ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenant_resources (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    resource_type TEXT NOT NULL DEFAULT '',
    provider      TEXT NOT NULL DEFAULT '',
    region        TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'online' CHECK (status IN ('online','degraded','offline','maintenance','provisioning')),
    ip_address    TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);
ALTER TABLE tenant_resources ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'online';
ALTER TABLE tenant_resources ADD COLUMN IF NOT EXISTS resource_type TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_resources ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_resources ADD COLUMN IF NOT EXISTS region TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_resources ADD COLUMN IF NOT EXISTS ip_address TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_resources ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS tenant_resources_tenant_idx ON tenant_resources (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS tenant_resources_status_idx ON tenant_resources (tenant_id, status) WHERE deleted_at IS NULL;

-- ── monitoring_checks ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS monitoring_checks (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    check_type   TEXT NOT NULL DEFAULT 'http' CHECK (check_type IN ('http','ping','tcp','dns','ssl')),
    target       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'up' CHECK (status IN ('up','degraded','down','unknown')),
    uptime_30d   NUMERIC(5,2) NOT NULL DEFAULT 100,
    latency_ms   INTEGER NOT NULL DEFAULT 0,
    last_checked TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);
ALTER TABLE monitoring_checks ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'up';
ALTER TABLE monitoring_checks ADD COLUMN IF NOT EXISTS check_type TEXT NOT NULL DEFAULT 'http';
ALTER TABLE monitoring_checks ADD COLUMN IF NOT EXISTS target TEXT NOT NULL DEFAULT '';
ALTER TABLE monitoring_checks ADD COLUMN IF NOT EXISTS uptime_30d NUMERIC(5,2) NOT NULL DEFAULT 100;
ALTER TABLE monitoring_checks ADD COLUMN IF NOT EXISTS latency_ms INTEGER NOT NULL DEFAULT 0;
ALTER TABLE monitoring_checks ADD COLUMN IF NOT EXISTS last_checked TIMESTAMPTZ;
ALTER TABLE monitoring_checks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS monitoring_checks_tenant_idx ON monitoring_checks (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS monitoring_checks_status_idx ON monitoring_checks (tenant_id, status) WHERE deleted_at IS NULL;

-- ── RLS for all new tables ─────────────────────────────────────
ALTER TABLE abuse_reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE abuse_reports FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON abuse_reports;
CREATE POLICY tenant_isolation ON abuse_reports
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE refund_complaints ENABLE ROW LEVEL SECURITY;
ALTER TABLE refund_complaints FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON refund_complaints;
CREATE POLICY tenant_isolation ON refund_complaints
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE one_time_access ENABLE ROW LEVEL SECURITY;
ALTER TABLE one_time_access FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON one_time_access;
CREATE POLICY tenant_isolation ON one_time_access
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE tenant_resources ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_resources FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_resources;
CREATE POLICY tenant_isolation ON tenant_resources
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE monitoring_checks ENABLE ROW LEVEL SECURITY;
ALTER TABLE monitoring_checks FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON monitoring_checks;
CREATE POLICY tenant_isolation ON monitoring_checks
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');
