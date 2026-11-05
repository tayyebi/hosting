-- 0027: workspace resources — networking SD-WAN and AI training/hosting pages.
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── SD-WAN Networks ────────────────────────────────────────────────────────────
-- Covers networking/private/sdwan
-- topology: 'hub-and-spoke' | 'full-mesh' | 'partial-mesh'
CREATE TABLE IF NOT EXISTS workspace_sdwan_networks (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    topology     TEXT NOT NULL DEFAULT 'hub-and-spoke',
    sites        INTEGER NOT NULL DEFAULT 0,
    hubs         INTEGER NOT NULL DEFAULT 0,
    bandwidth    TEXT NOT NULL DEFAULT '0 Mbps',
    protocol     TEXT NOT NULL DEFAULT 'IPsec + VXLAN',
    encryption   TEXT NOT NULL DEFAULT 'AES-256-GCM',
    qos          BOOLEAN NOT NULL DEFAULT false,
    uptime       TEXT NOT NULL DEFAULT '—'
);

CREATE INDEX IF NOT EXISTS workspace_sdwan_networks_resource_idx
    ON workspace_sdwan_networks (resource_id);

-- ── AI Distributed Training Jobs ───────────────────────────────────────────────
-- Covers ai/training/distributed
-- framework: 'PyTorch + DeepSpeed' | 'PyTorch + FSDP' | 'JAX + Flax' | 'TensorFlow'
CREATE TABLE IF NOT EXISTS workspace_ai_distributed_training (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    framework     TEXT NOT NULL DEFAULT 'PyTorch',
    nodes         INTEGER NOT NULL DEFAULT 1,
    gpus_per_node INTEGER NOT NULL DEFAULT 1,
    gpu_type      TEXT NOT NULL DEFAULT 'A100 40GB',
    dataset       TEXT NOT NULL DEFAULT '',
    epochs        INTEGER NOT NULL DEFAULT 10,
    progress      TEXT NOT NULL DEFAULT '0%',
    runtime       TEXT NOT NULL DEFAULT '—',
    cost          TEXT NOT NULL DEFAULT '$0'
);

CREATE INDEX IF NOT EXISTS workspace_ai_distributed_training_resource_idx
    ON workspace_ai_distributed_training (resource_id);

-- ── AI AutoML Runs ─────────────────────────────────────────────────────────────
-- Covers ai/training/automl
-- task_type: 'classification' | 'regression' | 'time-series' | 'image-classification' | 'text-classification' | 'anomaly-detection'
CREATE TABLE IF NOT EXISTS workspace_ai_automl_runs (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    task_type    TEXT NOT NULL DEFAULT 'classification',
    dataset      TEXT NOT NULL DEFAULT '',
    target_col   TEXT NOT NULL DEFAULT '',
    trials       INTEGER NOT NULL DEFAULT 100,
    best_model   TEXT NOT NULL DEFAULT '—',
    best_score   TEXT NOT NULL DEFAULT '—',
    metric       TEXT NOT NULL DEFAULT 'AUC',
    runtime      TEXT NOT NULL DEFAULT '—',
    cost         TEXT NOT NULL DEFAULT '$0'
);

CREATE INDEX IF NOT EXISTS workspace_ai_automl_runs_resource_idx
    ON workspace_ai_automl_runs (resource_id);

-- ── AI Fine-tuning Jobs ────────────────────────────────────────────────────────
-- Covers ai/training/finetuning
-- method: 'full' | 'LoRA' | 'QLoRA' | 'Adapter'
CREATE TABLE IF NOT EXISTS workspace_ai_finetuning_jobs (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    base_model    TEXT NOT NULL DEFAULT '',
    method        TEXT NOT NULL DEFAULT 'LoRA',
    dataset       TEXT NOT NULL DEFAULT '',
    epochs        INTEGER NOT NULL DEFAULT 3,
    learning_rate TEXT NOT NULL DEFAULT '2e-4',
    progress      TEXT NOT NULL DEFAULT '0%',
    runtime       TEXT NOT NULL DEFAULT '—',
    cost          TEXT NOT NULL DEFAULT '$0',
    output_model  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_ai_finetuning_jobs_resource_idx
    ON workspace_ai_finetuning_jobs (resource_id);

-- ── AI Inference Endpoints ─────────────────────────────────────────────────────
-- Covers ai/hosting/inference
-- framework: 'vLLM' | 'TorchServe' | 'Triton' | 'TFServing'
CREATE TABLE IF NOT EXISTS workspace_ai_inference_endpoints (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    model         TEXT NOT NULL DEFAULT '',
    framework     TEXT NOT NULL DEFAULT 'vLLM',
    instance_type TEXT NOT NULL DEFAULT 'gpu.t4.1x',
    replicas      INTEGER NOT NULL DEFAULT 1,
    latency       TEXT NOT NULL DEFAULT '—',
    rps           TEXT NOT NULL DEFAULT '0',
    endpoint      TEXT NOT NULL DEFAULT '',
    auto_scale    BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS workspace_ai_inference_endpoints_resource_idx
    ON workspace_ai_inference_endpoints (resource_id);
