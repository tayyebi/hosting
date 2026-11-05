-- 0032: workspace resources — Security Network IDS + Compliance (policy, audit, regulatory) + DevOps CI/CD Source
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Intrusion Detection Sensors / Rules ───────────────────────────────────────
-- Covers security/network/ids
CREATE TABLE IF NOT EXISTS workspace_security_network_ids (
    resource_id     BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    severity        TEXT NOT NULL DEFAULT 'Medium',
    direction       TEXT NOT NULL DEFAULT 'inbound',
    matches_24h     TEXT NOT NULL DEFAULT '0',
    false_positives INTEGER NOT NULL DEFAULT 0,
    action          TEXT NOT NULL DEFAULT 'Log + alert',
    last_triggered  TEXT NOT NULL DEFAULT '—'
);
CREATE INDEX IF NOT EXISTS workspace_security_network_ids_resource_idx ON workspace_security_network_ids (resource_id);

-- ── Compliance Policies ────────────────────────────────────────────────────────
-- Covers security/compliance/policy
CREATE TABLE IF NOT EXISTS workspace_security_compliance_policy (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    severity         TEXT NOT NULL DEFAULT 'Medium',
    framework        TEXT NOT NULL DEFAULT '',
    compliant_count  INTEGER NOT NULL DEFAULT 0,
    total_count      INTEGER NOT NULL DEFAULT 0,
    drift_7d         TEXT NOT NULL DEFAULT '0',
    auto_remediate   TEXT NOT NULL DEFAULT 'No',
    last_evaluated   TEXT NOT NULL DEFAULT '—'
);
CREATE INDEX IF NOT EXISTS workspace_security_compliance_policy_resource_idx ON workspace_security_compliance_policy (resource_id);

-- ── Audit Log Entries ─────────────────────────────────────────────────────────
-- Covers security/compliance/audit
CREATE TABLE IF NOT EXISTS workspace_security_compliance_audit (
    resource_id BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    actor       TEXT NOT NULL DEFAULT '',
    source_ip   TEXT NOT NULL DEFAULT '',
    resource    TEXT NOT NULL DEFAULT '',
    region      TEXT NOT NULL DEFAULT '',
    severity    TEXT NOT NULL DEFAULT 'Info',
    audit_status TEXT NOT NULL DEFAULT 'Normal'
);
CREATE INDEX IF NOT EXISTS workspace_security_compliance_audit_resource_idx ON workspace_security_compliance_audit (resource_id);

-- ── Regulatory Frameworks ─────────────────────────────────────────────────────
-- Covers security/compliance/regulatory
CREATE TABLE IF NOT EXISTS workspace_security_compliance_regulatory (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    scope            TEXT NOT NULL DEFAULT '',
    controls_passed  INTEGER NOT NULL DEFAULT 0,
    controls_total   INTEGER NOT NULL DEFAULT 0,
    evidence_items   TEXT NOT NULL DEFAULT '0 items',
    last_assessment  TEXT NOT NULL DEFAULT '—',
    assessor         TEXT NOT NULL DEFAULT '',
    renewal          TEXT NOT NULL DEFAULT '—'
);
CREATE INDEX IF NOT EXISTS workspace_security_compliance_regulatory_resource_idx ON workspace_security_compliance_regulatory (resource_id);

-- ── Source Control Repositories ───────────────────────────────────────────────
-- Covers devops/cicd/source
CREATE TABLE IF NOT EXISTS workspace_devops_cicd_source (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    default_branch    TEXT NOT NULL DEFAULT 'main',
    language          TEXT NOT NULL DEFAULT '',
    last_push         TEXT NOT NULL DEFAULT '—',
    prs_open          INTEGER NOT NULL DEFAULT 0,
    coverage          TEXT NOT NULL DEFAULT '—',
    ci_status         TEXT NOT NULL DEFAULT 'Passing',
    branch_protection TEXT NOT NULL DEFAULT 'Enforced'
);
CREATE INDEX IF NOT EXISTS workspace_devops_cicd_source_resource_idx ON workspace_devops_cicd_source (resource_id);
