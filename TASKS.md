# Tasks

# API System
- [x] Implement mock REST api for Create a base API system RESTful with swagger, so that the UI can use it - UI must use JS to communicate with the REST api.

## Authentication

- [x] Implement mock REST api for `/login` — Login
- [x] Implement mock REST api for `/logout` — Logout
- [x] Implement mock REST api for `/sso-guide` — SSO Guide
- [x] Implement mock REST api for `/status` — Status

## Sysadmin (`/s/`)

- [x] Implement mock REST api for `/s/` — Sysadmin Dashboard
- [x] Implement mock REST api for `/s/tenants/` — Tenants
- [x] Implement mock REST api for `/s/logs/` — Logs
- [x] Implement mock REST api for `/s/jobs/` — Jobs
- [x] Implement mock REST api for `/s/database/` — Database
- [x] Implement mock REST api for `/s/replicas/` — Replicas
- [x] Implement mock REST api for `/s/accounting/` — Accounting
- [x] Implement mock REST api for `/s/current-replica/` — Current Replica
- [x] Implement mock REST api for `/s/sso/` — SSO

## User (`/u/`)

- [x] Implement mock REST api for `/u/settings/` — Settings
- [x] Implement mock REST api for `/u/workspaces/` — Workspaces

## Wizard

- [x] Implement mock REST api for `/wizard` — Wizard Landing
- [x] Implement mock REST api for `/wizard/{slug}` — Wizard Step


## Tenant (`/t/{slug}/`)

- [x] Implement mock REST api for `/t/{slug}/` — Tenant Dashboard
- [x] Implement mock REST api for `/t/{slug}/clients/` — Clients
- [x] Implement mock REST api for `/t/{slug}/employees/` — Employees
- [x] Implement mock REST api for `/t/{slug}/tickets/` — Tickets
- [x] Implement mock REST api for `/t/{slug}/abuse-reports/` — Abuse Reports
- [x] Implement mock REST api for `/t/{slug}/refund-complaints/` — Refund Complaints
- [x] Implement mock REST api for `/t/{slug}/one-time-access/` — One-time Access
- [x] Implement mock REST api for `/t/{slug}/resources/` — Resources
- [x] Implement mock REST api for `/t/{slug}/monitoring/` — Monitoring
- [x] Implement mock REST api for `/t/{slug}/ids/` — IDS
- [x] Implement mock REST api for `/t/{slug}/incidents/` — Incidents
- [x] Implement mock REST api for `/t/{slug}/status-page/` — Status Page
- [x] Implement mock REST api for `/t/{slug}/maintenance/` — Maintenance
- [x] Implement mock REST api for `/t/{slug}/transactions/` — Transactions
- [x] Implement mock REST api for `/t/{slug}/pricing/` — Pricing
- [x] Implement mock REST api for `/t/{slug}/tax/` — Tax
- [x] Implement mock REST api for `/t/{slug}/sla/` — SLA
- [x] Implement mock REST api for `/t/{slug}/brand/` — Brand
- [x] Implement mock REST api for `/t/{slug}/localization/` — Localization
- [x] Implement mock REST api for `/t/{slug}/registration/` — Registration Settings
- [x] Implement mock REST api for `/t/{slug}/register` — Register Form

## Workspace Overview

- [x] Implement mock REST api for `/t/{slug}/w/{id}/` — Workspace Overview

## Workspace Leaf Pages

### Compute

- [x] Implement mock REST api for `compute/vms/general` — General purpose VMs
- [x] Implement mock REST api for `compute/vms/gpu` — GPU / accelerated VMs
- [x] Implement mock REST api for `compute/vms/spot` — Spot / preemptible VMs
- [x] Implement mock REST api for `compute/containers/kubernetes` — Managed Kubernetes
- [x] Implement mock REST api for `compute/containers/registry` — Container registry
- [x] Implement mock REST api for `compute/containers/instances` — Container instances
- [x] Implement mock REST api for `compute/hosting/web` — Web app hosting
- [x] Implement mock REST api for `compute/hosting/static` — Static site hosting
- [x] Implement mock REST api for `compute/serverless/functions` — Functions as a service
- [x] Implement mock REST api for `compute/serverless/edge` — Edge functions
- [x] Implement mock REST api for `compute/batch/jobs` — Job scheduling
- [x] Implement mock REST api for `compute/batch/hpc` — HPC / grid computing

### Storage

- [x] Implement mock REST api for `storage/object/standard` — Standard tier
- [x] Implement mock REST api for `storage/object/archive` — Archive / cold tier
- [x] Implement mock REST api for `storage/object/datalake` — Data lake
- [x] Implement mock REST api for `storage/block/ssd` — SSD high performance
- [x] Implement mock REST api for `storage/block/hdd` — HDD standard
- [x] Implement mock REST api for `storage/file/nfs` — NFS / SMB shares
- [x] Implement mock REST api for `storage/file/distributed` — Distributed file systems
- [x] Implement mock REST api for `storage/backup/snapshots` — Snapshot management
- [x] Implement mock REST api for `storage/backup/dr` — Disaster recovery
- [x] Implement mock REST api for `storage/backup/archive` — Long-term archiving

### Databases

- [x] Implement mock REST api for `databases/relational/mysql-pg` — MySQL / PostgreSQL
- [x] Implement mock REST api for `databases/relational/sql-server` — SQL Server
- [x] Implement mock REST api for `databases/relational/cloud-native` — Cloud-native
- [x] Implement mock REST api for `databases/nosql/document` — Document
- [x] Implement mock REST api for `databases/nosql/keyvalue` — Key-value
- [x] Implement mock REST api for `databases/nosql/wide-column` — Wide-column
- [x] Implement mock REST api for `databases/nosql/graph` — Graph
- [x] Implement mock REST api for `databases/memory/cache` — Cache
- [x] Implement mock REST api for `databases/memory/session` — Session store
- [x] Implement mock REST api for `databases/analytics/warehouse` — Data warehouse
- [x] Implement mock REST api for `databases/analytics/timeseries` — Time-series
- [x] Implement mock REST api for `databases/analytics/search` — Search & full-text

### Networking

