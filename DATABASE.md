# Platform Database

Source of truth for the PostgreSQL schema that backs this platform. Tracks every table, its hierarchy, and how far the migration from in-memory stores to live Postgres has progressed.

---

## Architecture at a glance

| | |
|---|---|
| **Engine** | PostgreSQL 16+ |
| **Isolation** | Shared schema + Row-Level Security (RLS); `tenant_id` on every tenant-owned row |
| **Driver** | `github.com/jackc/pgx/v5` + `pgxpool` |
| **Queries** | `sqlc` — type-safe Go from hand-written SQL in `queries/*.sql` |
| **Migrations** | `golang-migrate` — SQL files in `migrations/` embedded via `//go:embed` |
| **Extensibility** | Class Table Inheritance: base `resources` table + one specialized table per kind, joined 1-1 |
| **Auditing** | `audit_log` (monthly partitioned) captures every write + every ad-hoc SQL run |
| **Roles** | `platform_app` (normal), `platform_admin` (BYPASSRLS, sysadmin only), `platform_readonly` (SQL console) |

---

## Status legend

| Symbol | Meaning |
|---|---|
| ⬜ | **Planned** — schema designed, migration not yet written |
| 🟨 | **Schema only** — migration applied, handler still uses in-memory store |
| 🟩 | **Wired** — handler reads/writes via sqlc against real Postgres |
| 🔒 | **RLS-enforced** — `tenant_id` column + policy active |
| 🧩 | **Seeded** — fixture data loaded for dev/demo |

**Current overall state:** all 187 tables are ⬜ Planned. No migrations have been applied yet; every handler still uses an in-memory `map`. This document will be updated as each phase lands.

---

## Rollout phases

| Phase | Migrations | Goal |
|---|---|---|
| **Phase 1** | 0001–0005, 0008 (partial), 0017 (partial) | Makes `/s/database` live. Sessions, api_keys, oauth_tokens, audit_log, monitoring, incidents persist. |
| **Phase 2** | 0004, 0006, 0007, 0009–0011 | Workspaces + resources CTI. Compute/storage/network/billing handlers switch to Postgres. |
| **Phase 3** | 0012–0016, 0018 | Full coverage: DB services, security, AI, apps, wizard, SSO. Every in-memory map removed. |

---

## Table inventory (187)

### Core identity — `migrations/0003_core_identity.up.sql` · Phase 1

| # | Table | Phase | Status | Partitioned | Current handler |
|---|---|---|---|---|---|
| 1 | `tenants` | 1 | ⬜ | — | `handlers/tenants.go` (in-memory) |
| 2 | `users` | 1 | ⬜ | — | `handlers/session.go` (in-memory) |
| 3 | `user_tenant_memberships` | 1 | ⬜ | — | NEW — currently implicit in `User.TenantSlug` |
| 4 | `sessions` | 1 | ⬜ | monthly | `handlers/session.go` (in-memory map) |
| 5 | `api_keys` | 1 | ⬜ | — | mocked in `handlers/database.go:114` |
| 6 | `oauth_tokens` | 1 | ⬜ | — | mocked in `handlers/database.go:125` |
| 7 | `password_reset_tokens` | 1 | ⬜ | — | NEW |
| 8 | `audit_log` | 1 | ⬜ | monthly | NEW — replaces scattered log calls |
| 9 | `user_preferences` | 1 | ⬜ | — | `handlers/api_user_settings.go` |
| 10 | `user_security_2fa` | 1 | ⬜ | — | `handlers/api_user_settings.go` |
| 11 | `user_personal_api_keys` | 1 | ⬜ | — | `handlers/api_user_settings.go` |
| 12 | `user_workspace_invitations` | 1 | ⬜ | — | `handlers/api_user_workspaces.go` |

### Workspaces & resources CTI base — `0004_workspaces_and_resources.up.sql` · Phase 1 schema / Phase 2 wired

