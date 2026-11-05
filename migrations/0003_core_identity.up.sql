-- 0003_core_identity: tenants, users, memberships, sessions, api keys,
-- oauth tokens, password resets, audit log, user settings sub-tables.
--
-- Partitioned tables (sessions, audit_log, system_logs) use RANGE partitioning
-- by month on (created_at / occurred_at). Initial partitions span
-- 2026-01 .. 2027-12 — a Phase-2 maintenance job will create future ones.

-- ── helper: monthly partition factory ─────────────────────────
CREATE OR REPLACE FUNCTION public.create_monthly_partition(
    parent_table text,
    partition_key_column text,
    start_month date
) RETURNS void AS $$
DECLARE
    partition_name text;
    start_ts text;
    end_ts text;
BEGIN
    partition_name := parent_table || '_' || to_char(start_month, 'YYYY_MM');
    start_ts := to_char(start_month, 'YYYY-MM-DD');
    end_ts := to_char(start_month + INTERVAL '1 month', 'YYYY-MM-DD');
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name, parent_table, start_ts, end_ts
    );
END;
$$ LANGUAGE plpgsql;

-- ── tenants ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenants (
    id          BIGSERIAL PRIMARY KEY,
    slug        CITEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    domain      CITEXT UNIQUE,
    plan        TEXT NOT NULL DEFAULT 'free',
    owner_name  TEXT,
    owner_email CITEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS tenants_active_idx ON tenants (id) WHERE deleted_at IS NULL;

-- ── users (global — one record even if user is in many tenants) ─
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         CITEXT NOT NULL UNIQUE,
    name          TEXT,
    avatar_url    TEXT,
    password_hash TEXT,               -- NULL for SSO-only users
    global_role   TEXT NOT NULL DEFAULT 'user' CHECK (global_role IN ('user','sysadmin')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ
);

-- ── user_tenant_memberships (the many-to-many with per-tenant role) ─
CREATE TABLE IF NOT EXISTS user_tenant_memberships (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role        TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('tenant','user')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, tenant_id)
);
CREATE INDEX IF NOT EXISTS user_tenant_memberships_tenant_idx ON user_tenant_memberships (tenant_id);

-- ── sessions (partitioned by created_at, month) ───────────────
CREATE TABLE IF NOT EXISTS sessions (
    id           UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id    BIGINT REFERENCES tenants(id) ON DELETE CASCADE, -- null for sysadmin
    token_hash   TEXT NOT NULL,
    ip_address   INET,
    user_agent   TEXT,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at   TIMESTAMPTZ,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
CREATE INDEX IF NOT EXISTS sessions_token_hash_idx ON sessions (token_hash);
CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at) WHERE revoked_at IS NULL;

-- ── api_keys (tenant-scoped) ──────────────────────────────────
CREATE TABLE IF NOT EXISTS api_keys (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    key_hash      TEXT NOT NULL UNIQUE,
    created_by    BIGINT REFERENCES users(id),
    last_used_at  TIMESTAMPTZ,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS api_keys_tenant_id_idx ON api_keys (tenant_id) WHERE revoked_at IS NULL;

-- ── oauth_tokens (third-party providers linked to a platform user) ─
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id      BIGINT REFERENCES tenants(id) ON DELETE CASCADE,
    provider       TEXT NOT NULL,
    provider_uid   TEXT NOT NULL,
    access_token   TEXT NOT NULL,
    refresh_token  TEXT,
    scopes         TEXT[] NOT NULL DEFAULT '{}',
    expires_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_uid)
);
CREATE INDEX IF NOT EXISTS oauth_tokens_user_id_idx ON oauth_tokens (user_id);

-- ── password_reset_tokens ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ
);

-- ── audit_log (partitioned by created_at) ─────────────────────
CREATE TABLE IF NOT EXISTS audit_log (
    id           BIGSERIAL,
    tenant_id    BIGINT,                     -- NULL for system/global actions
    actor_user_id BIGINT,                    -- NULL for anonymous / system
    actor_role   TEXT,                       -- 'sysadmin' | 'tenant' | 'user' | 'system'
    action       TEXT NOT NULL,              -- e.g. 'tenant.create', 'sql.adhoc.run'
    entity       TEXT,                       -- e.g. 'tenants', 'monitoring_checks'
    entity_id    TEXT,                       -- stringified (UUID or bigint)
    severity     TEXT NOT NULL DEFAULT 'info' CHECK (severity IN ('info','warn','critical')),
    request_id   TEXT,
    ip_address   INET,
    details      JSONB NOT NULL DEFAULT '{}'::jsonb,
    error        TEXT,
    duration_ms  INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
CREATE INDEX IF NOT EXISTS audit_log_tenant_created_idx ON audit_log (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_log_actor_idx ON audit_log (actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_log_action_idx ON audit_log (action);

-- ── system_logs (sysadmin /s/logs, partitioned by created_at) ─
CREATE TABLE IF NOT EXISTS system_logs (
    id           BIGSERIAL,
    level        TEXT NOT NULL CHECK (level IN ('debug','info','warn','error','fatal')),
    service      TEXT NOT NULL,
    message      TEXT NOT NULL,
    fields       JSONB NOT NULL DEFAULT '{}'::jsonb,
    trace_id     TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
CREATE INDEX IF NOT EXISTS system_logs_level_created_idx ON system_logs (level, created_at DESC);
CREATE INDEX IF NOT EXISTS system_logs_service_created_idx ON system_logs (service, created_at DESC);

-- ── user_preferences (1-1 with users) ─────────────────────────
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id             BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme               TEXT NOT NULL DEFAULT 'system' CHECK (theme IN ('light','dark','system')),
    timezone            TEXT NOT NULL DEFAULT 'UTC',
    locale              TEXT NOT NULL DEFAULT 'en-US',
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── user_security_2fa (1-1 with users) ────────────────────────
CREATE TABLE IF NOT EXISTS user_security_2fa (
    user_id        BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    enabled        BOOLEAN NOT NULL DEFAULT false,
    method         TEXT CHECK (method IN ('totp','webauthn','sms')),
    secret_hash    TEXT,
    recovery_codes TEXT[],
    last_used_at   TIMESTAMPTZ,
    enrolled_at    TIMESTAMPTZ
);

-- ── user_personal_api_keys (user-scoped, distinct from tenant api_keys) ─
CREATE TABLE IF NOT EXISTS user_personal_api_keys (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    key_hash     TEXT NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at   TIMESTAMPTZ
);

-- ── user_workspace_invitations ────────────────────────────────
CREATE TABLE IF NOT EXISTS user_workspace_invitations (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    inviter_id     BIGINT NOT NULL REFERENCES users(id),
    invitee_email  CITEXT NOT NULL,
    role           TEXT NOT NULL DEFAULT 'user',
    token_hash     TEXT NOT NULL UNIQUE,
    accepted_at    TIMESTAMPTZ,
    expires_at     TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS user_workspace_invitations_email_idx ON user_workspace_invitations (invitee_email);

-- ── create initial monthly partitions (2026-01 .. 2027-12) ────
DO $$
DECLARE
    m date;
BEGIN
    FOR m IN
        SELECT generate_series(date '2026-01-01', date '2027-12-01', interval '1 month')::date
    LOOP
        PERFORM create_monthly_partition('sessions', 'created_at', m);
        PERFORM create_monthly_partition('audit_log', 'created_at', m);
        PERFORM create_monthly_partition('system_logs', 'created_at', m);
    END LOOP;
END $$;
