-- 0015: pricing_plans, tax_rules, sla_policies, tenant_brand_settings, tenant_localization_settings

-- ── pricing_plans ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS pricing_plans (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL DEFAULT '',
    tier        TEXT NOT NULL DEFAULT 'starter' CHECK (tier IN ('free','starter','pro','enterprise')),
    status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('active','archived','draft')),
    price       NUMERIC(10,2) NOT NULL DEFAULT 0,
    currency    TEXT NOT NULL DEFAULT 'USD',
    interval    TEXT NOT NULL DEFAULT 'monthly' CHECK (interval IN ('monthly','annually','one-time')),
    features    TEXT NOT NULL DEFAULT '',
    max_users   INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'draft';
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS tier TEXT NOT NULL DEFAULT 'starter';
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD';
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS interval TEXT NOT NULL DEFAULT 'monthly';
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS features TEXT NOT NULL DEFAULT '';
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS max_users INTEGER NOT NULL DEFAULT 0;
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS pricing_plans_tenant_idx  ON pricing_plans (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS pricing_plans_status_idx  ON pricing_plans (tenant_id, status) WHERE deleted_at IS NULL;

-- ── tax_rules ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tax_rules (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL DEFAULT '',
    region      TEXT NOT NULL DEFAULT '',
    tax_type    TEXT NOT NULL DEFAULT 'vat' CHECK (tax_type IN ('vat','sales_tax','gst','withholding','exempt')),
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','pending')),
    rate        NUMERIC(6,2) NOT NULL DEFAULT 0,
    applies_to  TEXT NOT NULL DEFAULT 'all',
    valid_from  DATE,
    valid_until DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
ALTER TABLE tax_rules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE tax_rules ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE tax_rules ADD COLUMN IF NOT EXISTS tax_type TEXT NOT NULL DEFAULT 'vat';
ALTER TABLE tax_rules ADD COLUMN IF NOT EXISTS region TEXT NOT NULL DEFAULT '';
ALTER TABLE tax_rules ADD COLUMN IF NOT EXISTS rate NUMERIC(6,2) NOT NULL DEFAULT 0;
ALTER TABLE tax_rules ADD COLUMN IF NOT EXISTS applies_to TEXT NOT NULL DEFAULT 'all';
ALTER TABLE tax_rules ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS tax_rules_tenant_idx  ON tax_rules (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS tax_rules_status_idx  ON tax_rules (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS tax_rules_type_idx    ON tax_rules (tenant_id, tax_type) WHERE deleted_at IS NULL;

-- ── sla_policies ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS sla_policies (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name          TEXT NOT NULL DEFAULT '',
    service       TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','breached','at-risk')),
    uptime_target NUMERIC(7,3) NOT NULL DEFAULT 99.950,
    uptime_actual NUMERIC(7,3) NOT NULL DEFAULT 100.000,
    response_time TEXT NOT NULL DEFAULT '',
    credit_policy TEXT NOT NULL DEFAULT '',
    penalty_rate  TEXT NOT NULL DEFAULT '',
    review_period TEXT NOT NULL DEFAULT 'monthly' CHECK (review_period IN ('monthly','quarterly','annually')),
    last_reviewed DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS service TEXT NOT NULL DEFAULT '';
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS uptime_target NUMERIC(7,3) NOT NULL DEFAULT 99.950;
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS uptime_actual NUMERIC(7,3) NOT NULL DEFAULT 100.000;
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS response_time TEXT NOT NULL DEFAULT '';
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS credit_policy TEXT NOT NULL DEFAULT '';
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS penalty_rate TEXT NOT NULL DEFAULT '';
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS review_period TEXT NOT NULL DEFAULT 'monthly';
ALTER TABLE sla_policies ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS sla_policies_tenant_idx  ON sla_policies (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS sla_policies_status_idx  ON sla_policies (tenant_id, status) WHERE deleted_at IS NULL;

-- ── tenant_brand_settings ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenant_brand_settings (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    company_name    TEXT NOT NULL DEFAULT '',
    tagline         TEXT NOT NULL DEFAULT '',
    website         TEXT NOT NULL DEFAULT '',
    support_email   TEXT NOT NULL DEFAULT '',
    logo_url        TEXT NOT NULL DEFAULT '',
    colors          JSONB NOT NULL DEFAULT '[]',
    social          JSONB NOT NULL DEFAULT '{}',
    dev_portal      JSONB NOT NULL DEFAULT '{}',
    dns_entries     JSONB NOT NULL DEFAULT '[]',
    email_templates JSONB NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── tenant_localization_settings ──────────────────────────────
CREATE TABLE IF NOT EXISTS tenant_localization_settings (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    language         TEXT NOT NULL DEFAULT 'en',
    timezone         TEXT NOT NULL DEFAULT 'UTC',
    rtl_enabled      BOOLEAN NOT NULL DEFAULT false,
    date_format      TEXT NOT NULL DEFAULT 'DD/MM/YYYY',
    time_format      TEXT NOT NULL DEFAULT '24-hour',
    number_format    TEXT NOT NULL DEFAULT '1,234,567.89',
    first_day_of_week TEXT NOT NULL DEFAULT 'Monday',
    currency         TEXT NOT NULL DEFAULT 'USD',
    symbol_position  TEXT NOT NULL DEFAULT 'before',
    tax_inclusive    BOOLEAN NOT NULL DEFAULT false,
    tax_breakdown    BOOLEAN NOT NULL DEFAULT false,
    routing_rules    JSONB NOT NULL DEFAULT '[]',
    api_locale       JSONB NOT NULL DEFAULT '{}',
    cdn_stats        JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── RLS ───────────────────────────────────────────────────────
ALTER TABLE pricing_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE pricing_plans FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON pricing_plans;
CREATE POLICY tenant_isolation ON pricing_plans
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE tax_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE tax_rules FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tax_rules;
CREATE POLICY tenant_isolation ON tax_rules
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE sla_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE sla_policies FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON sla_policies;
CREATE POLICY tenant_isolation ON sla_policies
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE tenant_brand_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_brand_settings FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_brand_settings;
CREATE POLICY tenant_isolation ON tenant_brand_settings
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');

ALTER TABLE tenant_localization_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_localization_settings FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_localization_settings;
CREATE POLICY tenant_isolation ON tenant_localization_settings
    USING (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin')
    WITH CHECK (tenant_id = public.current_tenant_id() OR public.current_actor_role() = 'sysadmin');
