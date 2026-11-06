-- 0045_tenant_scoped_supply_hosts (rollback)

ALTER TABLE tenant_resources
    DROP COLUMN IF EXISTS disk_gb,
    DROP COLUMN IF EXISTS ram_mb,
    DROP COLUMN IF EXISTS cpu_cores,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS credential,
    DROP COLUMN IF EXISTS auth_method,
    DROP COLUMN IF EXISTS api_url;

DROP POLICY IF EXISTS tenant_isolation ON supply_host_ips;
ALTER TABLE supply_host_ips NO FORCE ROW LEVEL SECURITY;
ALTER TABLE supply_host_ips DISABLE ROW LEVEL SECURITY;
DROP INDEX IF EXISTS supply_host_ips_tenant_idx;
ALTER TABLE supply_host_ips DROP COLUMN IF EXISTS tenant_id;

DROP POLICY IF EXISTS tenant_isolation ON supply_hosts;
ALTER TABLE supply_hosts NO FORCE ROW LEVEL SECURITY;
ALTER TABLE supply_hosts DISABLE ROW LEVEL SECURITY;
DROP INDEX IF EXISTS supply_hosts_tenant_sched_idx;
DROP INDEX IF EXISTS supply_hosts_tenant_name_uq;
ALTER TABLE supply_hosts DROP COLUMN IF EXISTS tenant_id;
