#!/usr/bin/env bash
# Runs once on first-boot of the postgres container (docker-entrypoint-initdb.d).
# Creates the three roles used by the platform:
#   - platform_admin   (owner; runs migrations; BYPASSRLS implicitly as owner)
#   - platform_app     (normal runtime role; RLS applied)
#   - platform_readonly (SELECT-only; used by the sysadmin SQL console)
#
# POSTGRES_USER in compose.yaml is already set to platform_admin, so the
# cluster is bootstrapped with platform_admin as the superuser. This script
# only creates the two additional runtime roles and grants base privileges.

set -euo pipefail

: "${PG_APP_PASSWORD:?PG_APP_PASSWORD must be set}"
: "${PG_READONLY_PASSWORD:?PG_READONLY_PASSWORD must be set}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'platform_app') THEN
            CREATE ROLE platform_app LOGIN PASSWORD '${PG_APP_PASSWORD}';
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'platform_readonly') THEN
            CREATE ROLE platform_readonly LOGIN PASSWORD '${PG_READONLY_PASSWORD}';
        END IF;
    END
    \$\$;

    -- Base privileges on the platform database
    GRANT CONNECT ON DATABASE platform TO platform_app, platform_readonly;
    GRANT USAGE ON SCHEMA public TO platform_app, platform_readonly;

    -- Migrations (run as platform_admin) use ALTER DEFAULT PRIVILEGES to
    -- extend DML/SELECT grants to these roles for every future table.
EOSQL
