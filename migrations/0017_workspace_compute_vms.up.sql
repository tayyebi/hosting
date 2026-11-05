-- 0017: workspace_compute_vms — general-purpose and GPU virtual machine instances
-- Extends the resources CTI base table with VM-specific columns.

CREATE TABLE IF NOT EXISTS workspace_compute_vms (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    vm_class       TEXT NOT NULL DEFAULT 'general' CHECK (vm_class IN ('general','gpu','spot')),
    instance_type  TEXT NOT NULL DEFAULT '',
    vcpus          INTEGER NOT NULL DEFAULT 0,
    memory_gb      NUMERIC(8,2) NOT NULL DEFAULT 0,
    gpu_model      TEXT NOT NULL DEFAULT '',
    gpu_count      INTEGER NOT NULL DEFAULT 0,
    ip_address     TEXT NOT NULL DEFAULT '',
    os_image       TEXT NOT NULL DEFAULT '',
    disk_gb        INTEGER NOT NULL DEFAULT 0
);

ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS vm_class TEXT NOT NULL DEFAULT 'general';
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS instance_type TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS vcpus INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS memory_gb NUMERIC(8,2) NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS gpu_model TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS gpu_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS ip_address TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS os_image TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace_compute_vms ADD COLUMN IF NOT EXISTS disk_gb INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS workspace_compute_vms_class_idx
    ON workspace_compute_vms (vm_class);
