-- 0010_sso: SSO identity provider configuration, registered OAuth clients,
-- and per-session CORS origins for the platform's built-in IdP.

-- ── sso_config (single-row, seeded on first boot) ─────────────
CREATE TABLE sso_config (
    id                  BIGSERIAL PRIMARY KEY,
    enabled             BOOLEAN NOT NULL DEFAULT true,
    provider_name       TEXT NOT NULL DEFAULT 'Platform Identity',
    issuer_url          TEXT NOT NULL DEFAULT '',
    session_lifetime_h  INTEGER NOT NULL DEFAULT 8,
    oauth2_enabled      BOOLEAN NOT NULL DEFAULT true,
    oidc_enabled        BOOLEAN NOT NULL DEFAULT true,
    saml_enabled        BOOLEAN NOT NULL DEFAULT false,
    -- OAuth2 grant types
    grant_auth_code     BOOLEAN NOT NULL DEFAULT true,
    grant_pkce          BOOLEAN NOT NULL DEFAULT true,
    grant_client_creds  BOOLEAN NOT NULL DEFAULT true,
    grant_refresh_token BOOLEAN NOT NULL DEFAULT true,
    grant_device_code   BOOLEAN NOT NULL DEFAULT false,
    access_token_ttl_h  INTEGER NOT NULL DEFAULT 1,
    refresh_token_ttl_d INTEGER NOT NULL DEFAULT 30,
    -- OIDC
    oidc_signing_algo   TEXT NOT NULL DEFAULT 'RS256',
    id_token_ttl_h      INTEGER NOT NULL DEFAULT 1,
    claim_email         BOOLEAN NOT NULL DEFAULT true,
    claim_roles         BOOLEAN NOT NULL DEFAULT true,
    claim_tenant        BOOLEAN NOT NULL DEFAULT false,
    -- SAML
    saml_signing_algo   TEXT NOT NULL DEFAULT 'RSA-SHA256',
    saml_sign_assert    BOOLEAN NOT NULL DEFAULT true,
    saml_enc_assert     BOOLEAN NOT NULL DEFAULT false,
    -- Security
    brute_force_enabled BOOLEAN NOT NULL DEFAULT true,
    max_failed_attempts INTEGER NOT NULL DEFAULT 5,
    lockout_duration_m  INTEGER NOT NULL DEFAULT 15,
    notify_on_lockout   BOOLEAN NOT NULL DEFAULT true,
    ip_restrict         BOOLEAN NOT NULL DEFAULT false,
    allowed_cidrs       TEXT NOT NULL DEFAULT '',
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── sso_oauth_clients ─────────────────────────────────────────
CREATE TABLE sso_oauth_clients (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    client_id   TEXT NOT NULL UNIQUE DEFAULT 'client_' || substr(gen_random_uuid()::text, 1, 6),
    protocol    TEXT NOT NULL DEFAULT 'OAuth 2.0 / OIDC',
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── sso_cors_origins ─────────────────────────────────────────
CREATE TABLE sso_cors_origins (
    id         BIGSERIAL PRIMARY KEY,
    origin     TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
