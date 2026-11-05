-- 0004_workspaces_and_resources: regions (global), workspaces,
-- workspace_memberships, resources (CTI base), replicas.

-- ── regions (global seed; referenced by resources.region_id) ──
CREATE TABLE IF NOT EXISTS regions (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,         -- 'us-east-1'
    name        TEXT NOT NULL,                -- 'US East (N. Virginia)'
    country     TEXT,                         -- ISO 3166-1 alpha-2
    geo_area    TEXT,                         -- 'north-america', 'europe', ...
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── workspaces (tenant-scoped container for resources) ────────
CREATE TABLE IF NOT EXISTS workspaces (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    slug        TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT,
    created_by  BIGINT REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (tenant_id, slug)
);
CREATE INDEX IF NOT EXISTS workspaces_tenant_idx ON workspaces (tenant_id) WHERE deleted_at IS NULL;

-- ── workspace_memberships ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS workspace_memberships (
    id           BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id    BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role         TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner','admin','member','viewer')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, user_id)
);
CREATE INDEX IF NOT EXISTS workspace_memberships_user_idx ON workspace_memberships (user_id);
CREATE INDEX IF NOT EXISTS workspace_memberships_tenant_idx ON workspace_memberships (tenant_id);

-- ── resources: CTI base table for every workspace-level resource ──
-- Every specialized table (compute_k8s_clusters, storage_block_volumes, …)
-- has a BIGINT resource_id PRIMARY KEY → resources.id (ON DELETE CASCADE).
-- The `kind` column is the discriminator used when joining back to the parent.
CREATE TABLE IF NOT EXISTS resources (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    kind         TEXT NOT NULL,
    name         TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','provisioning','active','degraded',
                                   'stopped','failed','deleting','deleted')),
    region_id    BIGINT REFERENCES regions(id),
    tags         JSONB NOT NULL DEFAULT '{}'::jsonb,     -- user-facing k/v only; NOT a spec escape hatch
    created_by   BIGINT REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE (workspace_id, kind, name)
);
CREATE INDEX IF NOT EXISTS resources_tenant_workspace_idx ON resources (tenant_id, workspace_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS resources_kind_idx ON resources (kind) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS resources_status_idx ON resources (status) WHERE deleted_at IS NULL;

-- NOTE: the previous `replicas` table definition (tenant-scoped, keyed by
-- merchant_code) was never referenced by any handler and has been removed.
-- The canonical platform-wide `replicas` registry lives in 0009.

-- ── seed regions (sysadmin can extend; keeping a global fixed list) ──
INSERT INTO regions (slug, name, country, geo_area) VALUES
    ('us-east-1',      'US East (N. Virginia)',   'US', 'north-america'),
    ('us-west-2',      'US West (Oregon)',        'US', 'north-america'),
    ('eu-west-1',      'EU West (Ireland)',       'IE', 'europe'),
    ('eu-central-1',   'EU Central (Frankfurt)',  'DE', 'europe'),
    ('ap-southeast-1', 'AP Southeast (Singapore)','SG', 'asia-pacific'),
    ('ap-northeast-1', 'AP Northeast (Tokyo)',    'JP', 'asia-pacific')
ON CONFLICT (slug) DO NOTHING;
