-- 0012_tenant_features: clients, employee profile extras, tickets

-- ── tenant_clients ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenant_clients (
    id         BIGSERIAL PRIMARY KEY,
    tenant_id  BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    email      CITEXT NOT NULL,
    company    TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, email)
);
CREATE INDEX IF NOT EXISTS tenant_clients_tenant_idx ON tenant_clients (tenant_id) WHERE deleted_at IS NULL;

-- ── tenant_client_workspaces ────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenant_client_workspaces (
    id         BIGSERIAL PRIMARY KEY,
    client_id  BIGINT NOT NULL REFERENCES tenant_clients(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (client_id, name)
);
CREATE INDEX IF NOT EXISTS tenant_client_workspaces_client_idx ON tenant_client_workspaces (client_id);

-- ── employee extras on user_tenant_memberships ─────────────────
ALTER TABLE user_tenant_memberships
    ADD COLUMN IF NOT EXISTS title       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS active      BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS permissions JSONB NOT NULL DEFAULT '{}';

-- ── tickets ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tickets (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject      TEXT NOT NULL,
    client_email CITEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','in-progress','resolved','closed')),
    priority     TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low','medium','high','critical')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS tickets_tenant_idx    ON tickets (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS tickets_status_idx   ON tickets (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS tickets_priority_idx ON tickets (tenant_id, priority) WHERE deleted_at IS NULL;

-- ── RLS for new tenant-scoped tables ───────────────────────────
ALTER TABLE tenant_clients ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_clients FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_clients;
CREATE POLICY tenant_isolation ON tenant_clients
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE tickets ENABLE ROW LEVEL SECURITY;
ALTER TABLE tickets FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tickets;
CREATE POLICY tenant_isolation ON tickets
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');