| # | Table | Phase | Status | Notes |
|---|---|---|---|---|
| 13 | `regions` | 1 | ⬜ | Global seed (us-east-1, eu-west-1, …) |
| 14 | `workspaces` | 1/2 | ⬜ | `handlers/api_user_workspaces.go` |
| 15 | `workspace_memberships` | 2 | ⬜ | NEW |
| 16 | `resources` | 1/2 | ⬜ | **CTI base** — discriminator `kind`, 1-1 FK from specialized tables |
| 17 | `replicas` | 1 | ⬜ | mocked in `handlers/database.go:196`; global |

### RLS policies — `0005_rls_policies.up.sql` · Phase 1
### Additional RLS — `0017_rls_policies_all.up.sql` · Phase 3

Not tables — policy-only migrations. Enable `ROW LEVEL SECURITY` and attach the `tenant_isolation` policy to every tenant-scoped table as it's created.

### Billing — `0006_billing.up.sql` · Phase 2

| # | Table | Phase | Status | Partitioned | Handler |
|---|---|---|---|---|---|
| 18 | `pricing_plans` | 2 | ⬜ | — | `handlers/api_tenant_pricing.go` |
| 19 | `pricing_plan_features` | 2 | ⬜ | — | NEW |
| 20 | `tax_rules` | 2 | ⬜ | — | `handlers/api_tenant_tax.go` |
| 21 | `sla_policies` | 2 | ⬜ | — | `handlers/api_tenant_sla.go` |
| 22 | `transactions` | 2 | ⬜ | — | `handlers/api_tenant_transactions.go` |
| 23 | `invoices` | 2 | ⬜ | — | mocked in `handlers/database.go:141` |
| 24 | `invoice_line_items` | 2 | ⬜ | — | NEW |
| 25 | `usage_events` | 2 | ⬜ | monthly | mocked in `handlers/database.go:153` |
| 26 | `refund_complaints` | 2 | ⬜ | — | `handlers/api_tenant_refund_complaints.go` |

### Tenant business — `0007_tenant_business.up.sql` · Phase 2

| # | Table | Phase | Status | Handler |
|---|---|---|---|---|
| 27 | `clients` | 2 | ⬜ | `handlers/api_tenant_clients.go` |
| 28 | `client_workspaces` | 2 | ⬜ | NEW |
| 29 | `employees` | 2 | ⬜ | `handlers/api_tenant_employees.go` |
| 30 | `tickets` | 2 | ⬜ | `handlers/api_tenant_tickets.go` |
| 31 | `ticket_messages` | 2 | ⬜ | NEW |
| 32 | `abuse_reports` | 2 | ⬜ | `handlers/api_tenant_abuse_reports.go` |
| 33 | `one_time_access_tokens` | 2 | ⬜ | `handlers/api_tenant_one_time_access.go` |
| 34 | `tenant_brand` | 2 | ⬜ | `handlers/api_tenant_brand.go` |
| 35 | `tenant_localization` | 2 | ⬜ | `handlers/api_tenant_localization.go` |
| 36 | `tenant_registration_config` | 2 | ⬜ | `handlers/api_tenant_registration.go` |

### Operations — `0008_ops.up.sql` · Phase 1 schema / Phase 2 wired

| # | Table | Phase | Status | Partitioned | Handler |
|---|---|---|---|---|---|
| 37 | `monitoring_checks` | 1 | ⬜ | — | `handlers/api_tenant_monitoring.go` |
| 38 | `monitoring_results` | 1 | ⬜ | monthly | NEW |
| 39 | `ids_events` | 1 | ⬜ | monthly | `handlers/api_tenant_ids.go` |
| 40 | `incidents` | 1 | ⬜ | — | `handlers/api_tenant_incidents.go` |
| 41 | `incident_timeline_entries` | 1 | ⬜ | — | NEW |
| 42 | `incident_affected_services` | 1 | ⬜ | — | NEW |
| 43 | `status_page_components` | 2 | ⬜ | — | `handlers/api_tenant_status_page.go` |
| 44 | `maintenance_windows` | 2 | ⬜ | — | `handlers/api_tenant_maintenance.go` |
| 45 | `maintenance_affected_services` | 2 | ⬜ | — | NEW |
| 46 | `background_jobs` | 2 | ⬜ | monthly | `handlers/api_sysadmin_jobs.go` |
| 47 | `background_job_workers` | 2 | ⬜ | — | `handlers/api_sysadmin_jobs.go` |
| 48 | `background_job_queues` | 2 | ⬜ | — | `handlers/api_sysadmin_jobs.go` |

