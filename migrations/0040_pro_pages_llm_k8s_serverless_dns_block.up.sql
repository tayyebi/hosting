-- AI & Automation: LLM Inference Endpoints
CREATE TABLE IF NOT EXISTS workspace_llm_endpoints (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model            TEXT NOT NULL DEFAULT '',
    model_format     TEXT NOT NULL DEFAULT 'SafeTensors',
    gpu_type         TEXT NOT NULL DEFAULT '',
    gpu_count        INTEGER NOT NULL DEFAULT 1,
    context_window   INTEGER NOT NULL DEFAULT 4096,
    max_tokens       INTEGER NOT NULL DEFAULT 2048,
    quant_bits       INTEGER NOT NULL DEFAULT 16,
    replicas_current INTEGER NOT NULL DEFAULT 0,
    replicas_max     INTEGER NOT NULL DEFAULT 1,
    rpm              INTEGER NOT NULL DEFAULT 0,
    avg_latency_ms   INTEGER NOT NULL DEFAULT 0,
    cost_per_hour    NUMERIC(10,4) NOT NULL DEFAULT 0,
    region           TEXT NOT NULL DEFAULT 'us-east-1',
    scale_to_zero    BOOLEAN NOT NULL DEFAULT true,
    endpoint_status  TEXT NOT NULL DEFAULT 'provisioning'
);

-- Compute: Managed Kubernetes Clusters (Pro Feature)
CREATE TABLE IF NOT EXISTS workspace_managed_k8s_clusters (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    k8s_version       TEXT NOT NULL DEFAULT '1.30.2',
    node_pools        INTEGER NOT NULL DEFAULT 1,
    total_nodes       INTEGER NOT NULL DEFAULT 3,
    vcpus             INTEGER NOT NULL DEFAULT 12,
    memory_gb         INTEGER NOT NULL DEFAULT 48,
    pod_count         INTEGER NOT NULL DEFAULT 0,
    pod_capacity      INTEGER NOT NULL DEFAULT 165,
    cni_plugin        TEXT NOT NULL DEFAULT 'Cilium',
    ingress_ctrl      TEXT NOT NULL DEFAULT 'NGINX',
    service_mesh      BOOLEAN NOT NULL DEFAULT false,
    auto_scale        BOOLEAN NOT NULL DEFAULT true,
    region            TEXT NOT NULL DEFAULT 'us-east-1',
    control_plane_ha  BOOLEAN NOT NULL DEFAULT false,
    cost_per_hour     NUMERIC(10,4) NOT NULL DEFAULT 0,
    cluster_status    TEXT NOT NULL DEFAULT 'provisioning'
);

-- Apps: Serverless Functions (Pro Feature) — extended schema
CREATE TABLE IF NOT EXISTS workspace_pro_serverless_functions (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    runtime          TEXT NOT NULL DEFAULT 'Node.js 20',
    trigger          TEXT NOT NULL DEFAULT 'HTTP GET',
    route            TEXT NOT NULL DEFAULT '/',
    memory_mb        INTEGER NOT NULL DEFAULT 256,
    timeout_sec      INTEGER NOT NULL DEFAULT 10,
    env_vars         INTEGER NOT NULL DEFAULT 0,
    region           TEXT NOT NULL DEFAULT 'us-east-1',
    invocations_24h  INTEGER NOT NULL DEFAULT 0,
    avg_duration_ms  INTEGER NOT NULL DEFAULT 0,
    error_rate       NUMERIC(8,4) NOT NULL DEFAULT 0,
    last_deployed    TEXT NOT NULL DEFAULT '',
    code_size_kb     INTEGER NOT NULL DEFAULT 0,
    cold_start_ms    INTEGER NOT NULL DEFAULT 0,
    func_status      TEXT NOT NULL DEFAULT 'deploying'
);

-- Networking: DNS Zones (Pro Feature) — extended schema
CREATE TABLE IF NOT EXISTS workspace_pro_dns_zones (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    domain           TEXT NOT NULL DEFAULT '',
    zone_type        TEXT NOT NULL DEFAULT 'Primary',
    name_servers     TEXT NOT NULL DEFAULT '',
    record_count     INTEGER NOT NULL DEFAULT 0,
    dnssec           BOOLEAN NOT NULL DEFAULT false,
    dnssec_algo      TEXT NOT NULL DEFAULT '',
    ttl_default      INTEGER NOT NULL DEFAULT 300,
    geo_routing      BOOLEAN NOT NULL DEFAULT false,
    aname_flat       BOOLEAN NOT NULL DEFAULT false,
    queries_day      INTEGER NOT NULL DEFAULT 0,
    propagation_ms   INTEGER NOT NULL DEFAULT 0,
    registrar        TEXT NOT NULL DEFAULT '',
    expires_at       TEXT NOT NULL DEFAULT '',
    auto_renew       BOOLEAN NOT NULL DEFAULT false,
    zone_status      TEXT NOT NULL DEFAULT 'pending_ns'
);

-- Storage: Block Storage Volumes (Pro Feature) — extended schema
CREATE TABLE IF NOT EXISTS workspace_pro_block_volumes (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    size_gb           INTEGER NOT NULL DEFAULT 50,
    volume_type       TEXT NOT NULL DEFAULT 'SSD',
    iops              INTEGER NOT NULL DEFAULT 3000,
    throughput_mbps   INTEGER NOT NULL DEFAULT 125,
    encrypted         BOOLEAN NOT NULL DEFAULT true,
    kms_key           TEXT NOT NULL DEFAULT '',
    snapshot_count    INTEGER NOT NULL DEFAULT 0,
    attached_to       TEXT NOT NULL DEFAULT '',
    region            TEXT NOT NULL DEFAULT 'us-east-1',
    availability_zone TEXT NOT NULL DEFAULT 'a',
    multi_attach      BOOLEAN NOT NULL DEFAULT false,
    vol_status        TEXT NOT NULL DEFAULT 'available'
);
