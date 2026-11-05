-- 0041: Pro feature pages — Document Database, Application Load Balancer,
--        Vector Search, Identity Platform, Cloud WAF.
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Document Database (databases/nosql/document-db) ──────────────────────────
CREATE TABLE IF NOT EXISTS workspace_pro_document_dbs (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine            TEXT          NOT NULL DEFAULT 'MongoDB',
    engine_version    TEXT          NOT NULL DEFAULT '7.0',
    cluster_nodes     INTEGER       NOT NULL DEFAULT 1,
    storage_gb        INTEGER       NOT NULL DEFAULT 50,
    document_count    BIGINT        NOT NULL DEFAULT 0,
    collection_count  INTEGER       NOT NULL DEFAULT 0,
    reads_per_sec     INTEGER       NOT NULL DEFAULT 0,
    writes_per_sec    INTEGER       NOT NULL DEFAULT 0,
    avg_latency_ms    NUMERIC(8,2)  NOT NULL DEFAULT 0,
    encrypted         BOOLEAN       NOT NULL DEFAULT true,
    backup_enabled    BOOLEAN       NOT NULL DEFAULT false,
    backup_retention  INTEGER       NOT NULL DEFAULT 0,
    region            TEXT          NOT NULL DEFAULT 'us-east-1',
    multi_az          BOOLEAN       NOT NULL DEFAULT false,
    cost_per_month    NUMERIC(10,2) NOT NULL DEFAULT 0,
    db_status         TEXT          NOT NULL DEFAULT 'provisioning'
);
CREATE INDEX IF NOT EXISTS workspace_pro_document_dbs_resource_idx
    ON workspace_pro_document_dbs (resource_id);

-- ── Application Load Balancer (networking/load-balancer/application-lb) ──────
CREATE TABLE IF NOT EXISTS workspace_pro_app_load_balancers (
    resource_id        BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    protocol           TEXT          NOT NULL DEFAULT 'HTTPS',
    listener_ports     TEXT          NOT NULL DEFAULT '443',
    target_count       INTEGER       NOT NULL DEFAULT 0,
    healthy_targets    INTEGER       NOT NULL DEFAULT 0,
    rps                INTEGER       NOT NULL DEFAULT 0,
    active_conns       INTEGER       NOT NULL DEFAULT 0,
    bandwidth_gbps     NUMERIC(8,2)  NOT NULL DEFAULT 0,
    tls_version        TEXT          NOT NULL DEFAULT 'TLS 1.3',
    cert_domain        TEXT          NOT NULL DEFAULT '',
    cert_expiry        TEXT          NOT NULL DEFAULT '',
    waf_enabled        BOOLEAN       NOT NULL DEFAULT false,
    access_logs        BOOLEAN       NOT NULL DEFAULT false,
    sticky_session     BOOLEAN       NOT NULL DEFAULT false,
    idle_timeout_sec   INTEGER       NOT NULL DEFAULT 60,
    region             TEXT          NOT NULL DEFAULT 'us-east-1',
    cost_per_month     NUMERIC(10,2) NOT NULL DEFAULT 0,
    lb_status          TEXT          NOT NULL DEFAULT 'provisioning'
);
CREATE INDEX IF NOT EXISTS workspace_pro_app_load_balancers_resource_idx
    ON workspace_pro_app_load_balancers (resource_id);

