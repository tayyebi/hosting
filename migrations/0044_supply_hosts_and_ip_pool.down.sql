DROP TABLE IF EXISTS workspace_assigned_ips;
DROP TABLE IF EXISTS supply_host_ips;

ALTER TABLE resources
    DROP COLUMN IF EXISTS retail_price_cents,
    DROP COLUMN IF EXISTS supply_cost_cents,
    DROP COLUMN IF EXISTS sku_code,
    DROP COLUMN IF EXISTS external_ref,
    DROP COLUMN IF EXISTS supply_host_id,
    DROP COLUMN IF EXISTS parent_resource_id;

DROP TABLE IF EXISTS supply_hosts;
