-- Revoke defaults. Roles themselves live outside migration lifecycle.
ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM platform_app;
ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    REVOKE SELECT ON TABLES FROM platform_readonly;
ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    REVOKE USAGE, SELECT ON SEQUENCES FROM platform_app;
ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    REVOKE SELECT ON SEQUENCES FROM platform_readonly;
ALTER DEFAULT PRIVILEGES FOR ROLE platform_admin IN SCHEMA public
    REVOKE EXECUTE ON FUNCTIONS FROM platform_app, platform_readonly;
REVOKE USAGE ON SCHEMA public FROM platform_app, platform_readonly;
