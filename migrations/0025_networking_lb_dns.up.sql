-- 0025: workspace resources — networking load balancers (app, network, global)
--        and DNS (zones, domain registrations).
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── Application Load Balancers (Layer 7) ─────────────────────────────────────
-- Covers networking/lb/app
-- scheme: 'internet-facing' | 'internal'
-- protocol: 'HTTP' | 'HTTPS'
CREATE TABLE IF NOT EXISTS workspace_lb_app (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    scheme       TEXT NOT NULL DEFAULT 'internet-facing',
    protocol     TEXT NOT NULL DEFAULT 'HTTPS',
    port         INTEGER NOT NULL DEFAULT 443,
    targets      INTEGER NOT NULL DEFAULT 0,
    health_check TEXT NOT NULL DEFAULT '/health',
    ssl_arn      TEXT NOT NULL DEFAULT '',
    waf          BOOLEAN NOT NULL DEFAULT false,
    rps          INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_lb_app_resource_idx
    ON workspace_lb_app (resource_id);

-- ── Network Load Balancers (Layer 4) ──────────────────────────────────────────
-- Covers networking/lb/network
-- scheme: 'internet-facing' | 'internal'
-- protocol: 'TCP' | 'UDP' | 'TLS'
CREATE TABLE IF NOT EXISTS workspace_lb_network (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    scheme       TEXT NOT NULL DEFAULT 'internet-facing',
    protocol     TEXT NOT NULL DEFAULT 'TCP',
    port         INTEGER NOT NULL DEFAULT 443,
    targets      INTEGER NOT NULL DEFAULT 0,
    cross_zone   BOOLEAN NOT NULL DEFAULT true,
    static_ips   INTEGER NOT NULL DEFAULT 1,
    throughput   TEXT NOT NULL DEFAULT '0 Gbps',
    pps          INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS workspace_lb_network_resource_idx
    ON workspace_lb_network (resource_id);

-- ── Global Load Balancers (Anycast) ───────────────────────────────────────────
-- Covers networking/lb/global
-- routing: 'latency-based' | 'geo-proximity' | 'weighted'
-- failover: 'automatic' | 'manual'
CREATE TABLE IF NOT EXISTS workspace_lb_global (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    protocol     TEXT NOT NULL DEFAULT 'HTTPS',
    anycast_ip   TEXT NOT NULL DEFAULT '',
    regions      INTEGER NOT NULL DEFAULT 0,
    backends     INTEGER NOT NULL DEFAULT 0,
    routing      TEXT NOT NULL DEFAULT 'latency-based',
    failover     TEXT NOT NULL DEFAULT 'automatic',
    health_check TEXT NOT NULL DEFAULT '/health',
    rps          INTEGER NOT NULL DEFAULT 0,
    latency      TEXT NOT NULL DEFAULT '0ms'
);

CREATE INDEX IF NOT EXISTS workspace_lb_global_resource_idx
    ON workspace_lb_global (resource_id);

-- ── DNS Zones ─────────────────────────────────────────────────────────────────
-- Covers networking/dns/manage
-- zone_type: 'public' | 'private'
-- dnssec: 'enabled' | 'disabled'
CREATE TABLE IF NOT EXISTS workspace_dns_zones (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    zone_type    TEXT NOT NULL DEFAULT 'public',
    records      INTEGER NOT NULL DEFAULT 0,
    ttl          TEXT NOT NULL DEFAULT '300',
    dnssec       TEXT NOT NULL DEFAULT 'disabled',
    nameservers  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_dns_zones_resource_idx
    ON workspace_dns_zones (resource_id);

-- ── Domain Registrations ──────────────────────────────────────────────────────
-- Covers networking/dns/domains
CREATE TABLE IF NOT EXISTS workspace_domain_registrations (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    domain       TEXT NOT NULL DEFAULT '',
    tld          TEXT NOT NULL DEFAULT '.com',
    registrar    TEXT NOT NULL DEFAULT 'CloudRegistrar',
    expires_at   DATE,
    auto_renew   BOOLEAN NOT NULL DEFAULT true,
    privacy      BOOLEAN NOT NULL DEFAULT true,
    nameservers  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS workspace_domain_registrations_resource_idx
    ON workspace_domain_registrations (resource_id);
