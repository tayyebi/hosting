# Hosting Platform

A multi-tenant cloud and infrastructure hosting control plane built with Go and PostgreSQL. It unifies supply-host drivers (starting with Linux Nested VPS auto-provisioning over SSH), assignable IP pool inventory, hierarchical payable resources, and cost-centric accounting across system administrators, reseller tenants, and client workspaces.

## Setup Guide

### Quick Start (Docker Compose)

Pull the repository and bring up the full stack:

```bash
git pull
docker-compose up
```

*(Or `docker compose up -d` to run in the background.)*

- **Database isolation**: PostgreSQL 16 runs passwordless (`trust`) inside an `internal: true` Docker network (`internal`) with no ports exposed outside Docker, initializing the three database roles (`platform_admin`, `platform_app`, `platform_readonly`).
- **Web control plane**: The application runs embedded SQL migrations (`migrations/*.sql`), seeds the initial System Super Administrator, and exposes only port **`8021`** (`http://localhost:8021`, health check at `GET /healthz`).
- **Optional overrides**: Set `APP_PORT`, `SYSADMIN_EMAIL`, `SYSADMIN_NAME`, or `SYSADMIN_PASSWORD` in `.env` if you want to customize defaults.

### Manual Setup (Without Docker)

1. Initialize the three roles on an existing PostgreSQL 16+ instance using [`scripts/db-init.sh`](scripts/db-init.sh) (`POSTGRES_USER=platform_admin`, `POSTGRES_DB=platform`, and optional `PG_APP_PASSWORD` / `PG_READONLY_PASSWORD`).
2. Copy `.env.example` to `.env` and adjust `DATABASE_URL`, `DATABASE_URL_ADMIN`, `DATABASE_URL_READONLY`, and `PORT` (defaults to `8021`).
3. Run from a GitHub Release archive (`hosting-linux-amd64.tar.gz` or `hosting-linux-arm64.tar.gz`, keeping `templates/` and `static/` next to the binary) or build from source (Go 1.26+):

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
