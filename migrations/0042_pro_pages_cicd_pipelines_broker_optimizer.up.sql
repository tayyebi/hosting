-- 0042: Pro feature pages — CI/CD Pipelines, Message Broker, Cloud Optimizer.
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── CI/CD Pipelines (devops/cicd/pipelines) ───────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_pro_cicd_pipelines (
    resource_id        BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    repository         TEXT          NOT NULL DEFAULT '',
    branch             TEXT          NOT NULL DEFAULT 'main',
    trigger            TEXT          NOT NULL DEFAULT 'push'
                       CHECK (trigger IN ('push','merge_request','tag','schedule','manual')),
    build_tool         TEXT          NOT NULL DEFAULT 'Docker',
    runtime            TEXT          NOT NULL DEFAULT 'Node 20',
    stage_count        INTEGER       NOT NULL DEFAULT 3,
    last_run_status    TEXT          NOT NULL DEFAULT 'pending'
                       CHECK (last_run_status IN ('success','failed','running','cancelled','pending')),
    last_run_duration  INTEGER       NOT NULL DEFAULT 0,
    total_runs         INTEGER       NOT NULL DEFAULT 0,
    success_rate       NUMERIC(6,2)  NOT NULL DEFAULT 0,
    avg_duration       INTEGER       NOT NULL DEFAULT 0,
    artifacts          BOOLEAN       NOT NULL DEFAULT true,
    cache_enabled      BOOLEAN       NOT NULL DEFAULT true,
    auto_deploy        BOOLEAN       NOT NULL DEFAULT false,
    deploy_target      TEXT          NOT NULL DEFAULT '',
    region             TEXT          NOT NULL DEFAULT 'us-east-1',
    cost_per_month     NUMERIC(10,2) NOT NULL DEFAULT 0,
    pipeline_status    TEXT          NOT NULL DEFAULT 'active'
                       CHECK (pipeline_status IN ('active','paused','failing'))
);
CREATE INDEX IF NOT EXISTS workspace_pro_cicd_pipelines_resource_idx
    ON workspace_pro_cicd_pipelines (resource_id);

-- ── Message Broker (apps/queues/message-broker) ───────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_pro_message_brokers (
    resource_id        BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    protocol           TEXT          NOT NULL DEFAULT 'AMQP 0.9.1',
    engine             TEXT          NOT NULL DEFAULT 'RabbitMQ',
    version            TEXT          NOT NULL DEFAULT '3.13',
    queue_count        INTEGER       NOT NULL DEFAULT 0,
    exchange_count     INTEGER       NOT NULL DEFAULT 0,
    consumer_count     INTEGER       NOT NULL DEFAULT 0,
    messages_ready     BIGINT        NOT NULL DEFAULT 0,
    messages_unacked   BIGINT        NOT NULL DEFAULT 0,
    publish_rate       INTEGER       NOT NULL DEFAULT 0,
    consume_rate       INTEGER       NOT NULL DEFAULT 0,
    storage_gb         NUMERIC(10,2) NOT NULL DEFAULT 0,
    retention_hours    INTEGER       NOT NULL DEFAULT 72,
    tls_enabled        BOOLEAN       NOT NULL DEFAULT true,
    ha_mode            BOOLEAN       NOT NULL DEFAULT false,
    cluster_nodes      INTEGER       NOT NULL DEFAULT 1,
    vhost_count        INTEGER       NOT NULL DEFAULT 1,
    region             TEXT          NOT NULL DEFAULT 'us-east-1',
    cost_per_month     NUMERIC(10,2) NOT NULL DEFAULT 0,
    broker_status      TEXT          NOT NULL DEFAULT 'provisioning'
                       CHECK (broker_status IN ('running','degraded','stopped','provisioning'))
);
CREATE INDEX IF NOT EXISTS workspace_pro_message_brokers_resource_idx
    ON workspace_pro_message_brokers (resource_id);

-- ── Cloud Optimizer (multicloud/cost/cloud-optimizer) ─────────────────────────
CREATE TABLE IF NOT EXISTS workspace_pro_cost_optimizer_accounts (
    resource_id              BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    provider                 TEXT          NOT NULL DEFAULT 'Multi-Cloud',
    account_id               TEXT          NOT NULL DEFAULT '',
    monthly_spend            NUMERIC(12,2) NOT NULL DEFAULT 0,
    optimized_spend          NUMERIC(12,2) NOT NULL DEFAULT 0,
    savings_percent          NUMERIC(6,2)  NOT NULL DEFAULT 0,
    recommendation_count     INTEGER       NOT NULL DEFAULT 0,
    implemented_count        INTEGER       NOT NULL DEFAULT 0,
    rightsizing_alerts       INTEGER       NOT NULL DEFAULT 0,
    unused_resources         INTEGER       NOT NULL DEFAULT 0,
    reserved_coverage        NUMERIC(6,2)  NOT NULL DEFAULT 0,
    spot_usage               NUMERIC(6,2)  NOT NULL DEFAULT 0,
    budget_limit             NUMERIC(12,2) NOT NULL DEFAULT 0,
    budget_used              NUMERIC(12,2) NOT NULL DEFAULT 0,
    anomaly_threshold        NUMERIC(6,2)  NOT NULL DEFAULT 20,
    auto_remediation         BOOLEAN       NOT NULL DEFAULT false,
    tag_compliance           NUMERIC(6,2)  NOT NULL DEFAULT 0,
    region                   TEXT          NOT NULL DEFAULT 'multi-region',
    cost_per_month           NUMERIC(10,2) NOT NULL DEFAULT 0,
    optimizer_status         TEXT          NOT NULL DEFAULT 'active'
                             CHECK (optimizer_status IN ('active','scanning','paused'))
);
CREATE INDEX IF NOT EXISTS workspace_pro_cost_optimizer_accounts_resource_idx
    ON workspace_pro_cost_optimizer_accounts (resource_id);
