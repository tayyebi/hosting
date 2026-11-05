-- 0026: workspace resources — networking CDN (static, video, edge)
--        and private connectivity (VPN gateway, direct connect).
-- All tables follow the CTI pattern: resources + specialised table keyed by resource_id.

-- ── CDN: Static Asset Delivery ────────────────────────────────────────────────
-- Covers networking/cdn/static
-- cache_policy: 'aggressive' | 'standard' | 'no-cache' | 'api'
CREATE TABLE IF NOT EXISTS workspace_cdn_static (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    domain       TEXT NOT NULL DEFAULT '',
    origin       TEXT NOT NULL DEFAULT '',
    cache_policy TEXT NOT NULL DEFAULT 'standard',
    ssl          BOOLEAN NOT NULL DEFAULT true,
    bandwidth    TEXT NOT NULL DEFAULT '0 B/mo',
    requests     TEXT NOT NULL DEFAULT '0/mo',
    hit_rate     TEXT NOT NULL DEFAULT '—'
);

CREATE INDEX IF NOT EXISTS workspace_cdn_static_resource_idx
    ON workspace_cdn_static (resource_id);

-- ── CDN: Video Streaming ──────────────────────────────────────────────────────
-- Covers networking/cdn/video
-- protocol: 'HLS' | 'DASH' | 'HLS + DASH' | 'MP3 128k'
-- transcoding: 'adaptive' | 'realtime' | 'none'
CREATE TABLE IF NOT EXISTS workspace_cdn_video (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    domain       TEXT NOT NULL DEFAULT '',
    origin       TEXT NOT NULL DEFAULT '',
    protocol     TEXT NOT NULL DEFAULT 'HLS',
    transcoding  TEXT NOT NULL DEFAULT 'adaptive',
    drm          BOOLEAN NOT NULL DEFAULT false,
    bandwidth    TEXT NOT NULL DEFAULT '0 B/mo',
    viewers      TEXT NOT NULL DEFAULT '0/mo'
);

CREATE INDEX IF NOT EXISTS workspace_cdn_video_resource_idx
    ON workspace_cdn_video (resource_id);

-- ── CDN: Edge Caching ─────────────────────────────────────────────────────────
-- Covers networking/cdn/edge
CREATE TABLE IF NOT EXISTS workspace_cdn_edge (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    pops         INTEGER NOT NULL DEFAULT 0,
    rules        INTEGER NOT NULL DEFAULT 0,
    cache_size   TEXT NOT NULL DEFAULT '0 B',
    hit_rate     TEXT NOT NULL DEFAULT '—',
    bandwidth    TEXT NOT NULL DEFAULT '0 B/mo',
    purges       TEXT NOT NULL DEFAULT '0/day',
    waf          BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS workspace_cdn_edge_resource_idx
    ON workspace_cdn_edge (resource_id);

-- ── VPN Gateways ──────────────────────────────────────────────────────────────
-- Covers networking/private/vpn
-- vpn_type: 'site-to-site' | 'client'
CREATE TABLE IF NOT EXISTS workspace_vpn_gateways (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    vpn_type     TEXT NOT NULL DEFAULT 'site-to-site',
    public_ip    TEXT NOT NULL DEFAULT 'pending',
    remote_ip    TEXT NOT NULL DEFAULT '—',
    tunnels      INTEGER NOT NULL DEFAULT 0,
    throughput   TEXT NOT NULL DEFAULT '0 Mbps',
    encryption   TEXT NOT NULL DEFAULT 'AES-256-GCM',
    ike_version  TEXT NOT NULL DEFAULT 'IKEv2',
    uptime       TEXT NOT NULL DEFAULT '—'
);

CREATE INDEX IF NOT EXISTS workspace_vpn_gateways_resource_idx
    ON workspace_vpn_gateways (resource_id);

-- ── Direct Connect ────────────────────────────────────────────────────────────
-- Covers networking/private/direct
-- redundancy: 'single' | 'dual'
CREATE TABLE IF NOT EXISTS workspace_direct_connects (
    resource_id  BIGINT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    location     TEXT NOT NULL DEFAULT '',
    port         TEXT NOT NULL DEFAULT '1 GbE',
    bandwidth    TEXT NOT NULL DEFAULT '1 Gbps',
    vlan         INTEGER NOT NULL DEFAULT 0,
    bgp_asn      TEXT NOT NULL DEFAULT '65000',
    redundancy   TEXT NOT NULL DEFAULT 'single',
    latency      TEXT NOT NULL DEFAULT '—',
    macsec       BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS workspace_direct_connects_resource_idx
    ON workspace_direct_connects (resource_id);
