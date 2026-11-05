# Hosting Platform

A multi-tenant cloud and infrastructure hosting control plane built with Go and PostgreSQL. It unifies supply-host drivers (starting with Linux Nested VPS auto-provisioning over SSH), assignable IP pool inventory, hierarchical payable resources, and cost-centric accounting across system administrators, reseller tenants, and client workspaces.

## Setup Guide

### Prerequisites

- **PostgreSQL 16+** (or Docker with Docker Compose)
- **Go 1.26+** (if building from source)

### 1. Start PostgreSQL and Initialize Roles

The platform uses three database roles (`platform_admin` for migrations and sysadmin operations, `platform_app` with Row-Level Security enforced, and `platform_readonly` for the SQL console).

Using Docker Compose:

```bash
cp .env.example .env
docker compose up -d
```

Or on an existing PostgreSQL instance, run [`scripts/db-init.sh`](scripts/db-init.sh) as your Postgres superuser with `POSTGRES_USER=platform_admin`, `POSTGRES_DB=platform`, `PG_APP_PASSWORD`, and `PG_READONLY_PASSWORD` set.

### 2. Configure Environment Variables

Edit `.env` (or export variables directly in your environment / systemd unit):

```bash
DATABASE_URL=postgres://platform_app:app_dev_password@localhost:5432/platform?sslmode=disable
DATABASE_URL_ADMIN=postgres://platform_admin:admin_dev_password@localhost:5432/platform?sslmode=disable
DATABASE_URL_READONLY=postgres://platform_readonly:readonly_dev_password@localhost:5432/platform?sslmode=disable

# Initial System Super Administrator seeded on first boot:
SYSADMIN_EMAIL=sysadmin@platform.io
SYSADMIN_NAME="System Administrator"
SYSADMIN_PASSWORD=sysadmin123
```

### 3. Build and Run

#### Option A — From a Released Archive

Download the archive for your architecture (`hosting-linux-amd64.tar.gz` or `hosting-linux-arm64.tar.gz`) from GitHub Releases and extract it into a working directory (keep `templates/` and `static/` next to the binary):

```bash
mkdir -p /opt/hosting
tar -xzf hosting-linux-amd64.tar.gz -C /opt/hosting
cd /opt/hosting
set -a && source .env && set +a
./hosting
```

#### Option B — From Source

```bash
go build -trimpath -ldflags="-s -w" -o hosting .
set -a && source .env && set +a
./hosting
```

On startup, embedded database migrations (`migrations/*.sql`) and the System Super Administrator seeder run automatically, and the HTTP server listens on `0.0.0.0:3000` (health check at `GET /healthz`).

### 4. Optional Automated Linux Deployment (GitHub Actions)

Configure the following repository secrets to enable automatic deployment to a target Linux server on release:

- `HOST_ADDRESS` — Target Linux server IP or hostname (optional `:port`, defaults to `22`)
- `HOST_USERNAME` — SSH username (e.g. `root` or a user with passwordless `sudo`)
- `HOST_KEY_BASE64` — Base64-encoded SSH private key (`base64 -w0 ~/.ssh/id_ed25519`)
- `HOST_PASSWORD` — SSH password (used as a fallback when `HOST_KEY_BASE64` is not set)