- [x] Implement mock REST api for `networking/vnet/vpc` — VPC / subnets
- [x] Implement mock REST api for `networking/vnet/peering` — Peering & transit
- [x] Implement mock REST api for `networking/vnet/nsg` — Network security groups
- [x] Implement mock REST api for `networking/lb/app` — Application load balancer
- [x] Implement mock REST api for `networking/lb/network` — Network load balancer
- [x] Implement mock REST api for `networking/lb/global` — Global load balancing
- [x] Implement mock REST api for `networking/dns/manage` — DNS management
- [x] Implement mock REST api for `networking/dns/domains` — Domain registration
- [x] Implement mock REST api for `networking/cdn/static` — Static asset delivery
- [x] Implement mock REST api for `networking/cdn/video` — Video streaming
- [x] Implement mock REST api for `networking/cdn/edge` — Edge caching
- [x] Implement mock REST api for `networking/private/vpn` — VPN gateway
- [x] Implement mock REST api for `networking/private/direct` — Direct connect
- [x] Implement mock REST api for `networking/private/sdwan` — SD-WAN

### AI & Data Intelligence

- [x] Implement mock REST api for `ai/training/distributed` — Distributed training
- [x] Implement mock REST api for `ai/training/automl` — AutoML
- [x] Implement mock REST api for `ai/training/finetuning` — Fine-tuning
- [x] Implement mock REST api for `ai/hosting/inference` — Inference endpoints
- [x] Implement mock REST api for `ai/hosting/batch` — Batch inference
- [x] Implement mock REST api for `ai/hosting/edge` — Edge deployment
- [x] Implement mock REST api for `ai/apis/vision` — Vision
- [x] Implement mock REST api for `ai/apis/language` — Language / NLP
- [x] Implement mock REST api for `ai/apis/speech` — Speech
- [x] Implement mock REST api for `ai/apis/recommendation` — Recommendation
- [x] Implement mock REST api for `ai/data/etl` — ETL / ELT pipelines
- [x] Implement mock REST api for `ai/data/streaming` — Streaming ingestion
- [x] Implement mock REST api for `ai/data/catalog` — Data catalog
- [x] Implement mock REST api for `ai/mlops/experiments` — Experiment tracking
- [x] Implement mock REST api for `ai/mlops/registry` — Model registry
- [x] Implement mock REST api for `ai/mlops/monitoring` — Monitoring & drift detection

### Security & Identity

- [x] Implement mock REST api for `security/iam/sso` — SSO / federation
- [x] Implement mock REST api for `security/iam/mfa` — Multi-factor authentication
- [x] Implement mock REST api for `security/iam/rbac` — Role-based access control
- [x] Implement mock REST api for `security/data/enc-rest` — Encryption at rest
- [x] Implement mock REST api for `security/data/enc-transit` — Encryption in transit
- [x] Implement mock REST api for `security/data/kms` — Key management
- [x] Implement mock REST api for `security/network/waf` — Web application firewall
- [x] Implement mock REST api for `security/network/ddos` — DDoS protection
- [x] Implement mock REST api for `security/network/ids` — Intrusion detection
- [x] Implement mock REST api for `security/compliance/policy` — Policy management
- [x] Implement mock REST api for `security/compliance/audit` — Audit logging
- [x] Implement mock REST api for `security/compliance/regulatory` — Regulatory frameworks

### Developer Tools & DevOps

- [x] Implement mock REST api for `devops/cicd/source` — Source control
- [x] Implement mock REST api for `devops/cicd/build` — Build automation
- [x] Implement mock REST api for `devops/cicd/deploy` — Deployment automation
- [x] Implement mock REST api for `devops/iac/templates` — Templating
- [x] Implement mock REST api for `devops/iac/config` — Config management
- [x] Implement mock REST api for `devops/api/gateway` — API gateway
- [x] Implement mock REST api for `devops/api/portal` — Developer portal
- [x] Implement mock REST api for `devops/api/analytics` — API analytics
- [x] Implement mock REST api for `devops/obs/logging` — Logging
- [x] Implement mock REST api for `devops/obs/metrics` — Metrics & dashboards
- [x] Implement mock REST api for `devops/obs/tracing` — Distributed tracing

### Application Services

- [x] Implement mock REST api for `apps/messaging/queues` — Message queues
- [x] Implement mock REST api for `apps/messaging/streaming` — Event streaming
- [x] Implement mock REST api for `apps/messaging/servicebus` — Service bus
- [x] Implement mock REST api for `apps/workflow/automation` — Workflow automation
- [x] Implement mock REST api for `apps/workflow/scheduler` — Scheduler / cron jobs
- [x] Implement mock REST api for `apps/media/encoding` — Video encoding
- [x] Implement mock REST api for `apps/media/livestream` — Live streaming
- [x] Implement mock REST api for `apps/media/storage` — Media storage

### Multi-Cloud & Hybrid

- [x] Implement mock REST api for `multicloud/hybrid/onprem` — On-premises extension
- [x] Implement mock REST api for `multicloud/hybrid/control` — Unified control plane
- [x] Implement mock REST api for `multicloud/hybrid/edge` — Edge computing
- [x] Implement mock REST api for `multicloud/mgmt/cost` — Cost management
- [x] Implement mock REST api for `multicloud/mgmt/policy` — Policy governance
- [x] Implement mock REST api for `multicloud/mgmt/workload` — Workload portability
- [x] Implement mock REST api for `multicloud/migration/assessment` — Assessment & discovery
- [x] Implement mock REST api for `multicloud/migration/data` — Data migration
- [x] Implement mock REST api for `multicloud/migration/app` — Application migration

## Workspace Pro Feature Pages

