-- 0002_roles: grant default privileges on objects that migrations (run as
-- platform_admin) will subsequently create. The three roles themselves
-- (platform_admin, platform_app, platform_readonly) must exist before this
-- migration runs — see scripts/db-init.sh for creation with passwords.

-- Safety: verify the expected roles exist. Fail fast if db-init.sh was skipped.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'platform_app') THEN
        RAISE EXCEPTION 'role platform_app does not exist — run scripts/db-init.sh first';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'platform_readonly') THEN
        RAISE EXCEPTION 'role platform_readonly does not exist — run scripts/db-init.sh first';
    END IF;
END $$;

-- Connect privilege (idempotent; may already be granted by db-init.sh).
GRANT CONNECT ON DATABASE platform TO platform_app, platform_readonly;

-- Usage on the public schema.
GRANT USAGE ON SCHEMA public TO platform_app, platform_readonly;

-- Default privileges for objects created henceforth by platform_admin:
-- platform_app gets full DML, platform_readonly gets SELECT only.
ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO platform_app;

ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    GRANT SELECT ON TABLES TO platform_readonly;

ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO platform_app;

ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    GRANT SELECT ON SEQUENCES TO platform_readonly;

ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    GRANT EXECUTE ON FUNCTIONS TO platform_app, platform_readonly;
