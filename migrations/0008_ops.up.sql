-- 0008_ops: operations-facing tables — monitoring, IDS, incidents,
-- status page, maintenance windows, background jobs.

-- ── monitoring_checks ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS monitoring_checks (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    kind        TEXT NOT NULL CHECK (kind IN ('http','https','tcp','ping','dns','ssl')),
    endpoint    TEXT NOT NULL,
    interval_s  INTEGER NOT NULL DEFAULT 60 CHECK (interval_s BETWEEN 10 AND 3600),
    timeout_ms  INTEGER NOT NULL DEFAULT 5000 CHECK (timeout_ms BETWEEN 100 AND 60000),
    region_id   BIGINT REFERENCES regions(id),
    enabled     BOOLEAN NOT NULL DEFAULT true,
    last_status TEXT,
    last_run_at TIMESTAMPTZ,
    created_by  BIGINT REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (tenant_id, name)
);
CREATE INDEX IF NOT EXISTS monitoring_checks_tenant_idx ON monitoring_checks (tenant_id)
    WHERE deleted_at IS NULL;

-- ── monitoring_results (partitioned by observed_at) ───────────
CREATE TABLE IF NOT EXISTS monitoring_results (
    id            BIGSERIAL,
    tenant_id     BIGINT NOT NULL,
    check_id      BIGINT NOT NULL,
    status        TEXT NOT NULL CHECK (status IN ('ok','degraded','down','timeout','error')),
    latency_ms    INTEGER,
    status_code   INTEGER,
    error_message TEXT,
    observed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, observed_at)
) PARTITION BY RANGE (observed_at);
CREATE INDEX IF NOT EXISTS monitoring_results_check_observed_idx
    ON monitoring_results (check_id, observed_at DESC);
CREATE INDEX IF NOT EXISTS monitoring_results_tenant_observed_idx
    ON monitoring_results (tenant_id, observed_at DESC);