- [x] Implement mock REST api for `apps/app-hosting/static-sites` — Websites & Apps
- [x] Implement mock REST api for `compute/virtual-machines/linux-vms` — Cloud Servers
- [x] Implement mock REST api for `databases/relational/postgresql` — Managed Databases
- [x] Implement mock REST api for `storage/object-storage/s3-compatible` — Storage & Backups
- [x] Implement mock REST api for `apps/communication/transactional-email` — Email & Messaging
- [x] Implement mock REST api for `networking/cdn/pop-management` — Content Delivery
- [x] Implement mock REST api for `security/certificates/ssl-tls` — Security & SSL
- [x] Implement mock REST api for `devops/monitoring/metrics` — Monitoring & Alerts
- [x] Implement mock REST api for `ai/inference/llm-endpoints` — AI & Automation
- [x] Implement mock REST api for `apps/serverless/functions` — Serverless Functions
- [x] Implement mock REST api for `networking/dns/zones` — Domains & DNS
- [x] Implement mock REST api for `compute/kubernetes/managed-k8s` — Managed Kubernetes
- [x] Implement mock REST api for `storage/volumes/block-storage` — Block Storage
- [x] Implement mock REST api for `databases/nosql/document-db` — Document Database
- [x] Implement mock REST api for `networking/load-balancer/application-lb` — Application Load Balancer
- [x] Implement mock REST api for `ai/vector-search/embeddings` — Vector Search
- [x] Implement mock REST api for `security/iam/identity-platform` — Identity Platform
- [x] Implement mock REST api for `security/firewall/cloud-waf` — Cloud WAF
- [x] Implement mock REST api for `devops/cicd/pipelines` — CI/CD Pipelines
- [x] Implement mock REST api for `apps/queues/message-broker` — Message Broker
- [x] Implement mock REST api for `multicloud/cost/cloud-optimizer` — Cost Optimizer


## Sysadmin (`/s/`)

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/s/logs/` — Logs
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/s/jobs/` — Jobs
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/s/database/` — Database
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/s/replicas/` — Replicas
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/s/accounting/` — Accounting
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/s/current-replica/` — Current Replica
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/s/sso/` — SSO

## User (`/u/`)

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/u/settings/` — Settings
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/u/workspaces/` — Workspaces

## Wizard

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/wizard` — Wizard Landing
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/wizard/{slug}` — Wizard Step

## Error Pages

- [x] make the page lean, enhance the ux, make them specificly match to its duty: Error page (`templates/error.html`) — All HTTP error codes

## Tenant (`/t/{slug}/`)

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/` — Tenant Dashboard
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/clients/` — Clients
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/employees/` — Employees
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/tickets/` — Tickets
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/abuse-reports/` — Abuse Reports
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/refund-complaints/` — Refund Complaints
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/one-time-access/` — One-time Access
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/resources/` — Resources
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/monitoring/` — Monitoring
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/ids/` — IDS
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/incidents/` — Incidents
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/status-page/` — Status Page
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/maintenance/` — Maintenance
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/transactions/` — Transactions
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/pricing/` — Pricing
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/tax/` — Tax
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/sla/` — SLA
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/brand/` — Brand
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/localization/` — Localization
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/registration/` — Registration Settings
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/register` — Register Form

## Workspace Overview

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `/t/{slug}/w/{id}/` — Workspace Overview

## Workspace Leaf Pages

### Compute

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/vms/general` — General purpose VMs
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/vms/gpu` — GPU / accelerated VMs
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/vms/spot` — Spot / preemptible VMs
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/containers/kubernetes` — Managed Kubernetes
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/containers/registry` — Container registry
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/containers/instances` — Container instances
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/hosting/web` — Web app hosting
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/hosting/static` — Static site hosting
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/serverless/functions` — Functions as a service
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/serverless/edge` — Edge functions
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/batch/jobs` — Job scheduling
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/batch/hpc` — HPC / grid computing

### Storage

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/object/standard` — Standard tier
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/object/archive` — Archive / cold tier
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/object/datalake` — Data lake
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/block/ssd` — SSD high performance
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/block/hdd` — HDD standard
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/file/nfs` — NFS / SMB shares
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/file/distributed` — Distributed file systems
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/backup/snapshots` — Snapshot management
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/backup/dr` — Disaster recovery
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/backup/archive` — Long-term archiving

### Databases

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/relational/mysql-pg` — MySQL / PostgreSQL
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/relational/sql-server` — SQL Server
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/relational/cloud-native` — Cloud-native
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/nosql/document` — Document
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/nosql/keyvalue` — Key-value
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/nosql/wide-column` — Wide-column
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/nosql/graph` — Graph
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/memory/cache` — Cache
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/memory/session` — Session store
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/analytics/warehouse` — Data warehouse
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/analytics/timeseries` — Time-series
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/analytics/search` — Search & full-text

### Networking

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/vnet/vpc` — VPC / subnets
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/vnet/peering` — Peering & transit
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/vnet/nsg` — Network security groups
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/lb/app` — Application load balancer
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/lb/network` — Network load balancer
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/lb/global` — Global load balancing
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/dns/manage` — DNS management
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/dns/domains` — Domain registration
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/cdn/static` — Static asset delivery
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/cdn/video` — Video streaming
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/cdn/edge` — Edge caching
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/private/vpn` — VPN gateway
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/private/direct` — Direct connect
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/private/sdwan` — SD-WAN

### AI & Data Intelligence

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/training/distributed` — Distributed training
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/training/automl` — AutoML
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/training/finetuning` — Fine-tuning
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/hosting/inference` — Inference endpoints
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/hosting/batch` — Batch inference
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/hosting/edge` — Edge deployment
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/apis/vision` — Vision
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/apis/language` — Language / NLP
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/apis/speech` — Speech
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/apis/recommendation` — Recommendation
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/data/etl` — ETL / ELT pipelines
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/data/streaming` — Streaming ingestion
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/data/catalog` — Data catalog
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/mlops/experiments` — Experiment tracking
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/mlops/registry` — Model registry
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/mlops/monitoring` — Monitoring & drift detection

### Security & Identity

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/iam/sso` — SSO / federation
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/iam/mfa` — Multi-factor authentication
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/iam/rbac` — Role-based access control
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/data/enc-rest` — Encryption at rest
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/data/enc-transit` — Encryption in transit
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/data/kms` — Key management
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/network/waf` — Web application firewall
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/network/ddos` — DDoS protection
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/network/ids` — Intrusion detection
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/compliance/policy` — Policy management
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/compliance/audit` — Audit logging
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/compliance/regulatory` — Regulatory frameworks

