# Hosting Platform

A multi-tenant cloud and infrastructure hosting control plane built with Go and PostgreSQL. It unifies supply-host drivers (starting with Linux Nested VPS auto-provisioning over SSH), assignable IP pool inventory, hierarchical payable resources, and cost-centric accounting across system administrators, reseller tenants, and client workspaces.

## Setup Guide

### Quick Start (Docker Compose)

With Docker and Docker Compose installed, pull the repository and bring up the full stack (PostgreSQL 16 + control plane):

```bash
git pull
docker-compose up
```

*(Or `docker compose up -d` to run in the background.)*

On startup, PostgreSQL initializes the three database roles (`platform_admin`, `platform_app`, `platform_readonly`), the application container runs embedded SQL migrations (`migrations/*.sql`), seeds the initial System Super Administrator, and listens on `http://localhost:3000` (health check at `GET /healthz`).

Optional overrides can be configured by copying `.env.example` to `.env` before starting (`PG_ADMIN_PASSWORD`, `PG_APP_PASSWORD`, `PG_READONLY_PASSWORD`, `SYSADMIN_EMAIL`, `SYSADMIN_NAME`, `SYSADMIN_PASSWORD`, `APP_PORT`).

### Manual Setup (Without Docker)

#### 1. Initialize PostgreSQL Roles

On an existing PostgreSQL 16+ instance, run [`scripts/db-init.sh`](scripts/db-init.sh) as your Postgres superuser with `POSTGRES_USER=platform_admin`, `POSTGRES_DB=platform`, `PG_APP_PASSWORD`, and `PG_READONLY_PASSWORD` set.

#### 2. Configure Environment Variables

Copy `.env.example` to `.env` and adjust credentials for your PostgreSQL host:

```bash
DATABASE_URL=postgres://platform_app:app_dev_password@localhost:5432/platform?sslmode=disable
DATABASE_URL_ADMIN=postgres://platform_admin:admin_dev_password@localhost:5432/platform?sslmode=disable
DATABASE_URL_READONLY=postgres://platform_readonly:readonly_dev_password@localhost:5432/platform?sslmode=disable

SYSADMIN_EMAIL=sysadmin@platform.io
SYSADMIN_NAME="System Administrator"
SYSADMIN_PASSWORD=sysadmin123
```

#### 3. Run From a Released Archive or Source

From a GitHub Release archive (`hosting-linux-amd64.tar.gz` or `hosting-linux-arm64.tar.gz`, keeping `templates/` and `static/` next to the binary):

```bash
mkdir -p /opt/hosting
tar -xzf hosting-linux-amd64.tar.gz -C /opt/hosting
cd /opt/hosting
set -a && source .env && set +a
./hosting
```

Or build directly from source (Go 1.26+):

```bash
go build -trimpath -ldflags="-s -w" -o hosting .
set -a && source .env && set +a
./hosting
```

### Optional Automated Linux Deployment (GitHub Actions)

Configure the following repository secrets to enable automatic deployment to a target Linux server on release:

- `HOST_ADDRESS` — Target Linux server IP or hostname (optional `:port`, defaults to `22`)
- `HOST_USERNAME` — SSH username (e.g. `root` or a user with passwordless `sudo`)
- `HOST_KEY_BASE64` — Base64-encoded SSH private key (`base64 -w0 ~/.ssh/id_ed25519`)
- `HOST_PASSWORD` — SSH password (used as a fallback when `HOST_KEY_BASE64` is not set)
