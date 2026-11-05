-- 0044_supply_hosts_and_ip_pool:
--   1. `supply_hosts` — platform infrastructure nodes (starting with Linux
--      Nested VPS hosts onboarded via SSH + automated QEMU/KVM/libvirt
--      virtualization stack installation).
--   2. `supply_host_ips` — manual assignable IPv4/IPv6 pool per server,
--      allocatable to customers purchasing VMs on that host.
--   3. Unified payable resource tree columns on `resources`
--      (`parent_resource_id`, `supply_host_id`, `external_ref`, `sku_code`,
--       `supply_cost_cents`, `retail_price_cents`) plus `workspace_assigned_ips`.

-- ── 1. Supply Hosts (Infrastructure Inventory) ──────────────────────────────
CREATE TABLE IF NOT EXISTS supply_hosts (
    id                  BIGSERIAL PRIMARY KEY,
    name                TEXT        NOT NULL UNIQUE,
    hostname_or_ip      TEXT        NOT NULL,
    ssh_port            INTEGER     NOT NULL DEFAULT 22 CHECK (ssh_port BETWEEN 1 AND 65535),
    ssh_user            TEXT        NOT NULL DEFAULT 'root',
    auth_method         TEXT        NOT NULL DEFAULT 'ssh_private_key'
                        CHECK (auth_method IN ('ssh_private_key','ssh_password','api_token')),
    ssh_credential      TEXT        NOT NULL DEFAULT '',
    driver_kind         TEXT        NOT NULL DEFAULT 'linux_nested_qemu',
    ownership_type      TEXT        NOT NULL DEFAULT 'leased_nested_vps',
    region_id           BIGINT      REFERENCES regions(id),
    region_slug         TEXT        NOT NULL DEFAULT '',
    bridge_interface    TEXT        NOT NULL DEFAULT 'virbr0',
    storage_pool_path   TEXT        NOT NULL DEFAULT '/var/lib/libvirt/images/platform',
    cpu_cores           INTEGER     NOT NULL DEFAULT 0,
    ram_mb              INTEGER     NOT NULL DEFAULT 0,
    disk_gb             INTEGER     NOT NULL DEFAULT 0,
    kvm_available       BOOLEAN     NOT NULL DEFAULT false,
    os_info             TEXT        NOT NULL DEFAULT '',
    hypervisor_version  TEXT        NOT NULL DEFAULT '',
    monthly_cost_cents  BIGINT      NOT NULL DEFAULT 0,
    currency            TEXT        NOT NULL DEFAULT 'USD',
    virt_stack_status   TEXT        NOT NULL DEFAULT 'unprovisioned'
                        CHECK (virt_stack_status IN ('unprovisioned','provisioning','ready','failed')),
    inventory_status    TEXT        NOT NULL DEFAULT 'pending'
                        CHECK (inventory_status IN ('pending','online','maintenance','offline')),
    provision_log       TEXT        NOT NULL DEFAULT '',
    last_error          TEXT        NOT NULL DEFAULT '',
    provisioned_at      TIMESTAMPTZ,
    last_checked_at     TIMESTAMPTZ,
    created_by          BIGINT      REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS supply_hosts_schedulable_idx
    ON supply_hosts (inventory_status, virt_stack_status, region_slug)
    WHERE deleted_at IS NULL;

-- ── 2. Unified Payable Resource Tree & Cost Columns on `resources` ──────────
ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS parent_resource_id BIGINT REFERENCES resources(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS supply_host_id     BIGINT REFERENCES supply_hosts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS external_ref       TEXT   NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sku_code           TEXT   NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS supply_cost_cents  BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS retail_price_cents BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS resources_parent_idx
    ON resources (parent_resource_id)
    WHERE parent_resource_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS resources_supply_host_idx
    ON resources (supply_host_id)
    WHERE supply_host_id IS NOT NULL AND deleted_at IS NULL;

-- ── 3. Manual Assignable IP Pool per Server (`supply_host_ips`) ─────────────
CREATE TABLE IF NOT EXISTS supply_host_ips (
    id                      BIGSERIAL PRIMARY KEY,
    host_id                 BIGINT      NOT NULL REFERENCES supply_hosts(id) ON DELETE CASCADE,
    ip_address              TEXT        NOT NULL UNIQUE,
    cidr_prefix             INTEGER     NOT NULL DEFAULT 24 CHECK (cidr_prefix BETWEEN 1 AND 128),
    gateway                 TEXT        NOT NULL DEFAULT '',
    mac_address             TEXT        NOT NULL DEFAULT '',
    ip_version              INTEGER     NOT NULL DEFAULT 4 CHECK (ip_version IN (4, 6)),
    monthly_cost_cents      BIGINT      NOT NULL DEFAULT 0,
    retail_price_cents      BIGINT      NOT NULL DEFAULT 200,
    currency                TEXT        NOT NULL DEFAULT 'USD',
    status                  TEXT        NOT NULL DEFAULT 'available'
                            CHECK (status IN ('available','reserved','assigned','quarantined')),
    assigned_resource_id    BIGINT      REFERENCES resources(id) ON DELETE SET NULL,
    assigned_vm_resource_id BIGINT      REFERENCES resources(id) ON DELETE SET NULL,
    assigned_at             TIMESTAMPTZ,
    notes                   TEXT        NOT NULL DEFAULT '',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS supply_host_ips_available_idx
    ON supply_host_ips (host_id, status)
    WHERE status = 'available';

-- ── 4. CTI Payable IP Resource (`workspace_assigned_ips`) ───────────────────
CREATE TABLE IF NOT EXISTS workspace_assigned_ips (
    resource_id  BIGINT  PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    supply_ip_id BIGINT  REFERENCES supply_host_ips(id) ON DELETE SET NULL,
    ip_address   TEXT    NOT NULL,
    cidr_prefix  INTEGER NOT NULL DEFAULT 24,
    gateway      TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_assigned_ips_supply_idx
    ON workspace_assigned_ips (supply_ip_id);