### Developer Tools & DevOps

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/cicd/source` — Source control
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/cicd/build` — Build automation
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/cicd/deploy` — Deployment automation
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/iac/templates` — Templating
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/iac/config` — Config management
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/api/gateway` — API gateway
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/api/portal` — Developer portal
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/api/analytics` — API analytics
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/obs/logging` — Logging
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/obs/metrics` — Metrics & dashboards
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/obs/tracing` — Distributed tracing

### Application Services

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/messaging/queues` — Message queues
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/messaging/streaming` — Event streaming
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/messaging/servicebus` — Service bus
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/workflow/automation` — Workflow automation
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/workflow/scheduler` — Scheduler / cron jobs
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/media/encoding` — Video encoding
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/media/livestream` — Live streaming
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/media/storage` — Media storage

### Multi-Cloud & Hybrid

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/hybrid/onprem` — On-premises extension
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/hybrid/control` — Unified control plane
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/hybrid/edge` — Edge computing
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/mgmt/cost` — Cost management
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/mgmt/policy` — Policy governance
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/mgmt/workload` — Workload portability
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/migration/assessment` — Assessment & discovery
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/migration/data` — Data migration
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/migration/app` — Application migration

## Workspace Pro Feature Pages

- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/app-hosting/static-sites` — Websites & Apps
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/virtual-machines/linux-vms` — Cloud Servers
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/relational/postgresql` — Managed Databases
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/object-storage/s3-compatible` — Storage & Backups
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/communication/transactional-email` — Email & Messaging
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/cdn/pop-management` — Content Delivery
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/certificates/ssl-tls` — Security & SSL
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/monitoring/metrics` — Monitoring & Alerts
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/inference/llm-endpoints` — AI & Automation
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/serverless/functions` — Serverless Functions
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/dns/zones` — Domains & DNS
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `compute/kubernetes/managed-k8s` — Managed Kubernetes
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `storage/volumes/block-storage` — Block Storage
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `databases/nosql/document-db` — Document Database
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `networking/load-balancer/application-lb` — Application Load Balancer
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `ai/vector-search/embeddings` — Vector Search
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/iam/identity-platform` — Identity Platform
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `security/firewall/cloud-waf` — Cloud WAF
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `devops/cicd/pipelines` — CI/CD Pipelines
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `apps/queues/message-broker` — Message Broker
- [x] make the page lean, enhance the ux, make them specificly match to its duty: `multicloud/cost/cloud-optimizer` — Cost Optimizer


## Sysadmin (`/s/`)

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/s/logs/` — Logs
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/s/jobs/` — Jobs
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/s/database/` — Database
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/s/replicas/` — Replicas
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/s/accounting/` — Accounting
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/s/current-replica/` — Current Replica
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/s/sso/` — SSO

## User (`/u/`)

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/u/settings/` — Settings
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/u/workspaces/` — Workspaces

## Wizard

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/wizard` — Wizard Landing
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/wizard/{slug}` — Wizard Step

## Error Pages

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: Error page (`templates/error.html`) — All HTTP error codes

## Tenant (`/t/{slug}/`)

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/` — Tenant Dashboard
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/clients/` — Clients
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/employees/` — Employees
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/tickets/` — Tickets
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/abuse-reports/` — Abuse Reports
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/refund-complaints/` — Refund Complaints
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/one-time-access/` — One-time Access
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/resources/` — Resources
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/monitoring/` — Monitoring
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/ids/` — IDS
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/incidents/` — Incidents
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/status-page/` — Status Page
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/maintenance/` — Maintenance
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/transactions/` — Transactions
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/pricing/` — Pricing
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/tax/` — Tax
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/sla/` — SLA
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/brand/` — Brand
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/localization/` — Localization
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/registration/` — Registration Settings
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/register` — Register Form

## Workspace Overview

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `/t/{slug}/w/{id}/` — Workspace Overview

## Workspace Leaf Pages

### Compute

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/vms/general` — General purpose VMs
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/vms/gpu` — GPU / accelerated VMs
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/vms/spot` — Spot / preemptible VMs
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/containers/kubernetes` — Managed Kubernetes
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/containers/registry` — Container registry
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/containers/instances` — Container instances
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/hosting/web` — Web app hosting
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/hosting/static` — Static site hosting
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/serverless/functions` — Functions as a service
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/serverless/edge` — Edge functions
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/batch/jobs` — Job scheduling
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/batch/hpc` — HPC / grid computing

### Storage

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/object/standard` — Standard tier
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/object/archive` — Archive / cold tier
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/object/datalake` — Data lake
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/block/ssd` — SSD high performance
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/block/hdd` — HDD standard
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/file/nfs` — NFS / SMB shares
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/file/distributed` — Distributed file systems
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/backup/snapshots` — Snapshot management
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/backup/dr` — Disaster recovery
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/backup/archive` — Long-term archiving

### Databases

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/relational/mysql-pg` — MySQL / PostgreSQL
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/relational/sql-server` — SQL Server
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/relational/cloud-native` — Cloud-native
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/nosql/document` — Document
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/nosql/keyvalue` — Key-value
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/nosql/wide-column` — Wide-column
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/nosql/graph` — Graph
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/memory/cache` — Cache
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/memory/session` — Session store
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/analytics/warehouse` — Data warehouse
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/analytics/timeseries` — Time-series
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/analytics/search` — Search & full-text

### Networking

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/vnet/vpc` — VPC / subnets
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/vnet/peering` — Peering & transit
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/vnet/nsg` — Network security groups
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/lb/app` — Application load balancer
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/lb/network` — Network load balancer
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/lb/global` — Global load balancing
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/dns/manage` — DNS management
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/dns/domains` — Domain registration
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/cdn/static` — Static asset delivery
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/cdn/video` — Video streaming
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/cdn/edge` — Edge caching
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/private/vpn` — VPN gateway
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/private/direct` — Direct connect
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/private/sdwan` — SD-WAN

### AI & Data Intelligence

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/training/distributed` — Distributed training
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/training/automl` — AutoML
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/training/finetuning` — Fine-tuning
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/hosting/inference` — Inference endpoints
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/hosting/batch` — Batch inference
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/hosting/edge` — Edge deployment
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/apis/vision` — Vision
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/apis/language` — Language / NLP
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/apis/speech` — Speech
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/apis/recommendation` — Recommendation
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/data/etl` — ETL / ELT pipelines
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/data/streaming` — Streaming ingestion
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/data/catalog` — Data catalog
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/mlops/experiments` — Experiment tracking
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/mlops/registry` — Model registry
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/mlops/monitoring` — Monitoring & drift detection

