-- 0047: third-party sign-in provider availability is configured per tenant.
-- The platform's first-party identity provider remains in the system-level
-- sso_config table from migration 0010.

CREATE TABLE IF NOT EXISTS tenant_sso_providers (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_key TEXT NOT NULL CHECK (provider_key IN (
        'google', 'microsoft', 'github', 'okta', 'custom_oidc', 'custom_saml'
    )),
    enabled      BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, provider_key)
);

CREATE INDEX IF NOT EXISTS tenant_sso_providers_enabled_idx
    ON tenant_sso_providers (tenant_id)
    WHERE enabled = true;

ALTER TABLE tenant_sso_providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_sso_providers FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_sso_providers;
CREATE POLICY tenant_isolation ON tenant_sso_providers
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');
