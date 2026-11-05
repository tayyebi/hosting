-- 0033: workspace resources — DevOps CI/CD Build + Deploy + IaC Templates + IaC Config + API Gateway
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── CI/CD Build Pipelines ─────────────────────────────────────────────────────
-- Covers devops/cicd/build
CREATE TABLE IF NOT EXISTS workspace_devops_cicd_build (
    resource_id BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    repo        TEXT NOT NULL DEFAULT '',
    trigger     TEXT NOT NULL DEFAULT 'push to main',
    runtime     TEXT NOT NULL DEFAULT 'Docker',
    last_run    TEXT NOT NULL DEFAULT '—',
    duration    TEXT NOT NULL DEFAULT '—',
    cache_hit   TEXT NOT NULL DEFAULT '—',
    status      TEXT NOT NULL DEFAULT 'pending'
);
CREATE INDEX IF NOT EXISTS workspace_devops_cicd_build_resource_idx ON workspace_devops_cicd_build (resource_id);

-- ── CI/CD Deploy Pipelines ────────────────────────────────────────────────────
-- Covers devops/cicd/deploy
CREATE TABLE IF NOT EXISTS workspace_devops_cicd_deploy (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    app          TEXT NOT NULL DEFAULT '',
    environment  TEXT NOT NULL DEFAULT 'staging',
    strategy     TEXT NOT NULL DEFAULT 'rolling',
    traffic      TEXT NOT NULL DEFAULT '—',
    error_rate   TEXT NOT NULL DEFAULT '—',
    latency_p99  TEXT NOT NULL DEFAULT '—',
    duration     TEXT NOT NULL DEFAULT '—',
    deploy_status TEXT NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_devops_cicd_deploy_resource_idx ON workspace_devops_cicd_deploy (resource_id);

-- ── IaC Templates / Stacks ────────────────────────────────────────────────────
-- Covers devops/iac/templates
CREATE TABLE IF NOT EXISTS workspace_devops_iac_templates (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    tool           TEXT NOT NULL DEFAULT 'Terraform',
    resource_count INTEGER NOT NULL DEFAULT 0,
    state_backend  TEXT NOT NULL DEFAULT '',
    last_applied   TEXT NOT NULL DEFAULT '—',
    plan_output    TEXT NOT NULL DEFAULT '—',
    drift_status   TEXT NOT NULL DEFAULT 'Clean'
);
CREATE INDEX IF NOT EXISTS workspace_devops_iac_templates_resource_idx ON workspace_devops_iac_templates (resource_id);

-- ── IaC Config Management ─────────────────────────────────────────────────────
-- Covers devops/iac/config
CREATE TABLE IF NOT EXISTS workspace_devops_iac_config (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    tool         TEXT NOT NULL DEFAULT 'Ansible',
    host_count   INTEGER NOT NULL DEFAULT 0,
    inventory    TEXT NOT NULL DEFAULT '',
    last_apply   TEXT NOT NULL DEFAULT '—',
    changed      TEXT NOT NULL DEFAULT '0 changed',
    compliance   TEXT NOT NULL DEFAULT '—',
    cfg_status   TEXT NOT NULL DEFAULT 'In sync'
);
CREATE INDEX IF NOT EXISTS workspace_devops_iac_config_resource_idx ON workspace_devops_iac_config (resource_id);

-- ── API Gateway ───────────────────────────────────────────────────────────────
-- Covers devops/api/gateway
CREATE TABLE IF NOT EXISTS workspace_devops_api_gateway (
    resource_id BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    api_version TEXT NOT NULL DEFAULT 'v1.0.0',
    protocol    TEXT NOT NULL DEFAULT 'REST',
    calls_24h   TEXT NOT NULL DEFAULT '—',
    latency_p95 TEXT NOT NULL DEFAULT '—',
    rate_limit  TEXT NOT NULL DEFAULT '—',
    auth_type   TEXT NOT NULL DEFAULT 'JWT',
    error_rate  TEXT NOT NULL DEFAULT '—',
    gw_status   TEXT NOT NULL DEFAULT 'Live'
);
CREATE INDEX IF NOT EXISTS workspace_devops_api_gateway_resource_idx ON workspace_devops_api_gateway (resource_id);