### Security & Identity

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/iam/sso` — SSO / federation
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/iam/mfa` — Multi-factor authentication
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/iam/rbac` — Role-based access control
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/data/enc-rest` — Encryption at rest
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/data/enc-transit` — Encryption in transit
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/data/kms` — Key management
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/network/waf` — Web application firewall
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/network/ddos` — DDoS protection
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/network/ids` — Intrusion detection
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/compliance/policy` — Policy management
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/compliance/audit` — Audit logging
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/compliance/regulatory` — Regulatory frameworks

### Developer Tools & DevOps

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/cicd/source` — Source control
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/cicd/build` — Build automation
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/cicd/deploy` — Deployment automation
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/iac/templates` — Templating
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/iac/config` — Config management
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/api/gateway` — API gateway
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/api/portal` — Developer portal
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/api/analytics` — API analytics
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/obs/logging` — Logging
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/obs/metrics` — Metrics & dashboards
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/obs/tracing` — Distributed tracing

### Application Services

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/messaging/queues` — Message queues
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/messaging/streaming` — Event streaming
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/messaging/servicebus` — Service bus
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/workflow/automation` — Workflow automation
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/workflow/scheduler` — Scheduler / cron jobs
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/media/encoding` — Video encoding
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/media/livestream` — Live streaming
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/media/storage` — Media storage

### Multi-Cloud & Hybrid

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/hybrid/onprem` — On-premises extension
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/hybrid/control` — Unified control plane
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/hybrid/edge` — Edge computing
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/mgmt/cost` — Cost management
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/mgmt/policy` — Policy governance
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/mgmt/workload` — Workload portability
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/migration/assessment` — Assessment & discovery
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/migration/data` — Data migration
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/migration/app` — Application migration

## Workspace Pro Feature Pages

- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/app-hosting/static-sites` — Websites & Apps
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/virtual-machines/linux-vms` — Cloud Servers
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/relational/postgresql` — Managed Databases
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/object-storage/s3-compatible` — Storage & Backups
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/communication/transactional-email` — Email & Messaging
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/cdn/pop-management` — Content Delivery
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/certificates/ssl-tls` — Security & SSL
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/monitoring/metrics` — Monitoring & Alerts
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/inference/llm-endpoints` — AI & Automation
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/serverless/functions` — Serverless Functions
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/dns/zones` — Domains & DNS
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `compute/kubernetes/managed-k8s` — Managed Kubernetes
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `storage/volumes/block-storage` — Block Storage
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `databases/nosql/document-db` — Document Database
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `networking/load-balancer/application-lb` — Application Load Balancer
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `ai/vector-search/embeddings` — Vector Search
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/iam/identity-platform` — Identity Platform
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `security/firewall/cloud-waf` — Cloud WAF
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `devops/cicd/pipelines` — CI/CD Pipelines
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `apps/queues/message-broker` — Message Broker
- [x] make the page and related feature production ready. implement database, background-jobs, logs, wiring, etc.: `multicloud/cost/cloud-optimizer` — Cost Optimizer

## Hardening & UX

