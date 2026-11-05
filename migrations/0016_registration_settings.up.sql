-- 0016: tenant_registration_settings

CREATE TABLE IF NOT EXISTS tenant_registration_settings (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    enabled             BOOLEAN NOT NULL DEFAULT true,
    redirect_after      TEXT NOT NULL DEFAULT '/dashboard',
    default_role        TEXT NOT NULL DEFAULT 'Client',
    access_mode         TEXT NOT NULL DEFAULT 'open' CHECK (access_mode IN ('open','invite','approval')),
    domain_restrict     BOOLEAN NOT NULL DEFAULT false,
    allowed_domains     JSONB NOT NULL DEFAULT '[]',
    block_disposable    BOOLEAN NOT NULL DEFAULT true,
    max_per_hour_per_ip INTEGER NOT NULL DEFAULT 5,
    max_per_day         INTEGER NOT NULL DEFAULT 200,
    geo_restrict        BOOLEAN NOT NULL DEFAULT false,
    email_required      BOOLEAN NOT NULL DEFAULT true,
    link_expiry         TEXT NOT NULL DEFAULT '24 hours',
    captcha_enabled     BOOLEAN NOT NULL DEFAULT true,
    captcha_provider    TEXT NOT NULL DEFAULT 'hCaptcha',
    min_password_len    INTEGER NOT NULL DEFAULT 8,
    require_uppercase   BOOLEAN NOT NULL DEFAULT true,
    require_number      BOOLEAN NOT NULL DEFAULT true,
    require_special     BOOLEAN NOT NULL DEFAULT false,
    notify_admins       BOOLEAN NOT NULL DEFAULT true,
    admin_emails        JSONB NOT NULL DEFAULT '[]',
    welcome_email       BOOLEAN NOT NULL DEFAULT true,
    welcome_from_name   TEXT NOT NULL DEFAULT '',
    welcome_reply_to    TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS redirect_after TEXT NOT NULL DEFAULT '/dashboard';
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS default_role TEXT NOT NULL DEFAULT 'Client';
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS access_mode TEXT NOT NULL DEFAULT 'open';
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS domain_restrict BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS block_disposable BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS email_required BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS captcha_enabled BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS notify_admins BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS welcome_email BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE tenant_registration_settings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE tenant_registration_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_registration_settings FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_registration_settings;
CREATE POLICY tenant_isolation ON tenant_registration_settings
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');
