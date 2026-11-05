DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'user_tenant_memberships', 'sessions', 'api_keys', 'oauth_tokens',
        'user_workspace_invitations', 'workspaces', 'workspace_memberships',
        'resources'
    ] LOOP
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
        EXECUTE format('ALTER TABLE %I DISABLE ROW LEVEL SECURITY', t);
    END LOOP;
END $$;
DROP POLICY IF EXISTS audit_read   ON audit_log;
DROP POLICY IF EXISTS audit_insert ON audit_log;
ALTER TABLE audit_log DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS users_self_or_admin ON users;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
DROP FUNCTION IF EXISTS public.current_actor_role();
DROP FUNCTION IF EXISTS public.current_tenant_id();