### Compute (CTI specialized + sub-resources) — `0009_compute.up.sql` · Phase 2

| # | Table | Phase | Status | CTI parent |
|---|---|---|---|---|
| 49 | `compute_k8s_clusters` | 2 | ⬜ | `resources` |
| 50 | `k8s_node_pools` | 2 | ⬜ | `compute_k8s_clusters` |
| 51 | `k8s_workloads` | 2 | ⬜ | `compute_k8s_clusters` |
| 52 | `k8s_autoscaling_policies` | 2 | ⬜ | `k8s_node_pools` |
| 53 | `compute_vm_instances` | 2 | ⬜ | `resources` |
| 54 | `vm_nics` | 2 | ⬜ | `compute_vm_instances` |
| 55 | `vm_disk_attachments` | 2 | ⬜ | `compute_vm_instances` |
| 56 | `vm_gpu_profiles` | 2 | ⬜ | `compute_vm_instances` |
| 57 | `compute_serverless_fns` | 2 | ⬜ | `resources` |
| 58 | `serverless_triggers` | 2 | ⬜ | `compute_serverless_fns` |
| 59 | `serverless_deployments` | 2 | ⬜ | `compute_serverless_fns` |
| 60 | `compute_containers` | 2 | ⬜ | `resources` |
| 61 | `container_registries` | 2 | ⬜ | `resources` |
| 62 | `container_images` | 2 | ⬜ | `container_registries` |
| 63 | `compute_batch_jobs` | 2 | ⬜ | `resources` |
| 64 | `batch_job_tasks` | 2 | ⬜ | `compute_batch_jobs` |
| 65 | `compute_web_sites` | 2 | ⬜ | `resources` |
| 66 | `web_site_deployments` | 2 | ⬜ | `compute_web_sites` |

### Storage — `0010_storage.up.sql` · Phase 2

| # | Table | Phase | Status | CTI parent |
|---|---|---|---|---|
| 67 | `storage_block_volumes` | 2 | ⬜ | `resources` |
| 68 | `storage_file_shares` | 2 | ⬜ | `resources` |
| 69 | `file_share_access_rules` | 2 | ⬜ | `storage_file_shares` |
| 70 | `storage_object_buckets` | 2 | ⬜ | `resources` |
| 71 | `object_lifecycle_rules` | 2 | ⬜ | `storage_object_buckets` |
| 72 | `object_bucket_cors_rules` | 2 | ⬜ | `storage_object_buckets` |
| 73 | `storage_backups` | 2 | ⬜ | `resources` |
| 74 | `backup_snapshots` | 2 | ⬜ | `storage_backups` |
| 75 | `backup_retention_policies` | 2 | ⬜ | `storage_backups` |
| 76 | `storage_archive_vaults` | 2 | ⬜ | `resources` |
| 77 | `storage_datalake_zones` | 2 | ⬜ | `resources` |

### Networking — `0011_network.up.sql` · Phase 2

