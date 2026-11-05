-- 0031: workspace resources — Security Data (enc-rest, enc-transit, kms) + Network Security (waf, ddos)
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Encryption at Rest ────────────────────────────────────────────────────────
-- Covers security/data/enc-rest
CREATE TABLE IF NOT EXISTS workspace_security_data_enc_rest (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL DEFAULT 'Block volume',
    algorithm     TEXT NOT NULL DEFAULT 'AES-256',
    kms_key       TEXT NOT NULL DEFAULT '',
    key_age       TEXT NOT NULL DEFAULT '—',
    last_verified TEXT NOT NULL DEFAULT '—',
    compliance    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS workspace_security_data_enc_rest_resource_idx ON workspace_security_data_enc_rest (resource_id);

-- ── Encryption in Transit ─────────────────────────────────────────────────────
-- Covers security/data/enc-transit
CREATE TABLE IF NOT EXISTS workspace_security_data_enc_transit (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    protocol     TEXT NOT NULL DEFAULT 'TLS 1.3',
    cipher       TEXT NOT NULL DEFAULT '',
    cert_issuer  TEXT NOT NULL DEFAULT '',
    grade        TEXT NOT NULL DEFAULT 'A+',
    expires      TEXT NOT NULL DEFAULT '—',
    hsts         TEXT NOT NULL DEFAULT 'Enabled',
    last_scan    TEXT NOT NULL DEFAULT '—'
);
CREATE INDEX IF NOT EXISTS workspace_security_data_enc_transit_resource_idx ON workspace_security_data_enc_transit (resource_id);

-- ── KMS Keys ──────────────────────────────────────────────────────────────────
-- Covers security/data/kms
CREATE TABLE IF NOT EXISTS workspace_security_data_kms (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    algorithm      TEXT NOT NULL DEFAULT 'AES-256',
    key_type       TEXT NOT NULL DEFAULT 'Symmetric',
    age_days       TEXT NOT NULL DEFAULT '—',
    rotation       TEXT NOT NULL DEFAULT 'Auto · 90d',
    operations_24h TEXT NOT NULL DEFAULT '0',
    grants_count   INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS workspace_security_data_kms_resource_idx ON workspace_security_data_kms (resource_id);

-- ── WAF Rules ─────────────────────────────────────────────────────────────────
-- Covers security/network/waf
CREATE TABLE IF NOT EXISTS workspace_security_network_waf (
    resource_id     BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    priority        INTEGER NOT NULL DEFAULT 1,
    action          TEXT NOT NULL DEFAULT 'block',
    source          TEXT NOT NULL DEFAULT 'Any origin',
    matches_24h     TEXT NOT NULL DEFAULT '0',
    false_positives INTEGER NOT NULL DEFAULT 0,
    severity        TEXT NOT NULL DEFAULT 'Medium',
    last_triggered  TEXT NOT NULL DEFAULT '—'
);
CREATE INDEX IF NOT EXISTS workspace_security_network_waf_resource_idx ON workspace_security_network_waf (resource_id);

-- ── DDoS Protection Profiles ──────────────────────────────────────────────────
-- Covers security/network/ddos
CREATE TABLE IF NOT EXISTS workspace_security_network_ddos (
    resource_id     BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    method          TEXT NOT NULL DEFAULT 'Always-on anycast',
    threshold       TEXT NOT NULL DEFAULT '',
    attacks_30d     TEXT NOT NULL DEFAULT '0',
    last_triggered  TEXT NOT NULL DEFAULT '—',
    false_positives INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS workspace_security_network_ddos_resource_idx ON workspace_security_network_ddos (resource_id);
