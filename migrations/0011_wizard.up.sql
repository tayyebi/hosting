-- 0011_wizard: Track one-click app deployments initiated from the
-- /wizard marketplace. Each deploy creates a background_job.

CREATE TABLE wizard_deploys (
    id           BIGSERIAL PRIMARY KEY,
    app_slug     TEXT NOT NULL,
    app_name     TEXT NOT NULL,
    tenant_id    BIGINT REFERENCES tenants(id) ON DELETE SET NULL,
    workspace_id BIGINT REFERENCES workspaces(id) ON DELETE SET NULL,
    user_id      BIGINT REFERENCES users(id) ON DELETE SET NULL,
    job_id       BIGINT,             -- FK into background_jobs.id (omit FK due to partition)
    status       TEXT NOT NULL DEFAULT 'queued'
                 CHECK (status IN ('queued','provisioning','active','failed','cancelled')),
    region       TEXT NOT NULL DEFAULT '',
    version      TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX wizard_deploys_slug_idx ON wizard_deploys (app_slug);
CREATE INDEX wizard_deploys_tenant_idx ON wizard_deploys (tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX wizard_deploys_user_idx ON wizard_deploys (user_id) WHERE user_id IS NOT NULL;