| # | Table | Phase | Status | CTI parent |
|---|---|---|---|---|
| 78 | `network_vpcs` | 2 | ⬜ | `resources` |
| 79 | `vpc_subnets` | 2 | ⬜ | `network_vpcs` |
| 80 | `vpc_peerings` | 2 | ⬜ | `network_vpcs` |
| 81 | `vpc_route_tables` | 2 | ⬜ | `network_vpcs` |
| 82 | `vpc_routes` | 2 | ⬜ | `vpc_route_tables` |
| 83 | `network_security_groups` | 2 | ⬜ | `resources` |
| 84 | `nsg_rules` | 2 | ⬜ | `network_security_groups` |
| 85 | `network_dns_zones` | 2 | ⬜ | `resources` |
| 86 | `dns_records` | 2 | ⬜ | `network_dns_zones` |
| 87 | `network_load_balancers` | 2 | ⬜ | `resources` |
| 88 | `lb_listeners` | 2 | ⬜ | `network_load_balancers` |
| 89 | `lb_target_groups` | 2 | ⬜ | `network_load_balancers` |
| 90 | `lb_targets` | 2 | ⬜ | `lb_target_groups` |
| 91 | `lb_health_checks` | 2 | ⬜ | `lb_target_groups` |
| 92 | `network_cdn_distributions` | 2 | ⬜ | `resources` |
| 93 | `cdn_origin_configs` | 2 | ⬜ | `network_cdn_distributions` |
| 94 | `cdn_cache_behaviors` | 2 | ⬜ | `network_cdn_distributions` |
| 95 | `network_waf_policies` | 2 | ⬜ | `resources` |
| 96 | `waf_rules` | 2 | ⬜ | `network_waf_policies` |
| 97 | `network_ddos_profiles` | 2 | ⬜ | `resources` |

### Database services — `0012_database_services.up.sql` · Phase 3

| # | Table | Phase | Status | CTI parent |
|---|---|---|---|---|
| 98 | `db_postgres_instances` | 3 | ⬜ | `resources` |
| 99 | `pg_databases` | 3 | ⬜ | `db_postgres_instances` |
| 100 | `pg_roles` | 3 | ⬜ | `db_postgres_instances` |
| 101 | `pg_replication_slots` | 3 | ⬜ | `db_postgres_instances` |
| 102 | `db_mysql_instances` | 3 | ⬜ | `resources` |
| 103 | `db_sqlserver_instances` | 3 | ⬜ | `resources` |
| 104 | `db_cloud_native_instances` | 3 | ⬜ | `resources` |
| 105 | `db_document_instances` | 3 | ⬜ | `resources` |
| 106 | `db_keyvalue_instances` | 3 | ⬜ | `resources` |
| 107 | `db_widecolumn_instances` | 3 | ⬜ | `resources` |
| 108 | `db_graph_instances` | 3 | ⬜ | `resources` |
| 109 | `db_memory_instances` | 3 | ⬜ | `resources` |
| 110 | `db_memory_session_stores` | 3 | ⬜ | `db_memory_instances` |
| 111 | `db_vector_indexes` | 3 | ⬜ | `resources` |
| 112 | `vector_index_embeddings_meta` | 3 | ⬜ | `db_vector_indexes` |
| 113 | `db_timeseries_instances` | 3 | ⬜ | `resources` |
| 114 | `db_search_clusters` | 3 | ⬜ | `resources` |

### Security services — `0013_security_services.up.sql` · Phase 3

| # | Table | Phase | Status | CTI parent |
|---|---|---|---|---|
| 115 | `security_identity_pools` | 3 | ⬜ | `resources` |
| 116 | `identity_policies` | 3 | ⬜ | `security_identity_pools` |
| 117 | `identity_role_bindings` | 3 | ⬜ | `security_identity_pools` |
| 118 | `identity_mfa_configs` | 3 | ⬜ | `security_identity_pools` |
| 119 | `security_kms_keys` | 3 | ⬜ | `resources` |
| 120 | `kms_key_versions` | 3 | ⬜ | `security_kms_keys` |
| 121 | `kms_key_grants` | 3 | ⬜ | `security_kms_keys` |
| 122 | `security_certificates` | 3 | ⬜ | `resources` |
| 123 | `certificate_renewals` | 3 | ⬜ | `security_certificates` |
| 124 | `security_ids_rules` | 3 | ⬜ | `resources` |
| 125 | `security_compliance_frameworks` | 3 | ⬜ | `resources` |
| 126 | `compliance_audit_reports` | 3 | ⬜ | `security_compliance_frameworks` |

### AI / ML services — `0014_ai_services.up.sql` · Phase 3

