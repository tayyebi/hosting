-- 0030 rollback: drop MLOps Registry + ML Monitoring + Security IAM tables
DROP TABLE IF EXISTS workspace_security_rbac_roles;
DROP TABLE IF EXISTS workspace_security_mfa_users;
DROP TABLE IF EXISTS workspace_security_sso_providers;
DROP TABLE IF EXISTS workspace_ai_ml_monitors;
DROP TABLE IF EXISTS workspace_ai_model_registry;
