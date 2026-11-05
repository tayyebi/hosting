package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── CDN: Static Asset Delivery ────────────────────────────────────────────────

type CDNStaticDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Domain      string
	Origin      string
	CachePolicy string
	SSL         bool
	Bandwidth   string
	Requests    string
	HitRate     string
	CreatedAt   time.Time
}

func ListCDNStatic(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]CDNStaticDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, 'global'),
		       c.domain, c.origin, c.cache_policy, c.ssl,
		       c.bandwidth, c.requests, c.hit_rate,
		       r.created_at
		FROM workspace_cdn_static c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListCDNStatic: %w", err)
	}
	defer rows.Close()
	var out []CDNStaticDB
	for rows.Next() {
		var v CDNStaticDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Domain, &v.Origin, &v.CachePolicy, &v.SSL,
			&v.Bandwidth, &v.Requests, &v.HitRate,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateCDNStatic(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, domain, origin, cachePolicy, region string,
	ssl bool,
) (int64, error) {
	var regionID *int64
	if region != "" && region != "global" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'cdn_static',$3,'deploying',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateCDNStatic resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_cdn_static
		    (resource_id, domain, origin, cache_policy, ssl)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, domain, origin, cachePolicy, ssl)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateCDNStatic cdn_static row: %w", err)
	}
	return resourceID, nil
}

// ── CDN: Video Streaming ──────────────────────────────────────────────────────

type CDNVideoDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Domain      string
	Origin      string
	Protocol    string
	Transcoding string
	DRM         bool
	Bandwidth   string
	Viewers     string
	CreatedAt   time.Time
}

func ListCDNVideo(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]CDNVideoDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, 'global'),
		       c.domain, c.origin, c.protocol,
		       c.transcoding, c.drm,
		       c.bandwidth, c.viewers,
		       r.created_at
		FROM workspace_cdn_video c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListCDNVideo: %w", err)
	}
	defer rows.Close()
	var out []CDNVideoDB
	for rows.Next() {
		var v CDNVideoDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Domain, &v.Origin, &v.Protocol,
			&v.Transcoding, &v.DRM,
			&v.Bandwidth, &v.Viewers,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateCDNVideo(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, domain, origin, protocol, transcoding, region string,
	drm bool,
) (int64, error) {
	var regionID *int64
	if region != "" && region != "global" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'cdn_video',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateCDNVideo resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_cdn_video
		    (resource_id, domain, origin, protocol, transcoding, drm)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, domain, origin, protocol, transcoding, drm)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateCDNVideo cdn_video row: %w", err)
	}
	return resourceID, nil
}

// ── CDN: Edge Caching ─────────────────────────────────────────────────────────

type CDNEdgeDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	POPs       int
	Rules      int
	CacheSize  string
	HitRate    string
	Bandwidth  string
	Purges     string
	WAF        bool
	CreatedAt  time.Time
}

func ListCDNEdge(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]CDNEdgeDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, 'global'),
		       c.pops, c.rules, c.cache_size,
		       c.hit_rate, c.bandwidth, c.purges, c.waf,
		       r.created_at
		FROM workspace_cdn_edge c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListCDNEdge: %w", err)
	}
	defer rows.Close()
	var out []CDNEdgeDB
	for rows.Next() {
		var v CDNEdgeDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.POPs, &v.Rules, &v.CacheSize,
			&v.HitRate, &v.Bandwidth, &v.Purges, &v.WAF,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateCDNEdge(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, region string,
	waf bool,
) (int64, error) {
	var regionID *int64
	if region != "" && region != "global" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	pops := 84
	switch region {
	case "eu":
		pops = 28
	case "us":
		pops = 32
	case "apac":
		pops = 18
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'cdn_edge',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateCDNEdge resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_cdn_edge
		    (resource_id, pops, waf)
		VALUES ($1,$2,$3)`,
		resourceID, pops, waf)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateCDNEdge cdn_edge row: %w", err)
	}
	return resourceID, nil
}

// ── VPN Gateways ──────────────────────────────────────────────────────────────

type VPNGatewayDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	VPNType    string
	PublicIP   string
	RemoteIP   string
	Tunnels    int
	Throughput string
	Encryption string
	IKEVersion string
	Uptime     string
	CreatedAt  time.Time
}

func ListVPNGateways(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]VPNGatewayDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       v.vpn_type, v.public_ip, v.remote_ip,
		       v.tunnels, v.throughput,
		       v.encryption, v.ike_version, v.uptime,
		       r.created_at
		FROM workspace_vpn_gateways v
		JOIN resources r ON r.id = v.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListVPNGateways: %w", err)
	}
	defer rows.Close()
	var out []VPNGatewayDB
	for rows.Next() {
		var v VPNGatewayDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.VPNType, &v.PublicIP, &v.RemoteIP,
			&v.Tunnels, &v.Throughput,
			&v.Encryption, &v.IKEVersion, &v.Uptime,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateVPNGateway(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, vpnType, region, remoteIP, encryption, ikeVersion string,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'vpn_gateway',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateVPNGateway resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_vpn_gateways
		    (resource_id, vpn_type, remote_ip, encryption, ike_version)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, vpnType, remoteIP, encryption, ikeVersion)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateVPNGateway vpn_gateways row: %w", err)
	}
	return resourceID, nil
}

// ── Direct Connect ────────────────────────────────────────────────────────────

type DirectConnectDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Location   string
	Port       string
	Bandwidth  string
	VLAN       int
	BGPASN     string
	Redundancy string
	Latency    string
	MACsec     bool
	CreatedAt  time.Time
}

func ListDirectConnects(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DirectConnectDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       d.location, d.port, d.bandwidth,
		       d.vlan, d.bgp_asn, d.redundancy,
		       d.latency, d.macsec,
		       r.created_at
		FROM workspace_direct_connects d
		JOIN resources r ON r.id = d.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDirectConnects: %w", err)
	}
	defer rows.Close()
	var out []DirectConnectDB
	for rows.Next() {
		var v DirectConnectDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Location, &v.Port, &v.Bandwidth,
			&v.VLAN, &v.BGPASN, &v.Redundancy,
			&v.Latency, &v.MACsec,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDirectConnect(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, region, location, port, bandwidth, bgpASN, redundancy string,
	macsec bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'direct_connect',$3,'pending-approval',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDirectConnect resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_direct_connects
		    (resource_id, location, port, bandwidth, bgp_asn, redundancy, macsec)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, location, port, bandwidth, bgpASN, redundancy, macsec)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDirectConnect direct_connects row: %w", err)
	}
	return resourceID, nil
}
