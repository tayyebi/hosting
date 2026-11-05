-- 0014: ids_events, incidents, status_components, maintenance_windows, tenant_transactions

-- ── ids_events ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ids_events (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    severity    TEXT NOT NULL DEFAULT 'medium' CHECK (severity IN ('critical','high','medium','low','info')),
    event_type  TEXT NOT NULL DEFAULT '',
    source_ip   TEXT NOT NULL DEFAULT '',
    target      TEXT NOT NULL DEFAULT '',
    protocol    TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','reviewing','acknowledged','mitigated','suppressed')),
    rule_id     TEXT NOT NULL DEFAULT '',
    details     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'open';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS severity TEXT NOT NULL DEFAULT 'medium';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS event_type TEXT NOT NULL DEFAULT '';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS source_ip TEXT NOT NULL DEFAULT '';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS target TEXT NOT NULL DEFAULT '';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS protocol TEXT NOT NULL DEFAULT '';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS action TEXT NOT NULL DEFAULT '';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS rule_id TEXT NOT NULL DEFAULT '';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS details TEXT NOT NULL DEFAULT '';
ALTER TABLE ids_events ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS ids_events_tenant_idx   ON ids_events (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ids_events_severity_idx ON ids_events (tenant_id, severity) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ids_events_status_idx   ON ids_events (tenant_id, status) WHERE deleted_at IS NULL;

-- ── incidents ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS incidents (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ref_num     TEXT NOT NULL DEFAULT '',
    title       TEXT NOT NULL DEFAULT '',
    severity    TEXT NOT NULL DEFAULT 'minor' CHECK (severity IN ('critical','major','minor','low')),
    status      TEXT NOT NULL DEFAULT 'investigating' CHECK (status IN ('investigating','identified','monitoring','resolved')),
    affected    TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'investigating';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS severity TEXT NOT NULL DEFAULT 'minor';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS ref_num TEXT NOT NULL DEFAULT '';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS affected TEXT NOT NULL DEFAULT '';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS incidents_tenant_idx ON incidents (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS incidents_status_idx ON incidents (tenant_id, status) WHERE deleted_at IS NULL;

-- ── status_components ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS status_components (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL DEFAULT '',
    group_name  TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'operational' CHECK (status IN ('operational','degraded','partial-outage','major-outage','maintenance')),
    uptime_90d  NUMERIC(5,2) NOT NULL DEFAULT 100,
    order_num   INTEGER NOT NULL DEFAULT 0,
    visible     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'operational';
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS group_name TEXT NOT NULL DEFAULT '';
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS uptime_90d NUMERIC(5,2) NOT NULL DEFAULT 100;
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS order_num INTEGER NOT NULL DEFAULT 0;
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS visible BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE status_components ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS status_components_tenant_idx ON status_components (tenant_id) WHERE deleted_at IS NULL;

-- ── maintenance_windows ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ref_num         TEXT NOT NULL DEFAULT '',
    title           TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    affected        TEXT NOT NULL DEFAULT '',
    scheduled_start TIMESTAMPTZ NOT NULL DEFAULT now(),
    scheduled_end   TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '2 hours'),
    risk            TEXT NOT NULL DEFAULT 'low' CHECK (risk IN ('low','medium','high','critical')),
    status          TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled','in-progress','completed','cancelled')),
    rollback        TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'scheduled';
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS ref_num TEXT NOT NULL DEFAULT '';
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS affected TEXT NOT NULL DEFAULT '';
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS risk TEXT NOT NULL DEFAULT 'low';
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS rollback TEXT NOT NULL DEFAULT '';
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS maintenance_windows_tenant_idx ON maintenance_windows (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS maintenance_windows_status_idx ON maintenance_windows (tenant_id, status) WHERE deleted_at IS NULL;

-- ── tenant_transactions ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenant_transactions (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reference   TEXT NOT NULL DEFAULT '',
    client      TEXT NOT NULL DEFAULT '',
    client_co   TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    txn_type    TEXT NOT NULL DEFAULT 'charge' CHECK (txn_type IN ('charge','credit','refund','metered','adjustment')),
    amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency    TEXT NOT NULL DEFAULT 'USD',
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','completed','failed','cancelled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS txn_type TEXT NOT NULL DEFAULT 'charge';
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS reference TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS client TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS client_co TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD';
ALTER TABLE tenant_transactions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS tenant_transactions_tenant_idx ON tenant_transactions (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS tenant_transactions_status_idx ON tenant_transactions (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS tenant_transactions_type_idx   ON tenant_transactions (tenant_id, txn_type) WHERE deleted_at IS NULL;

-- ── RLS ───────────────────────────────────────────────────────
ALTER TABLE ids_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE ids_events FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON ids_events;
CREATE POLICY tenant_isolation ON ids_events
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE incidents ENABLE ROW LEVEL SECURITY;
ALTER TABLE incidents FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON incidents;
CREATE POLICY tenant_isolation ON incidents
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE status_components ENABLE ROW LEVEL SECURITY;
ALTER TABLE status_components FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON status_components;
CREATE POLICY tenant_isolation ON status_components
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE maintenance_windows ENABLE ROW LEVEL SECURITY;
ALTER TABLE maintenance_windows FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON maintenance_windows;
CREATE POLICY tenant_isolation ON maintenance_windows
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE tenant_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_transactions FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_transactions;
CREATE POLICY tenant_isolation ON tenant_transactions
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');