-- ── Vector Search (ai/vector-search/embeddings) ───────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_pro_vector_collections (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    dimensions        INTEGER       NOT NULL DEFAULT 1536,
    vector_count      BIGINT        NOT NULL DEFAULT 0,
    index_type        TEXT          NOT NULL DEFAULT 'HNSW',
    index_params      TEXT          NOT NULL DEFAULT 'M=32, efConstruction=200',
    metric            TEXT          NOT NULL DEFAULT 'cosine',
    queries_per_sec   INTEGER       NOT NULL DEFAULT 0,
    avg_latency_ms    NUMERIC(8,2)  NOT NULL DEFAULT 0,
    recall_at_10      NUMERIC(6,2)  NOT NULL DEFAULT 0,
    storage_gb        NUMERIC(10,2) NOT NULL DEFAULT 0,
    replicas          INTEGER       NOT NULL DEFAULT 1,
    embed_model       TEXT          NOT NULL DEFAULT '',
    region            TEXT          NOT NULL DEFAULT 'us-east-1',
    cost_per_month    NUMERIC(10,2) NOT NULL DEFAULT 0,
    coll_status       TEXT          NOT NULL DEFAULT 'creating'
);
CREATE INDEX IF NOT EXISTS workspace_pro_vector_collections_resource_idx
    ON workspace_pro_vector_collections (resource_id);

-- ── Identity Platform (security/iam/identity-platform) ───────────────────────
CREATE TABLE IF NOT EXISTS workspace_pro_identity_providers (
    resource_id        BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    protocol           TEXT          NOT NULL DEFAULT 'OIDC',
    issuer             TEXT          NOT NULL DEFAULT '',
    client_id          TEXT          NOT NULL DEFAULT '',
    domain             TEXT          NOT NULL DEFAULT '',
    user_count         INTEGER       NOT NULL DEFAULT 0,
    group_count        INTEGER       NOT NULL DEFAULT 0,
    active_sessions    INTEGER       NOT NULL DEFAULT 0,
    mfa_enforced       BOOLEAN       NOT NULL DEFAULT false,
    mfa_method         TEXT          NOT NULL DEFAULT 'TOTP',
    password_policy    TEXT          NOT NULL DEFAULT 'standard',
    session_timeout    INTEGER       NOT NULL DEFAULT 480,
    jit_provisioning   BOOLEAN       NOT NULL DEFAULT false,
    region             TEXT          NOT NULL DEFAULT 'us-east-1',
    cost_per_month     NUMERIC(10,2) NOT NULL DEFAULT 0,
    idp_status         TEXT          NOT NULL DEFAULT 'configuring'
);
CREATE INDEX IF NOT EXISTS workspace_pro_identity_providers_resource_idx
    ON workspace_pro_identity_providers (resource_id);

-- ── Cloud WAF (security/firewall/cloud-waf) ───────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_pro_cloud_waf_policies (
    resource_id        BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    mode               TEXT          NOT NULL DEFAULT 'detection'
                       CHECK (mode IN ('blocking','detection','learning')),
    ruleset_version    TEXT          NOT NULL DEFAULT 'CRS 4.2.0',
    custom_rules       INTEGER       NOT NULL DEFAULT 0,
    managed_rules      INTEGER       NOT NULL DEFAULT 142,
    protected_domains  TEXT          NOT NULL DEFAULT '',
    requests_blocked   BIGINT        NOT NULL DEFAULT 0,
    requests_allowed   BIGINT        NOT NULL DEFAULT 0,
    threat_score       NUMERIC(6,2)  NOT NULL DEFAULT 0,
    owasp_top10        BOOLEAN       NOT NULL DEFAULT true,
    bot_protection     BOOLEAN       NOT NULL DEFAULT false,
    rate_limiting      BOOLEAN       NOT NULL DEFAULT false,
    rate_limit_rps     INTEGER       NOT NULL DEFAULT 0,
    geo_blocking       TEXT          NOT NULL DEFAULT 'none',
    ip_whitelist_count INTEGER       NOT NULL DEFAULT 0,
    ip_blacklist_count INTEGER       NOT NULL DEFAULT 0,
    region             TEXT          NOT NULL DEFAULT 'us-east-1',
    cost_per_month     NUMERIC(10,2) NOT NULL DEFAULT 0,
    waf_status         TEXT          NOT NULL DEFAULT 'learning'
);
CREATE INDEX IF NOT EXISTS workspace_pro_cloud_waf_policies_resource_idx
    ON workspace_pro_cloud_waf_policies (resource_id);