- [x] Implement POST /api/user/password — password change handler with current-password verification, length check, and confirmation match; wired to the Security tab in user_settings.html
- [x] Wire New Workspace modal — POST /u/workspaces/create/ handler + JS fetch in user_workspaces.html with inline error display and page reload on success
- [x] Add slug/name validation to tenant create/update handlers — slug must match ^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$; name capped at 100 chars
- [x] Add client-side pagination to ticket list — 15 rows per page, prev/next + numbered buttons, info label; CSS added to dashboard.css
- [x] Make migrations idempotent — all CREATE TABLE / CREATE INDEX converted to IF NOT EXISTS; self-healing dirty-flag logic added to db/migrate.go
- [x] Add HTTP security headers middleware — X-Frame-Options, X-Content-Type-Options, X-XSS-Protection, Referrer-Policy, Permissions-Policy on all responses via securityHeaders() wrapper in main.go
- [x] Add login rate limiting — 5 failed attempts per IP triggers 1-minute lockout; in-memory per-IP tracker with pruning; wired into LoginHandler in handlers/auth.go
- [x] Add client-side pagination to Clients list — 15 tbody groups per page, prev/next + numbered buttons, info label; reuses tk-pagination CSS; added to tenant/clients_content.html
- [x] Add client-side pagination to Employees list — 15 rows per page, same pagination pattern as tickets/clients; added to tenant/employees_content.html
- [x] Add session inactivity timeout warning — banner appears at 25 min idle with 5-min countdown, auto-logout at 30 min; "Stay logged in" resets timers; CSS in additions.css, JS in dashboard.js
- [x] Add live search/filter to sysadmin tenants list — instant client-side filter by name, slug, or domain; search input in toolbar; no-results message; zero backend changes
- [x] Add HTTP request logging middleware — log every request: method, path, status code, and duration in ms; wraps the top-level mux in main.go alongside securityHeaders
- [x] Add forgot-password page — GET /forgot-password renders a form; POST sends a mock "reset link sent" confirmation; fixes 404 from the login page link
- [x] Add Content-Security-Policy header — extend securityHeaders() in main.go with a strict CSP: default-src 'self', script-src 'self' 'unsafe-inline', style-src 'self' 'unsafe-inline', img-src 'self' data:
- [x] Add client-side pagination to sysadmin tenants table — 10 rows per page, prev/next + numbered buttons, info label; reuses existing tk-pagination CSS pattern
- [x] Add email format validation — net/mail.ParseAddress used in clientsCreateHandler, clientsUpdateHandler, employeesCreateHandler, employeesUpdateHandler, and apiTenantRegisterPostHandler; rejects malformed addresses with inline form error
- [x] Add maxlength attributes to form inputs — name=100, email=254, company/role=100 on all client and employee create/edit modals in clients_content.html and employees_content.html
- [x] Add server-side input length limits — handlers reject name/company/role > 100 chars and email > 254 chars with FormError before DB write in both client and employee create/update handlers
- [x] Add client-side pagination + live search to abuse-reports list — 15 rows per page, search input wired alongside type/status filter dropdowns; pagination reuses tk-pagination CSS
- [x] Add client-side pagination + live search to refund-complaints list — 15 rows per page; wired existing search input and status filter to JS filter+paginate logic; reuses tk-pagination CSS
- [x] Strip Server and X-Powered-By response headers — securityHeaders() in main.go calls h.Del("Server") and h.Del("X-Powered-By") before passing to next handler to suppress information disclosure
- [x] Add password complexity validation to apiChangePasswordHandler — new password must contain at least one uppercase letter and one digit; validated in handlers/api_user_settings.go using hasUppercase/hasDigit helpers
- [x] Add server-side pagination to GET /api/sysadmin/tenants — accepts ?limit=N&offset=M (default 50, max 100); response includes total/limit/offset; parsePagination helper added to api_sysadmin_tenants.go
- [x] Add server-side pagination to GET /api/tenant/{slug}/tickets|clients|employees — accepts ?limit=N&offset=M (default 50, max 200); response includes total/limit/offset; uses shared parsePagination helper
- [x] Add composite DB index on sessions(user_id, expires_at DESC) — migration 0043_sessions_composite_idx.up.sql adds sessions_user_expires_idx WHERE revoked_at IS NULL for efficient per-user active-session lookups
- [x] Set Secure flag on session cookie when PLATFORM_ENV=production — prevents cookie transmission over plain HTTP in production deployments; handlers/session.go SetCookie call
- [x] Validate RedirectAfter in registration settings — reject any value that is not a safe relative path (must start with / and not contain ://, preventing open redirect attacks); handlers/api_tenant_registration.go
- [x] Add Content-Type enforcement middleware on /api/* POST/PUT/PATCH — reject requests whose Content-Type is not application/json with 415 Unsupported Media Type; added to main.go alongside securityHeaders
- [x] Revoke all other sessions on password change — after a successful password change, mark all sessions except the current one as revoked so other devices are forced to re-authenticate; handlers/api_user_settings.go + db layer
- [x] Add apiValidationError helper for field-level errors — returns {"errors":{"field":"message"}} shape; adopt in client/employee create handlers so the frontend can surface which specific field failed
- [x] Write auth events to audit_log — login success, login failure, logout, and password change are written to audit_log via db.InsertAuditEvent; fires async (goroutine) so it never blocks the response path
- [x] Rate limit on POST /forgot-password — 3 attempts per IP per 15 minutes; in-memory per-IP tracker with sliding window; returns error message on template without leaking whether the address is registered
- [x] Add HSTS header in production + X-Request-ID per request — securityHeaders() adds Strict-Transport-Security: max-age=63072000; includeSubDomains; preload when PLATFORM_ENV=production; requestLogger generates a crypto-random X-Request-ID per request, echoes it in the response header, and includes it in the log line for correlation
- [x] Add active sessions list + per-session revoke — GET /api/user/sessions returns all live sessions for the current user (handle, IP, UA, age); DELETE /api/user/sessions/{id} revokes a specific session (scoped to caller's user_id); Security tab in user_settings.html loads and renders the list dynamically with individual Revoke buttons
- [x] Split st-session-dot CSS — inactive sessions rendered with grey dot; current session gets green dot + glow via st-session-dot--active modifier class
- [x] Add global API rate limit middleware — 100 req/min per IP on all /api/* routes; returns 429 with Retry-After: 60 header; uses shared ipRateLimiter in handlers/rate_limit.go with background pruning goroutine
- [x] Add request body size limit middleware — limitRequestBody() in main.go caps POST/PUT/PATCH bodies to 1 MiB via http.MaxBytesReader to prevent memory exhaustion from large payloads
- [x] Upgrade session cookie SameSite from Lax to Strict — prevents session cookie from being sent on cross-site navigations, mitigating CSRF attack vectors; handlers/session.go
- [x] Add tenant slug format validation in tenantMiddleware — validates {slug} path parameter against ^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$ before DB lookup; rejects invalid slugs with 404; handlers/tenant_context.go
- [x] Fix pruneLoginEntries memory-leak bug + add background pruning goroutines — fixed incorrect condition (lockedAt.IsZero() + time.Since(zero)) that never pruned non-locked entries; added lastAt field to loginEntry; replaced forgotPasswordRateLimiter with shared ipRateLimiter; all rate limiters now auto-prune every 5 minutes via goroutines
- [x] Add GET /healthz endpoint — unauthenticated; returns JSON {"status":"ok","uptime_seconds":N}; checks DB connectivity; registered outside the auth middleware so load balancers can probe it
- [x] Enrich InsertAuditEvent with request_id and ip_address — extend db.InsertAuditEvent signature to accept requestID and ipAddr string; update all callers in handlers/auth.go to pass r.Header.Get("X-Request-ID") and clientIP(r); populate request_id and ip_address columns in audit_log
- [x] Add audit.lockout event to audit_log — when a login attempt is blocked by the lockout (5 failed attempts), call InsertAuditEvent with action "auth.lockout" and severity "warn" before returning the error response; handlers/auth.go LoginHandler
- [x] Add POST /api/user/sessions/revoke-all endpoint + wire UI — handler revokes all sessions for current user (all rows, not only others), writes auth.revoke_all audit event, returns 200; replace the "Sign out all sessions" anchor in user_settings.html with a JS button that calls the endpoint then redirects to /login
- [x] Wire API Keys section with live backend — add GET/POST/DELETE /api/user/api-keys endpoints backed by user_personal_api_keys table; convert sec-apikeys in user_settings.html to dynamic JS (load on tab open, inline create modal with name input, show key once in plaintext, revoke with confirmation)
- [x] Apply login rate limiting + audit events to POST /api/login — enforce same IP-based lockout (5 failures → 1-minute block) and write auth.login.success / auth.login.failure / auth.lockout audit events on the JSON API login endpoint, matching the HTML form handler; handlers/api_auth.go
- [x] Add Cache-Control: no-store middleware for all /api/* responses — apiNoCache() middleware in main.go sets Cache-Control: no-store, must-revalidate and Pragma: no-cache on every /api/* response to prevent intermediary proxies and browsers from caching sensitive JSON data; wired into the middleware chain
- [x] Rate limit POST /api/user/password per user (5 per hour) — add userIDRateLimiter type and passwordChangeRateLimiter instance (5 changes/hour per user ID) in handlers/rate_limit.go; apply in apiChangePasswordHandler with 429 + Retry-After: 3600 on breach
- [x] Sanitize raw error messages in sysadmin DB handlers — replace "Introspection failed: "+err.Error() with generic messages in apiSysadminDatabaseHandler, apiSysadminDatabaseERDHandler, and apiSysadminDatabaseTableHandler; real errors are logged server-side via log.Printf; prevents DB internals leaking to clients
- [x] Enforce max 10 active API keys per user in POST /api/user/api-keys — add db.CountPersonalAPIKeys() helper; check count before creation and return 422 if limit reached; prevents API key proliferation per account
- [x] Use constant-time comparison for all mock password checks — replace == operator with crypto/subtle.ConstantTimeCompare in LoginHandler (auth.go), apiLoginHandler (api_auth.go), and apiChangePasswordHandler (api_user_settings.go) to prevent timing side-channel attacks
- [x] Add Cross-Origin-Resource-Policy and Cross-Origin-Opener-Policy headers — extend securityHeaders() in main.go with Cross-Origin-Resource-Policy: same-origin and Cross-Origin-Opener-Policy: same-origin to prevent cross-origin data leaks and Spectre-like exfiltration
- [x] Validate tenant name length and domain format in POST /api/sysadmin/tenants — reject name > 100 chars; when domain is non-empty, reject > 253 chars and validate against RFC-1123 hostname regexp; domainRE compiled once at package level in api_sysadmin_tenants.go
- [x] Validate theme allowlist and timezone max length in PATCH /api/user/settings — theme must be one of light/dark/system (return 400 otherwise); timezone rejected if > 64 characters; prevents arbitrary string storage in user_preferences table
- [x] Rate limit POST /api/user/api-keys per user (20 per hour) — add apiKeyCreationRateLimiter instance in handlers/rate_limit.go; apply in apiUserAPIKeysCreateHandler with 429 + Retry-After: 3600 to prevent rapid-fire revoke+create cycling beyond the 10-key cap
- [x] Add HTTP server timeouts to prevent Slowloris — replace http.ListenAndServe with http.Server{ReadHeaderTimeout:5s, ReadTimeout:30s, WriteTimeout:60s, IdleTimeout:120s} in main.go to prevent slow-client connection exhaustion attacks
- [x] Sanitize URL path in request log to prevent log injection — strip \r and \n from r.URL.Path before log.Printf in requestLogger in main.go; prevents forged multi-line log entries via crafted URL paths
- [x] Write audit events for API key creation and revocation — call db.InsertAuditEvent async goroutine in apiUserAPIKeysCreateHandler and apiUserAPIKeyRevokeHandler after success; actions api_key.created / api_key.revoked; handlers/api_user_api_keys.go
- [x] Validate timezone against IANA tz database in PATCH /api/user/settings — call time.LoadLocation on the submitted timezone value; reject with 400 if it fails; ensures only valid IANA names are stored; handlers/api_user_settings.go
- [x] Enforce unique API key names per user — add db.PersonalAPIKeyNameExists() helper; check before create in apiUserAPIKeysCreateHandler and return 422 if a live key with same name already exists; prevents confusing duplicate-named keys
- [x] Rate limit POST /api/tenant/{slug}/register per IP — 10 attempts per IP per hour; registerRateLimiter added to handlers/rate_limit.go; returns 429 + Retry-After: 3600 to slow bot-driven account creation
- [x] Enforce password complexity in tenant registration — POST /api/tenant/{slug}/register rejects passwords missing an uppercase letter or digit, matching the same rules enforced on password change; handlers/api_tenant_register.go
- [x] Cap max workspaces per user at 10 — add db.CountUserWorkspaces() helper; check count before db.CreateUserWorkspace() in userWorkspacesCreateHandler and return 422 if limit reached; prevents resource exhaustion
- [x] Write workspace.created audit event on workspace creation — call db.InsertAuditEvent async (goroutine) after successful workspace insert in userWorkspacesCreateHandler; action "workspace.created" severity "info"; handlers/user_workspaces.go
- [x] Validate description length (<=500 chars) in workspace create — TrimSpace + len check before DB insert in userWorkspacesCreateHandler; returns 400 if exceeded; handlers/user_workspaces.go
- [x] Add X-DNS-Prefetch-Control: off header — extend securityHeaders() in main.go to suppress browser DNS prefetching on cross-origin links; prevents resource discovery via pre-fetch probes
- [x] Rate limit GET /api/user/sessions per user (30 req/min) — add sessionListRateLimiter (userIDRateLimiter) in handlers/rate_limit.go; apply in apiUserSessionsHandler with 429 + Retry-After: 60 to prevent rapid session enumeration
- [x] Enforce unique workspace name per user — db.WorkspaceNameExistsForUser() helper checks case-insensitive name match among the user's live workspaces; returns 422 if duplicate detected before DB insert in userWorkspacesCreateHandler
- [x] Prune expired sessions background job — db.PruneExpiredSessions() deletes sessions where expires_at or revoked_at is older than 7 days; goroutine in main.go runs it hourly to prevent unbounded sessions table growth
- [x] Prune old audit_log entries background job — db.PruneOldAuditLogs(retentionDays) deletes rows older than N days using parameterized interval; goroutine in main.go runs daily; configurable via AUDIT_LOG_RETENTION_DAYS env var (default 90)
- [x] Add audit events for sysadmin tenant create/delete — go db.InsertAuditEvent() calls with tenant.created (severity info) and tenant.deleted (severity warn) added to apiSysadminTenantsCreateHandler and apiSysadminTenantsDeleteHandler; handlers/api_sysadmin_tenants.go
- [x] Add DELETE /u/workspaces/{id} endpoint with ownership check and audit event — db.DeleteUserWorkspace() soft-deletes (sets deleted_at) only when caller holds owner role; returns 404 if not found/not owner; writes workspace.deleted audit event async; handlers/user_workspaces.go + db/workspaces.go
- [x] Add PATCH /api/sysadmin/tenants/{id} update handler — allows sysadmin to update tenant name and domain (slug is immutable); same validation as create (name ≤100 chars, RFC-1123 domain); writes tenant.updated audit event; registered in handlers/api.go + Swagger spec updated
- [x] Rate limit workspace creation per user (5/hour) — workspaceCreateRateLimiter (userIDRateLimiter, 5 per hour) added to handlers/rate_limit.go; applied in userWorkspacesCreateHandler before validation with 429 + Retry-After: 3600 on breach
- [x] Prune expired unaccepted workspace invitations background job — db.PruneExpiredInvitations() deletes user_workspace_invitations where accepted_at IS NULL and expires_at older than 7 days; daily goroutine in main.go alongside sessions and audit_log pruning

# Core Domain Abstractions & First Real Infrastructure Roadmap

## 1. Supply & Provider Layer (Infrastructure Driver Placeholders)
- [x] Define `SupplyPool` / `HostNode` domain types (`OwnershipType`: `OwnedBareMetal`, `Colocated`, `LeasedServer`, `CloudResellerAPI`, `NestedCluster`; capacity & supply cost attributes) in `providers/`
- [x] Define `DriverRegistry` / `Provisioner` interfaces in `providers/`:
  - [x] `ComputeDriver` (`ValidateConnection`, `InspectCapacity`, `BootstrapHost`, `CreateInstance`, `StartInstance`, `StopInstance`, `ResizeInstance`, `SnapshotInstance`, `DeleteInstance`, `ConsoleInfo`)
  - [x] `HostingDriver` (`CreateAccount`, `BindDomain`, `IssueSSL`, `SetQuota`, `SuspendAccount`, `DeleteAccount`)
  - [x] `NetworkStorageDriver` (`AllocateIP`, `ReleaseIP`, `BindIP`, `CreateVolume`, `AttachVolume`, `DetachVolume`, `DeleteVolume`)
- [x] Implement big-picture placeholder drivers in `providers/`:
  - [x] `QEMULibvirtDriver` (bare-metal / nested Linux KVM+QEMU host driver)
  - [x] `ProxmoxDriver` (Proxmox VE REST API driver placeholder)
  - [x] `VMwareVSphereDriver` (vCenter/ESXi SOAP/REST driver placeholder, including nested ESXi)
  - [x] `UpstreamResellerDriver` (Hetzner / BuyVM / DigitalOcean / AWS / OpenStack API reseller placeholder)
  - [x] `SharedWebDriver` & `PaaSAppDriver` (shared web hosting & PaaS buildpack/container placeholders)
  - [x] `IPAMDriver` & `BlockStorageDriver` (IP pool allocation & Ceph/ZFS/LVM/cloud block storage placeholders)

## 2. Unified Resource & Payable Item Layer (Catalog & Billable Tree Placeholders)
- [x] Define `ProductSKU` catalog model in `catalog/` (`Kind`: `vm`, `ipv4`, `ipv6_subnet`, `block_ssd`, `shared_hosting`, `k8s_node`, `paas_app`, `backup_gb`; `BillingModel`: `RecurringFlat`, `HourlyMetered`, `UsageCounter`, `OneTimeSetup`; `ProvisionTarget` placement rules)
- [x] Define unified `PayableResource` tree model (`Resource` node with `ParentResourceID` for child/attached payable items such as `ipv4`, `block_nvme`, `snapshot`, `ddos_prot`, plus `ProviderNodeID` and `ExternalRef`)
- [x] Define `SubscriptionLineItem`, `UsageMeter`, and `CatalogService` placeholder interfaces & implementations in `catalog/`

## 3. Cost-Centric Accounting Layer (COGS, Pricing Engine & Margin Ledger Placeholders)
- [x] Layer A — `SupplyCostCalculator` (`DirectPassThrough`, `AmortizedHardware`, `SharedOverhead` per host/provider/resource) in `accounting/`
- [x] Layer B — `PricingEngine` (`SellPrice = f(SupplyCost, Provider, Region, TenantTier, ClientPromo, SLA)` with `CostFloor` guard, provider/region multipliers, platform→tenant wholesale pricing, and tenant→client retail pricing) in `accounting/`
- [x] Layer C — `MarginLedger` (`CostEvent`, `RevenueEvent`, and `MarginView` computing `GrossMargin = Revenue - SupplyCost - PaymentFee - SLACredit` grouped by Cost Center) in `accounting/`

## 4. Clean Slate — Drop Dummy Data & Keep Only Real Seeders
- [x] Remove dummy demo seeders from `db/seed.go` (demo tenants `buyvm`/`abrnoc`/`hosting`, fake `system_logs`, fake `background_job_workers`, fake `background_jobs`, fake `replicas`, fake `billing_*` logs/webhooks) and keep only the System Super Administrator seeder (`tayyebimohammadreza@gmail.com`, configurable via env)
- [x] Purge hardcoded in-memory dummy fixture maps across `handlers/` (`buyvm`/`abrnoc`/`hosting` tenants, fake clients, fake VMs/resources, static mock tables) and wire tenant & workspace resolution directly to PostgreSQL
- [x] Fix Linux build exclusion bug on `db/maintenance_windows.go` (`_windows.go` Go OS suffix) so the full codebase compiles cleanly across Linux and Windows

## 5. First Real Infrastructure Case — Linux Nested VPS + Auto-Provisioning + Manual IP Pool
- [x] Add migration `0044_supply_hosts_and_ip_pool` and `0045_tenant_scoped_supply_hosts` creating tenant-isolated `supply_hosts` and `supply_host_ips` (with `FORCE ROW LEVEL SECURITY` `tenant_isolation` policies) and extending `resources` with `parent_resource_id`, `supply_host_id`, and `external_ref`
- [x] Implement real `LinuxNestedQEMUDriver` (`providers/linux_nested_qemu.go`) supporting SSH connectivity & hardware/KVM capability probing, automated QEMU/KVM + libvirt virtualization stack installation (`BootstrapHost`), and guest VM creation/deletion with static IP cloud-init configuration
- [x] Implement tenant-scoped `db/supply_hosts.go` for CRUD on supply hosts, auto-provisioning status/log updates, manual bulk/single IP inventory management, and transactional `FOR UPDATE SKIP LOCKED` IP allocation & release strictly isolated by `tenant_id`
- [x] Wire Tenant Resources Console (`/t/{slug}/resources/`) and REST API (`/api/tenant/{slug}/resources/*`) — reusing the Resource Tree (`inv-tree`), unified inventory table (`inv-table`), slide-in inspector (`inv-detail`), modals (`#onboardModal`, `#ipPoolModal`), and live capacity panel (`inv-capacity-panel`) — to onboard Linux Nested VPS servers, trigger SSH auto-provisioning, and manage manual assignable IP pools (removing `/s/infrastructure/`)
- [x] Wire customer VM purchase/creation (`CreateComputeVM` & compute API handlers) to schedule onto an active `ready` Linux Nested VPS host owned by that tenant, allocate a free IP from that server's manual IP pool as a child payable `resources` item, and invoke the driver

