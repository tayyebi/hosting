-- 0005_rls_policies: Row-Level Security for every tenant-scoped table
-- introduced so far (phases 1–1). Sessions, actions by platform_app flow
-- through the `tenant_isolation` policy; platform_admin bypasses RLS via
-- its BYPASSRLS attribute (set in db-init.sh).
--
-- The app sets two custom GUCs per request, inside a transaction:
--   SET LOCAL app.current_tenant_id = '<id>';
--   SET LOCAL app.current_role      = 'tenant'|'user'|'sysadmin'|'system';
--
-- Missing GUCs (unconfigured connection) → policy returns 0 rows, which is
-- the correct default-deny for a misbehaving / forgotten wiring.

-- ── helper: read current tenant GUC, NULL-safe ────────────────
CREATE OR REPLACE FUNCTION public.current_tenant_id() RETURNS bigint AS $$
    SELECT NULLIF(current_setting('app.current_tenant_id', true), '')::bigint
$$ LANGUAGE SQL STABLE;

CREATE OR REPLACE FUNCTION public.current_actor_role() RETURNS text AS $$
    SELECT NULLIF(current_setting('app.current_role', true), '')
$$ LANGUAGE SQL STABLE;

-- ── enable RLS + attach uniform policy ────────────────────────
-- Macro: enable RLS on a table and attach the standard tenant_isolation policy.
DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'user_tenant_memberships',
        'sessions',
        'api_keys',
        'oauth_tokens',
        'user_workspace_invitations',
        'workspaces',
        'workspace_memberships',
        'resources'
    ] LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON %I
                USING (
                    tenant_id = public.current_tenant_id()
                    OR public.current_actor_role() = ''sysadmin''
                )
                WITH CHECK (
                    tenant_id = public.current_tenant_id()
                    OR public.current_actor_role() = ''sysadmin''
                )', t);
    END LOOP;
END $$;

-- audit_log: readable by tenant (their own entries) + sysadmin (all); writes
-- come from the app role, insert with tenant_id explicitly set — enforced in
-- app code, policy simply scopes reads.
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS audit_read ON audit_log;
CREATE POLICY audit_read ON audit_log FOR SELECT
    USING (
        tenant_id = public.current_tenant_id()
        OR tenant_id IS NULL      -- system actions, visible to sysadmin only via below
        OR public.current_actor_role() = 'sysadmin'
    );
DROP POLICY IF EXISTS audit_insert ON audit_log;
CREATE POLICY audit_insert ON audit_log FOR INSERT
    WITH CHECK (
        tenant_id = public.current_tenant_id()
        OR public.current_actor_role() IN ('sysadmin','system')
    );

-- users table: a user can see themselves; sysadmin sees all. Memberships
-- drive tenant-side visibility; plain SELECTs against users from tenant
-- context are discouraged — join via user_tenant_memberships.
-- (No tenant_id column on users, so we take a different tack.)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE users FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS users_self_or_admin ON users;
CREATE POLICY users_self_or_admin ON users
    USING (
        id = NULLIF(current_setting('app.current_user_id', true), '')::bigint
        OR public.current_actor_role() = 'sysadmin'
        OR EXISTS (
            SELECT 1 FROM user_tenant_memberships m
            WHERE m.user_id = users.id
              AND m.tenant_id = public.current_tenant_id()
        )
    );
