-- 0046_system_origins:
--   System-level origin configuration and ordered domain allowlist for the
--   platform control plane (/s/origin/).
--
--   Zero domains are seeded by default so the platform behaves in standard
--   unrestricted mode until the system administrator defines origin domains.

CREATE TABLE IF NOT EXISTS system_origin_settings (
    id                      INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    redirect_scheme         TEXT NOT NULL DEFAULT 'https' CHECK (redirect_scheme IN ('https', 'http')),
    trust_forwarded_headers BOOLEAN NOT NULL DEFAULT true,
    redirect_status_code    INTEGER NOT NULL DEFAULT 302 CHECK (redirect_status_code IN (301, 302, 307, 308)),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO system_origin_settings (id, redirect_scheme, trust_forwarded_headers, redirect_status_code)
VALUES (1, 'https', true, 302)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS system_origin_domains (
    id          BIGSERIAL PRIMARY KEY,
    domain      CITEXT NOT NULL UNIQUE,
    note        TEXT NOT NULL DEFAULT '',
    position    INTEGER NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS system_origin_domains_order_idx
    ON system_origin_domains (position ASC, id ASC)
    WHERE enabled = true;
