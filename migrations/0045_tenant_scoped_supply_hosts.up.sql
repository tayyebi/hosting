-- 0045_tenant_scoped_supply_hosts:
--   Moves supply hosts and assignable IP pools from system-global scope into
--   strict per-tenant scope (`/t/{slug}/resources/`) with Row-Level Security,
--   and enriches `tenant_resources` so every Resource Tree category in
--   `/t/{slug}/resources/` persists cleanly in PostgreSQL without dummy data.

-- ── 1. Tenant-scope `supply_hosts` + RLS ────────────────────────────────────
ALTER TABLE supply_hosts
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES tenants(id) ON DELETE CASCADE;

UPDATE supply_hosts
SET tenant_id = (SELECT id FROM tenants WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 1)
WHERE tenant_id IS NULL
  AND EXISTS (SELECT 1 FROM tenants WHERE deleted_at IS NULL);

DELETE FROM supply_hosts WHERE tenant_id IS NULL;

ALTER TABLE supply_hosts
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE supply_hosts
    DROP CONSTRAINT IF EXISTS supply_hosts_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS supply_hosts_tenant_name_uq
    ON supply_hosts (tenant_id, lower(name))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS supply_hosts_tenant_sched_idx
    ON supply_hosts (tenant_id, inventory_status, virt_stack_status, region_slug)
    WHERE deleted_at IS NULL;

ALTER TABLE supply_hosts ENABLE ROW LEVEL SECURITY;
ALTER TABLE supply_hosts FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON supply_hosts;
CREATE POLICY tenant_isolation ON supply_hosts
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

-- ── 2. Tenant-scope `supply_host_ips` + RLS ─────────────────────────────────
ALTER TABLE supply_host_ips
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES tenants(id) ON DELETE CASCADE;

UPDATE supply_host_ips ip
SET tenant_id = h.tenant_id
FROM supply_hosts h
WHERE ip.host_id = h.id
  AND ip.tenant_id IS NULL;

DELETE FROM supply_host_ips WHERE tenant_id IS NULL;

ALTER TABLE supply_host_ips
    ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS supply_host_ips_tenant_idx
    ON supply_host_ips (tenant_id, status);

ALTER TABLE supply_host_ips ENABLE ROW LEVEL SECURITY;
ALTER TABLE supply_host_ips FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON supply_host_ips;
CREATE POLICY tenant_isolation ON supply_host_ips
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

-- ── 3. Enrich `tenant_resources` for full Resource Tree categories ──────────
ALTER TABLE tenant_resources
    ADD COLUMN IF NOT EXISTS api_url     TEXT    NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS auth_method TEXT    NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS credential  TEXT    NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS tags        TEXT    NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS cpu_cores   INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ram_mb      INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS disk_gb     INTEGER NOT NULL DEFAULT 0;
