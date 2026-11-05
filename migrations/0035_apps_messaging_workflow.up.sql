-- 0035: workspace resources — Application Services: Messaging (Queues, Streaming, ServiceBus) + Workflow (Automation, Scheduler)
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Messaging: Message Queues ─────────────────────────────────────────────────
-- Covers apps/messaging/queues
CREATE TABLE IF NOT EXISTS workspace_apps_messaging_queues (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine       TEXT    NOT NULL DEFAULT '',
    queue_type   TEXT    NOT NULL DEFAULT 'Work queue',
    messages     TEXT    NOT NULL DEFAULT '0 pending',
    throughput   TEXT    NOT NULL DEFAULT '0 msg/s',
    consumers    INTEGER NOT NULL DEFAULT 0,
    retention    TEXT    NOT NULL DEFAULT '7 days',
    queue_status TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_apps_messaging_queues_resource_idx ON workspace_apps_messaging_queues (resource_id);

-- ── Messaging: Event Streaming ────────────────────────────────────────────────
-- Covers apps/messaging/streaming
CREATE TABLE IF NOT EXISTS workspace_apps_messaging_streaming (
    resource_id      BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    engine           TEXT    NOT NULL DEFAULT '',
    partitions       INTEGER NOT NULL DEFAULT 0,
    throughput       TEXT    NOT NULL DEFAULT '0 /s',
    consumer_groups  INTEGER NOT NULL DEFAULT 0,
    lag              TEXT    NOT NULL DEFAULT '—',
    retention        TEXT    NOT NULL DEFAULT '7 days',
    stream_status    TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_apps_messaging_streaming_resource_idx ON workspace_apps_messaging_streaming (resource_id);

-- ── Messaging: Service Bus ────────────────────────────────────────────────────
-- Covers apps/messaging/servicebus
CREATE TABLE IF NOT EXISTS workspace_apps_messaging_servicebus (
    resource_id   BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    tier          TEXT    NOT NULL DEFAULT 'Standard',
    subscriptions INTEGER NOT NULL DEFAULT 0,
    throughput    TEXT    NOT NULL DEFAULT '0 /s',
    dead_letter   INTEGER NOT NULL DEFAULT 0,
    avg_delivery  TEXT    NOT NULL DEFAULT '—',
    msg_size      TEXT    NOT NULL DEFAULT '—',
    bus_status    TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_apps_messaging_servicebus_resource_idx ON workspace_apps_messaging_servicebus (resource_id);

-- ── Workflow: Automation ──────────────────────────────────────────────────────
-- Covers apps/workflow/automation
CREATE TABLE IF NOT EXISTS workspace_apps_workflow_automation (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    trigger      TEXT    NOT NULL DEFAULT '',
    actions      INTEGER NOT NULL DEFAULT 0,
    runs         TEXT    NOT NULL DEFAULT '0',
    success_rate TEXT    NOT NULL DEFAULT '—',
    last_run     TEXT    NOT NULL DEFAULT '—',
    wf_status    TEXT    NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_apps_workflow_automation_resource_idx ON workspace_apps_workflow_automation (resource_id);

-- ── Workflow: Scheduler / Cron Jobs ──────────────────────────────────────────
-- Covers apps/workflow/scheduler
CREATE TABLE IF NOT EXISTS workspace_apps_workflow_scheduler (
    resource_id BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    schedule    TEXT NOT NULL DEFAULT '',
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    last_run    TEXT NOT NULL DEFAULT '—',
    duration    TEXT NOT NULL DEFAULT '—',
    next_run    TEXT NOT NULL DEFAULT '—',
    streak      TEXT NOT NULL DEFAULT '—',
    job_status  TEXT NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS workspace_apps_workflow_scheduler_resource_idx ON workspace_apps_workflow_scheduler (resource_id);