| # | Table | Phase | Status | CTI parent |
|---|---|---|---|---|
| 127 | `ai_llm_endpoints` | 3 | ⬜ | `resources` |
| 128 | `llm_deployments` | 3 | ⬜ | `ai_llm_endpoints` |
| 129 | `llm_model_variants` | 3 | ⬜ | `ai_llm_endpoints` |
| 130 | `ai_vision_apis` | 3 | ⬜ | `resources` |
| 131 | `ai_language_apis` | 3 | ⬜ | `resources` |
| 132 | `ai_speech_apis` | 3 | ⬜ | `resources` |
| 133 | `ai_recommendation_apis` | 3 | ⬜ | `resources` |
| 134 | `ai_etl_pipelines` | 3 | ⬜ | `resources` |
| 135 | `etl_stages` | 3 | ⬜ | `ai_etl_pipelines` |
| 136 | `etl_pipeline_runs` | 3 | ⬜ | `ai_etl_pipelines` |
| 137 | `ai_streaming_ingestions` | 3 | ⬜ | `resources` |
| 138 | `ai_data_catalog_assets` | 3 | ⬜ | `resources` |
| 139 | `ai_training_jobs` | 3 | ⬜ | `resources` |
| 140 | `training_job_checkpoints` | 3 | ⬜ | `ai_training_jobs` |
| 141 | `ai_finetune_jobs` | 3 | ⬜ | `resources` |
| 142 | `ai_automl_projects` | 3 | ⬜ | `resources` |
| 143 | `ai_experiments` | 3 | ⬜ | `resources` |
| 144 | `experiment_runs` | 3 | ⬜ | `ai_experiments` |
| 145 | `experiment_metrics` | 3 | ⬜ | `experiment_runs` |
| 146 | `ai_model_registry_entries` | 3 | ⬜ | `resources` |
| 147 | `model_versions` | 3 | ⬜ | `ai_model_registry_entries` |
| 148 | `model_deployments` | 3 | ⬜ | `model_versions` |
| 149 | `ai_batch_inference_jobs` | 3 | ⬜ | `resources` |
| 150 | `ai_edge_deployments` | 3 | ⬜ | `resources` |
| 151 | `ai_rag_systems` | 3 | ⬜ | `resources` |

### Apps & integrations — `0015_apps_services.up.sql` · Phase 3

| # | Table | Phase | Status | CTI parent | Partitioned |
|---|---|---|---|---|---|
| 152 | `apps_message_brokers` | 3 | ⬜ | `resources` | — |
| 153 | `broker_topics` | 3 | ⬜ | `apps_message_brokers` | — |
| 154 | `broker_subscriptions` | 3 | ⬜ | `broker_topics` | — |
| 155 | `broker_queues` | 3 | ⬜ | `apps_message_brokers` | — |
| 156 | `apps_email_senders` | 3 | ⬜ | `resources` | — |
| 157 | `email_templates` | 3 | ⬜ | `apps_email_senders` | — |
| 158 | `email_send_events` | 3 | ⬜ | `apps_email_senders` | monthly |
| 159 | `apps_media_encoders` | 3 | ⬜ | `resources` | — |
| 160 | `media_encoding_jobs` | 3 | ⬜ | `apps_media_encoders` | — |
| 161 | `apps_livestream_channels` | 3 | ⬜ | `resources` | — |
| 162 | `apps_workflow_automations` | 3 | ⬜ | `resources` | — |
| 163 | `workflow_triggers` | 3 | ⬜ | `apps_workflow_automations` | — |
| 164 | `workflow_runs` | 3 | ⬜ | `apps_workflow_automations` | — |
| 165 | `apps_scheduled_tasks` | 3 | ⬜ | `resources` | — |
| 166 | `apps_cicd_pipelines` | 3 | ⬜ | `resources` | — |
| 167 | `cicd_stages` | 3 | ⬜ | `apps_cicd_pipelines` | — |
| 168 | `cicd_runs` | 3 | ⬜ | `apps_cicd_pipelines` | — |
| 169 | `cicd_artifacts` | 3 | ⬜ | `cicd_runs` | — |
| 170 | `apps_static_sites` | 3 | ⬜ | `resources` | — |
| 171 | `apps_transactional_email_configs` | 3 | ⬜ | `resources` | — |

