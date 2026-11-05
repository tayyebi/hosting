DROP TABLE IF EXISTS tenant_client_workspaces;
DROP TABLE IF EXISTS tenant_clients;
ALTER TABLE user_tenant_memberships
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS active,
    DROP COLUMN IF EXISTS permissions;
DROP TABLE IF EXISTS tickets;
