-- 0030: workspace resources — MLOps Registry + ML Monitoring + Security IAM (SSO, MFA, RBAC)
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── AI Model Registry ─────────────────────────────────────────────────────────
-- Covers ai/mlops/registry
CREATE TABLE IF NOT EXISTS workspace_ai_model_registry (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    version       TEXT NOT NULL DEFAULT 'v1.0',
    format        TEXT NOT NULL DEFAULT 'ONNX',
    size_bytes    TEXT NOT NULL DEFAULT '0 MB',
    stage         TEXT NOT NULL DEFAULT 'staging',
    requests_s    TEXT NOT NULL DEFAULT '0/s',
    p99_latency   TEXT NOT NULL DEFAULT '—',
    owner         TEXT NOT NULL DEFAULT '',
    last_used     TEXT NOT NULL DEFAULT '—'
);
CREATE INDEX IF NOT EXISTS workspace_ai_model_registry_resource_idx ON workspace_ai_model_registry (resource_id);

-- ── AI ML Monitors ────────────────────────────────────────────────────────────
-- Covers ai/mlops/monitoring
CREATE TABLE IF NOT EXISTS workspace_ai_ml_monitors (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    drift_psi     TEXT NOT NULL DEFAULT '—',
    confidence    TEXT NOT NULL DEFAULT '—',
    requests_s    TEXT NOT NULL DEFAULT '0/s',
    p99_latency   TEXT NOT NULL DEFAULT '—',
    error_rate    TEXT NOT NULL DEFAULT '—',
    last_retrain  TEXT NOT NULL DEFAULT '—',
    monitor_status TEXT NOT NULL DEFAULT 'healthy'
);
CREATE INDEX IF NOT EXISTS workspace_ai_ml_monitors_resource_idx ON workspace_ai_ml_monitors (resource_id);

-- ── Security SSO Providers ────────────────────────────────────────────────────
-- Covers security/iam/sso
CREATE TABLE IF NOT EXISTS workspace_security_sso_providers (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    protocol      TEXT NOT NULL DEFAULT 'SAML',
    domain        TEXT NOT NULL DEFAULT '',
    users_count   INTEGER NOT NULL DEFAULT 0,
    logins_24h    INTEGER NOT NULL DEFAULT 0,
    scim          TEXT NOT NULL DEFAULT 'Manual',
    cert_expires  TEXT NOT NULL DEFAULT '—',
    last_rotation TEXT NOT NULL DEFAULT '—'
);
CREATE INDEX IF NOT EXISTS workspace_security_sso_providers_resource_idx ON workspace_security_sso_providers (resource_id);

-- ── Security MFA Users ────────────────────────────────────────────────────────
-- Covers security/iam/mfa
CREATE TABLE IF NOT EXISTS workspace_security_mfa_users (
    resource_id     BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    primary_method  TEXT NOT NULL DEFAULT 'TOTP',
    backup_method   TEXT NOT NULL DEFAULT '',
    roles           TEXT NOT NULL DEFAULT '',
    recovery_codes  TEXT NOT NULL DEFAULT '—',
    devices         INTEGER NOT NULL DEFAULT 1,
    last_mfa        TEXT NOT NULL DEFAULT '—',
    mfa_status      TEXT NOT NULL DEFAULT 'medium'
);
CREATE INDEX IF NOT EXISTS workspace_security_mfa_users_resource_idx ON workspace_security_mfa_users (resource_id);

-- ── Security RBAC Roles ───────────────────────────────────────────────────────
-- Covers security/iam/rbac
CREATE TABLE IF NOT EXISTS workspace_security_rbac_roles (
    resource_id     BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    role_type       TEXT NOT NULL DEFAULT 'Custom',
    users_count     INTEGER NOT NULL DEFAULT 0,
    permissions     TEXT NOT NULL DEFAULT '',
    scope           TEXT NOT NULL DEFAULT 'Workspace',
    last_modified   TEXT NOT NULL DEFAULT '—',
    review_status   TEXT NOT NULL DEFAULT 'Pending review'
);
CREATE INDEX IF NOT EXISTS workspace_security_rbac_roles_resource_idx ON workspace_security_rbac_roles (resource_id);