### Wizard & SSO — `0016_wizard_and_sso.up.sql` · Phase 3

| # | Table | Phase | Status | Scope |
|---|---|---|---|---|
| 172 | `wizard_categories` | 3 | ⬜ | global |
| 173 | `wizard_apps` | 3 | ⬜ | global |
| 174 | `wizard_app_versions` | 3 | ⬜ | global |
| 175 | `sso_oauth2_clients` | 3 | ⬜ | global |
| 176 | `sso_oauth2_authorization_codes` | 3 | ⬜ | global |
| 177 | `sso_oauth2_access_tokens` | 3 | ⬜ | global |
| 178 | `sso_oauth2_refresh_tokens` | 3 | ⬜ | global |
| 179 | `sso_oidc_providers` | 3 | ⬜ | global |
| 180 | `sso_saml_providers` | 3 | ⬜ | global |
| 181 | `sso_saml_assertions` | 3 | ⬜ | global |
| 182 | `sso_security_policies` | 3 | ⬜ | per-tenant |

### System / internal — appears across `0003`, `0008`, `0018` · mixed phases

| # | Table | Phase | Status | Notes |
|---|---|---|---|---|
| 183 | `schema_migrations` | 1 | ⬜ | Managed by golang-migrate itself |
| 184 | `system_logs` | 1 | ⬜ | monthly partitioned · `/s/logs` |
| 185 | `system_feature_flags` | 2 | ⬜ | global |
| 186 | `system_announcements` | 2 | ⬜ | global |
| 187 | `system_accounting_webhooks` | 2 | ⬜ | `/s/accounting` webhooks |

---

## Migration index

| File | Tables created | Phase |
|---|---|---|
| `migrations/0001_extensions.up.sql` | (none — installs `pgcrypto`, `citext`, `pg_trgm`) | 1 |
| `migrations/0002_roles.up.sql` | (none — creates `platform_app`, `platform_admin`, `platform_readonly`) | 1 |
| `migrations/0003_core_identity.up.sql` | 1–12, 183 | 1 |
| `migrations/0004_workspaces_and_resources.up.sql` | 13–17 | 1 |
| `migrations/0005_rls_policies.up.sql` | (policies only) | 1 |
| `migrations/0006_billing.up.sql` | 18–26 | 2 |
| `migrations/0007_tenant_business.up.sql` | 27–36 | 2 |
| `migrations/0008_ops.up.sql` | 37–48, 184 | 1/2 |
| `migrations/0009_compute.up.sql` | 49–66 | 2 |
| `migrations/0010_storage.up.sql` | 67–77 | 2 |
| `migrations/0011_network.up.sql` | 78–97 | 2 |
| `migrations/0012_database_services.up.sql` | 98–114 | 3 |
| `migrations/0013_security_services.up.sql` | 115–126 | 3 |
| `migrations/0014_ai_services.up.sql` | 127–151 | 3 |
| `migrations/0015_apps_services.up.sql` | 152–171 | 3 |
| `migrations/0016_wizard_and_sso.up.sql` | 172–182 | 3 |
| `migrations/0017_rls_policies_all.up.sql` | (policies only) | 3 |
| `migrations/0018_indexes.up.sql` | (indexes only) | 3 |

Every `.up.sql` has a matching `.down.sql` for safe rollback.

---

## Implementation progress

| Phase | Tables | ⬜ Planned | 🟨 Schema only | 🟩 Wired |
|---|---|---|---|---|
| 1 | 35 | 35 | 0 | 0 |
| 2 | 75 | 75 | 0 | 0 |
| 3 | 77 | 77 | 0 | 0 |
| **Total** | **187** | **187** | **0** | **0** |

Updated: 2026-04-18 — initial inventory, no migrations applied yet.

---

## How to update this document

1. When you add a migration, add/verify the table rows in the section above.
2. When you convert a handler from in-memory → sqlc, flip its row from ⬜ to 🟩.
3. When an RLS policy is attached, add the 🔒 badge.
4. Update the **Implementation progress** totals.

Each migration PR must update this file — enforced by `scripts/check-database-md.sh` in CI (Phase 2).