-- ── ids_events (partitioned by occurred_at) ───────────────────
CREATE TABLE IF NOT EXISTS ids_events (
    id            BIGSERIAL,
    tenant_id     BIGINT NOT NULL,
    event_type    TEXT NOT NULL,
    severity      TEXT NOT NULL CHECK (severity IN ('low','medium','high','critical')),
    source_ip     INET,
    destination_ip INET,
    rule_id       TEXT,
    summary       TEXT,
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, occurred_at)
) PARTITION BY RANGE (occurred_at);
CREATE INDEX IF NOT EXISTS ids_events_tenant_occurred_idx
    ON ids_events (tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS ids_events_severity_idx
    ON ids_events (severity, occurred_at DESC);

-- ── incidents ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS incidents (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title            TEXT NOT NULL,
    severity         TEXT NOT NULL CHECK (severity IN ('sev1','sev2','sev3','sev4')),
    status           TEXT NOT NULL DEFAULT 'open'
                     CHECK (status IN ('open','investigating','identified','monitoring','resolved','postmortem')),
    commander_id     BIGINT REFERENCES users(id),
    summary          TEXT,
    started_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    acknowledged_at  TIMESTAMPTZ,
    resolved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS incidents_tenant_started_idx ON incidents (tenant_id, started_at DESC);

CREATE TABLE IF NOT EXISTS incident_timeline_entries (
    id          BIGSERIAL PRIMARY KEY,
    incident_id BIGINT NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    author_id   BIGINT REFERENCES users(id),
    kind        TEXT NOT NULL CHECK (kind IN ('note','status_change','mitigation','communication')),
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS incident_timeline_incident_idx ON incident_timeline_entries (incident_id, created_at);

CREATE TABLE IF NOT EXISTS incident_affected_services (
    id          BIGSERIAL PRIMARY KEY,
    incident_id BIGINT NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service     TEXT NOT NULL,
    impact      TEXT NOT NULL CHECK (impact IN ('degraded','partial_outage','major_outage'))
);
CREATE INDEX IF NOT EXISTS incident_affected_services_incident_idx ON incident_affected_services (incident_id);

-- ── status_page_components ────────────────────────────────────
CREATE TABLE IF NOT EXISTS status_page_components (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    group_name   TEXT,
    name         TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'operational'
                 CHECK (status IN ('operational','degraded','partial_outage','major_outage','under_maintenance')),
    description  TEXT,
    position     INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);

-- ── maintenance_windows ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    kind            TEXT NOT NULL CHECK (kind IN ('planned','emergency')),
    scheduled_start TIMESTAMPTZ NOT NULL,
    scheduled_end   TIMESTAMPTZ NOT NULL,
    actual_start    TIMESTAMPTZ,
    actual_end      TIMESTAMPTZ,
    description     TEXT,
    created_by      BIGINT REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (scheduled_end > scheduled_start)
);
CREATE INDEX IF NOT EXISTS maintenance_windows_tenant_scheduled_idx
    ON maintenance_windows (tenant_id, scheduled_start);

CREATE TABLE IF NOT EXISTS maintenance_affected_services (
    id                      BIGSERIAL PRIMARY KEY,
    maintenance_window_id   BIGINT NOT NULL REFERENCES maintenance_windows(id) ON DELETE CASCADE,
    tenant_id               BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service                 TEXT NOT NULL
);

-- ── background_jobs (partitioned by scheduled_at) ─────────────
CREATE TABLE IF NOT EXISTS background_jobs (
    id            BIGSERIAL,
    tenant_id     BIGINT,                                 -- NULL for platform-wide jobs
    queue         TEXT NOT NULL,
    kind          TEXT NOT NULL,                          -- 'monitoring.run', 'email.send', ...
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    status        TEXT NOT NULL DEFAULT 'queued'
                  CHECK (status IN ('queued','running','succeeded','failed','cancelled')),
    priority      SMALLINT NOT NULL DEFAULT 5,
    attempts      INTEGER NOT NULL DEFAULT 0,
    max_attempts  INTEGER NOT NULL DEFAULT 5,
    last_error    TEXT,
    scheduled_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at    TIMESTAMPTZ,
    finished_at   TIMESTAMPTZ,
    worker_id     BIGINT,
    PRIMARY KEY (id, scheduled_at)
) PARTITION BY RANGE (scheduled_at);
CREATE INDEX IF NOT EXISTS background_jobs_queue_status_idx
    ON background_jobs (queue, status, scheduled_at)
    WHERE status IN ('queued','running');
CREATE INDEX IF NOT EXISTS background_jobs_tenant_status_idx
    ON background_jobs (tenant_id, status, scheduled_at DESC);

CREATE TABLE IF NOT EXISTS background_job_queues (
    name           TEXT PRIMARY KEY,
    description    TEXT,
    concurrency    INTEGER NOT NULL DEFAULT 4,
    paused         BOOLEAN NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS background_job_workers (
    id             BIGSERIAL PRIMARY KEY,
    hostname       TEXT NOT NULL,
    pid            INTEGER,
    queues         TEXT[] NOT NULL DEFAULT '{}',
    status         TEXT NOT NULL DEFAULT 'idle'
                   CHECK (status IN ('idle','busy','stopping','stopped')),
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── create initial monthly partitions (2026-01 .. 2027-12) ────
DO $$
DECLARE
    m date;
BEGIN
    FOR m IN
        SELECT generate_series(date '2026-01-01', date '2027-12-01', interval '1 month')::date
    LOOP
        PERFORM create_monthly_partition('monitoring_results', 'observed_at', m);
        PERFORM create_monthly_partition('ids_events',         'occurred_at', m);
        PERFORM create_monthly_partition('background_jobs',    'scheduled_at', m);
    END LOOP;
END $$;

-- ── RLS on the tenant-scoped ops tables ───────────────────────
DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'monitoring_checks', 'incidents', 'incident_timeline_entries',
        'incident_affected_services', 'status_page_components',
        'maintenance_windows', 'maintenance_affected_services'
    ] LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON %I
                USING (tenant_id = public.current_tenant_id()
                       OR public.current_actor_role() = ''sysadmin'')
                WITH CHECK (tenant_id = public.current_tenant_id()
                       OR public.current_actor_role() = ''sysadmin'')', t);
    END LOOP;
END $$;

-- monitoring_results + ids_events are partitioned; RLS enabled on the parent
-- propagates to all partitions.
ALTER TABLE monitoring_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE monitoring_results FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON monitoring_results;
CREATE POLICY tenant_isolation ON monitoring_results
    USING (tenant_id = public.current_tenant_id()
           OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id()
           OR public.current_actor_role() = 'sysadmin');

ALTER TABLE ids_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE ids_events FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON ids_events;
CREATE POLICY tenant_isolation ON ids_events
    USING (tenant_id = public.current_tenant_id()
           OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id()
           OR public.current_actor_role() = 'sysadmin');

-- background_jobs: tenant jobs visible to their tenant; system jobs to sysadmin.
ALTER TABLE background_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE background_jobs FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS jobs_isolation ON background_jobs;
CREATE POLICY jobs_isolation ON background_jobs
    USING (
        tenant_id = public.current_tenant_id()
        OR (tenant_id IS NULL AND public.current_actor_role() IN ('sysadmin','system'))
    )
    WITH CHECK (
        tenant_id = public.current_tenant_id()
        OR public.current_actor_role() IN ('sysadmin','system')
    );
