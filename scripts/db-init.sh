#!/usr/bin/env bash
# Runs once on first-boot of the postgres container (docker-entrypoint-initdb.d).
# Creates the three roles used by the platform:
#   - platform_admin   (owner; runs migrations; BYPASSRLS implicitly as owner)
#   - platform_app     (normal runtime role; RLS applied)
#   - platform_readonly (SELECT-only; used by the sysadmin SQL console)
#
# In compose.yaml, PostgreSQL runs passwordless (POSTGRES_HOST_AUTH_METHOD=trust)
# inside an isolated internal Docker network. Optional PG_APP_PASSWORD and
# PG_READONLY_PASSWORD values are still supported for external setups.

set -euo pipefail

APP_PW_SQL=""
if [ -n "${PG_APP_PASSWORD:-}" ]; then
    APP_PW_SQL=" PASSWORD '${PG_APP_PASSWORD}'"
fi

RO_PW_SQL=""
if [ -n "${PG_READONLY_PASSWORD:-}" ]; then
    RO_PW_SQL=" PASSWORD '${PG_READONLY_PASSWORD}'"
fi

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'platform_app') THEN
            CREATE ROLE platform_app LOGIN${APP_PW_SQL};
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'platform_readonly') THEN
            CREATE ROLE platform_readonly LOGIN${RO_PW_SQL};
        END IF;
    END
    \$\$;

    -- Base privileges on the platform database
    GRANT CONNECT ON DATABASE platform TO platform_app, platform_readonly;
    GRANT USAGE ON SCHEMA public TO platform_app, platform_readonly;

    -- Migrations (run as platform_admin) use ALTER DEFAULT PRIVILEGES to
    -- extend DML/SELECT grants to these roles for every future table.
EOSQL
