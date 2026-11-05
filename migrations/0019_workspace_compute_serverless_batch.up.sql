-- 0019: workspace_compute_serverless_batch — static site hosting, serverless functions,
-- edge functions, batch job scheduling, and HPC/grid computing.

-- ── Static site hosting ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_hosting_static (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    framework        TEXT NOT NULL DEFAULT '',
    branch           TEXT NOT NULL DEFAULT 'main',
    last_deploy_at   TIMESTAMPTZ,
    build_time_sec   INTEGER NOT NULL DEFAULT 0,
    cdn_pops         INTEGER NOT NULL DEFAULT 0,
    bandwidth_gb_mtd NUMERIC(10,2) NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_hosting_static_resource_idx
    ON workspace_hosting_static (resource_id);

-- ── Serverless functions ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_serverless_functions (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    runtime           TEXT NOT NULL DEFAULT '',
    trigger_type      TEXT NOT NULL DEFAULT '',
    memory_mb         INTEGER NOT NULL DEFAULT 128,
    timeout_sec       INTEGER NOT NULL DEFAULT 30,
    invocations_1d    INTEGER NOT NULL DEFAULT 0,
    errors_1d         INTEGER NOT NULL DEFAULT 0,
    avg_duration_ms   INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_serverless_functions_resource_idx
    ON workspace_serverless_functions (resource_id);

-- ── Edge functions ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_serverless_edge (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    path_pattern      TEXT NOT NULL DEFAULT '',
    pop_count         INTEGER NOT NULL DEFAULT 0,
    requests_per_sec  INTEGER NOT NULL DEFAULT 0,
    p99_latency_ms    INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_serverless_edge_resource_idx
    ON workspace_serverless_edge (resource_id);

-- ── Batch job scheduling ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_batch_jobs (
    resource_id       BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    schedule          TEXT NOT NULL DEFAULT '',
    last_run_at       TIMESTAMPTZ,
    last_duration_sec INTEGER NOT NULL DEFAULT 0,
    next_run_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS workspace_batch_jobs_resource_idx
    ON workspace_batch_jobs (resource_id);

-- ── HPC / grid computing ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_batch_hpc (
    resource_id        BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    user_name          TEXT NOT NULL DEFAULT '',
    framework          TEXT NOT NULL DEFAULT '',
    cores              INTEGER NOT NULL DEFAULT 0,
    node_count         INTEGER NOT NULL DEFAULT 0,
    queue_name         TEXT NOT NULL DEFAULT 'normal',
    elapsed_sec        INTEGER NOT NULL DEFAULT 0,
    walltime_sec       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_batch_hpc_resource_idx
    ON workspace_batch_hpc (resource_id);
