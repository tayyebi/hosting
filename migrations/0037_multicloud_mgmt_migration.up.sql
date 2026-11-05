-- 0037: workspace resources — Multi-Cloud & Hybrid: Edge Computing
--        + Multi-Cloud Management: Cost, Policy Governance, Workload Portability
--        + Multi-Cloud Migration: Assessment & Discovery
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Multi-Cloud: Edge Computing ───────────────────────────────────────────────
-- Covers multicloud/hybrid/edge
CREATE TABLE IF NOT EXISTS workspace_multicloud_edge_nodes (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    location      TEXT    NOT NULL DEFAULT '',
    provider      TEXT    NOT NULL DEFAULT '',
    cpu_cores     INTEGER NOT NULL DEFAULT 0,
    memory_gb     INTEGER NOT NULL DEFAULT 0,
    workloads     INTEGER NOT NULL DEFAULT 0,
    uptime        TEXT    NOT NULL DEFAULT '—',
    latency       TEXT    NOT NULL DEFAULT '—',
    last_contact  TEXT    NOT NULL DEFAULT '—',
    node_status   TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_edge_nodes_resource_idx ON workspace_multicloud_edge_nodes (resource_id);

-- ── Multi-Cloud Management: Cost Management ───────────────────────────────────
-- Covers multicloud/mgmt/cost
CREATE TABLE IF NOT EXISTS workspace_multicloud_cost_profiles (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    provider          TEXT    NOT NULL DEFAULT '',
    account           TEXT    NOT NULL DEFAULT '',
    monthly_budget    TEXT    NOT NULL DEFAULT '$0',
    current_spend     TEXT    NOT NULL DEFAULT '$0',
    forecast          TEXT    NOT NULL DEFAULT '—',
    savings_found     TEXT    NOT NULL DEFAULT '$0',
    alerts            INTEGER NOT NULL DEFAULT 0,
    optimization_pct  TEXT    NOT NULL DEFAULT '—',
    profile_status    TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_cost_profiles_resource_idx ON workspace_multicloud_cost_profiles (resource_id);

-- ── Multi-Cloud Management: Policy Governance ─────────────────────────────────
-- Covers multicloud/mgmt/policy
CREATE TABLE IF NOT EXISTS workspace_multicloud_governance_policies (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    category       TEXT    NOT NULL DEFAULT '',
    scope          TEXT    NOT NULL DEFAULT '',
    enforcement    TEXT    NOT NULL DEFAULT 'Audit',
    compliant      INTEGER NOT NULL DEFAULT 0,
    non_compliant  INTEGER NOT NULL DEFAULT 0,
    exemptions     INTEGER NOT NULL DEFAULT 0,
    severity       TEXT    NOT NULL DEFAULT 'Medium',
    last_eval      TEXT    NOT NULL DEFAULT '—',
    policy_status  TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_governance_policies_resource_idx ON workspace_multicloud_governance_policies (resource_id);

-- ── Multi-Cloud Management: Workload Portability ──────────────────────────────
-- Covers multicloud/mgmt/workload
CREATE TABLE IF NOT EXISTS workspace_multicloud_workload_profiles (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    source_provider   TEXT    NOT NULL DEFAULT '',
    target_provider   TEXT    NOT NULL DEFAULT '',
    workload_type     TEXT    NOT NULL DEFAULT '',
    components        INTEGER NOT NULL DEFAULT 0,
    portability       TEXT    NOT NULL DEFAULT '—',
    drift_status      TEXT    NOT NULL DEFAULT 'Pending assessment',
    last_sync         TEXT    NOT NULL DEFAULT '—',
    workload_status   TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_workload_profiles_resource_idx ON workspace_multicloud_workload_profiles (resource_id);

-- ── Multi-Cloud Migration: Assessment & Discovery ─────────────────────────────
-- Covers multicloud/migration/assessment
CREATE TABLE IF NOT EXISTS workspace_multicloud_migration_assessments (
    resource_id         BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    source_env          TEXT    NOT NULL DEFAULT '',
    target_env          TEXT    NOT NULL DEFAULT '',
    discovered_assets   INTEGER NOT NULL DEFAULT 0,
    ready_to_migrate    INTEGER NOT NULL DEFAULT 0,
    needs_remediation   INTEGER NOT NULL DEFAULT 0,
    unsupported         INTEGER NOT NULL DEFAULT 0,
    estimated_cost      TEXT    NOT NULL DEFAULT '—',
    complexity          TEXT    NOT NULL DEFAULT 'Pending',
    last_scan           TEXT    NOT NULL DEFAULT '—',
    assessment_status   TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_multicloud_migration_assessments_resource_idx ON workspace_multicloud_migration_assessments (resource_id);
