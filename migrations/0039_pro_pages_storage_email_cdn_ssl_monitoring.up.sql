-- Storage & Backups: S3-compatible object storage buckets
CREATE TABLE IF NOT EXISTS workspace_s3_buckets (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    region         TEXT NOT NULL DEFAULT '',
    storage_class  TEXT NOT NULL DEFAULT 'Standard',
    size_gb        NUMERIC(12,2) NOT NULL DEFAULT 0,
    object_count   BIGINT NOT NULL DEFAULT 0,
    versioning     BOOLEAN NOT NULL DEFAULT false,
    encryption     TEXT NOT NULL DEFAULT 'AES-256',
    access_policy  TEXT NOT NULL DEFAULT 'Private',
    endpoint       TEXT NOT NULL DEFAULT '',
    cdn_enabled    BOOLEAN NOT NULL DEFAULT false,
    lifecycle_rule TEXT NOT NULL DEFAULT 'None',
    bucket_status  TEXT NOT NULL DEFAULT 'active'
);

-- Email & Messaging: Transactional email sending services
CREATE TABLE IF NOT EXISTS workspace_transactional_emails (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    domain         TEXT NOT NULL DEFAULT '',
    sender_address TEXT NOT NULL DEFAULT '',
    dkim_status    TEXT NOT NULL DEFAULT 'Pending',
    spf_status     TEXT NOT NULL DEFAULT 'Pending',
    dmarc_status   TEXT NOT NULL DEFAULT 'None',
    daily_sent     INTEGER NOT NULL DEFAULT 0,
    daily_limit    INTEGER NOT NULL DEFAULT 1000,
    delivery_rate  NUMERIC(5,2) NOT NULL DEFAULT 0,
    bounce_rate    NUMERIC(5,2) NOT NULL DEFAULT 0,
    open_rate      NUMERIC(5,2) NOT NULL DEFAULT 0,
    webhook_url    TEXT NOT NULL DEFAULT '',
    api_key_prefix TEXT NOT NULL DEFAULT '',
    email_status   TEXT NOT NULL DEFAULT 'active'
);

-- Content Delivery: CDN Points of Presence
CREATE TABLE IF NOT EXISTS workspace_cdn_pops (
    resource_id    BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    location       TEXT NOT NULL DEFAULT '',
    provider       TEXT NOT NULL DEFAULT '',
    cache_hit_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    bandwidth_tb   NUMERIC(10,2) NOT NULL DEFAULT 0,
    requests_m     NUMERIC(10,2) NOT NULL DEFAULT 0,
    avg_latency_ms INTEGER NOT NULL DEFAULT 0,
    capacity       TEXT NOT NULL DEFAULT '',
    tls_version    TEXT NOT NULL DEFAULT '1.3',
    http2          BOOLEAN NOT NULL DEFAULT true,
    http3          BOOLEAN NOT NULL DEFAULT false,
    pop_status     TEXT NOT NULL DEFAULT 'active'
);

-- Security & SSL: TLS certificate management
CREATE TABLE IF NOT EXISTS workspace_ssl_certs (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    san           TEXT NOT NULL DEFAULT '',
    issuer        TEXT NOT NULL DEFAULT '',
    cert_type     TEXT NOT NULL DEFAULT 'DV',
    auto_renew    BOOLEAN NOT NULL DEFAULT true,
    issued_at     TEXT NOT NULL DEFAULT '',
    expires_at    TEXT NOT NULL DEFAULT '',
    key_algorithm TEXT NOT NULL DEFAULT 'ECDSA',
    key_size      INTEGER NOT NULL DEFAULT 256,
    protocol      TEXT NOT NULL DEFAULT 'TLS 1.3',
    ocsp_stapling BOOLEAN NOT NULL DEFAULT true,
    ct_logged     BOOLEAN NOT NULL DEFAULT true,
    cert_status   TEXT NOT NULL DEFAULT 'active'
);

-- Monitoring & Alerts: Alert rule definitions
CREATE TABLE IF NOT EXISTS workspace_monitoring_alerts (
    resource_id     BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    metric          TEXT NOT NULL DEFAULT '',
    condition       TEXT NOT NULL DEFAULT '>',
    threshold       NUMERIC(12,4) NOT NULL DEFAULT 0,
    alert_window    TEXT NOT NULL DEFAULT '5m',
    severity        TEXT NOT NULL DEFAULT 'warning',
    notify_channels TEXT NOT NULL DEFAULT '',
    eval_interval   TEXT NOT NULL DEFAULT '30s',
    state           TEXT NOT NULL DEFAULT 'ok',
    last_triggered  TEXT NOT NULL DEFAULT '',
    silenced        BOOLEAN NOT NULL DEFAULT false,
    alert_status    TEXT NOT NULL DEFAULT 'active'
);
