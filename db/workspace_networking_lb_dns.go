package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Application Load Balancers (L7) ──────────────────────────────────────────

type LBApp struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Scheme      string
	Protocol    string
	Port        int
	Targets     int
	HealthCheck string
	SSLARN      string
	WAF         bool
	RPS         int
	CreatedAt   time.Time
}

func ListLBApps(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]LBApp, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       l.scheme, l.protocol, l.port,
		       l.targets, l.health_check, l.ssl_arn,
		       l.waf, l.rps,
		       r.created_at
		FROM workspace_lb_app l
		JOIN resources r ON r.id = l.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListLBApps: %w", err)
	}
	defer rows.Close()
	var out []LBApp
	for rows.Next() {
		var v LBApp
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Scheme, &v.Protocol, &v.Port,
			&v.Targets, &v.HealthCheck, &v.SSLARN,
			&v.WAF, &v.RPS,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateLBApp(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, scheme, protocol, healthCheck, sslARN, region string,
	port int,
	waf bool,
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
		VALUES ($1,$2,'lb_app',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateLBApp resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_lb_app
		    (resource_id, scheme, protocol, port, health_check, ssl_arn, waf)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, scheme, protocol, port, healthCheck, sslARN, waf)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateLBApp lb_app row: %w", err)
	}
	return resourceID, nil
}

// ── Network Load Balancers (L4) ───────────────────────────────────────────────

type LBNetwork struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Scheme     string
	Protocol   string
	Port       int
	Targets    int
	CrossZone  bool
	StaticIPs  int
	Throughput string
	PPS        int
	CreatedAt  time.Time
}

func ListLBNetworks(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]LBNetwork, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       l.scheme, l.protocol, l.port,
		       l.targets, l.cross_zone, l.static_ips,
		       l.throughput, l.pps,
		       r.created_at
		FROM workspace_lb_network l
		JOIN resources r ON r.id = l.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListLBNetworks: %w", err)
	}
	defer rows.Close()
	var out []LBNetwork
	for rows.Next() {
		var v LBNetwork
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Scheme, &v.Protocol, &v.Port,
			&v.Targets, &v.CrossZone, &v.StaticIPs,
			&v.Throughput, &v.PPS,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateLBNetwork(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, scheme, protocol, region string,
	port int,
	crossZone bool,
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
		VALUES ($1,$2,'lb_network',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateLBNetwork resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_lb_network
		    (resource_id, scheme, protocol, port, cross_zone)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, scheme, protocol, port, crossZone)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateLBNetwork lb_network row: %w", err)
	}
	return resourceID, nil
}

// ── Global Load Balancers (Anycast) ───────────────────────────────────────────

type LBGlobal struct {
	ResourceID  int64
	Name        string
	Status      string
	Protocol    string
	AnycastIP   string
	Regions     int
	Backends    int
	Routing     string
	Failover    string
	HealthCheck string
	RPS         int
	Latency     string
	CreatedAt   time.Time
}

func ListLBGlobals(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]LBGlobal, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       l.protocol, l.anycast_ip, l.regions, l.backends,
		       l.routing, l.failover, l.health_check,
		       l.rps, l.latency,
		       r.created_at
		FROM workspace_lb_global l
		JOIN resources r ON r.id = l.resource_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListLBGlobals: %w", err)
	}
	defer rows.Close()
	var out []LBGlobal
	for rows.Next() {
		var v LBGlobal
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status,
			&v.Protocol, &v.AnycastIP, &v.Regions, &v.Backends,
			&v.Routing, &v.Failover, &v.HealthCheck,
			&v.RPS, &v.Latency,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateLBGlobal(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, protocol, routing, failover, healthCheck string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'lb_global',$3,'provisioning',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateLBGlobal resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_lb_global
		    (resource_id, protocol, routing, failover, health_check)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, protocol, routing, failover, healthCheck)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateLBGlobal lb_global row: %w", err)
	}
	return resourceID, nil
}

// ── DNS Zones ─────────────────────────────────────────────────────────────────

type DNSZoneDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	ZoneType    string
	Records     int
	TTL         string
	DNSSEC      string
	Nameservers string
	CreatedAt   time.Time
}

func ListDNSZones(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DNSZoneDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, 'global'),
		       d.zone_type, d.records, d.ttl,
		       d.dnssec, d.nameservers,
		       r.created_at
		FROM workspace_dns_zones d
		JOIN resources r ON r.id = d.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDNSZones: %w", err)
	}
	defer rows.Close()
	var out []DNSZoneDB
	for rows.Next() {
		var v DNSZoneDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.ZoneType, &v.Records, &v.TTL,
			&v.DNSSEC, &v.Nameservers,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDNSZone(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, zoneType, ttl, dnssec, region string,
) (int64, error) {
	var regionID *int64
	if region != "" && region != "global" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	ns := "ns1.cloud.example.com, ns2.cloud.example.com"
	if zoneType == "private" {
		ns = "ns-priv-1." + region + ".example.com"
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'dns_zone',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDNSZone resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_dns_zones
		    (resource_id, zone_type, ttl, dnssec, nameservers)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, zoneType, ttl, dnssec, ns)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDNSZone dns_zone row: %w", err)
	}
	return resourceID, nil
}

// ── Domain Registrations ──────────────────────────────────────────────────────

type DomainRegistrationDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Domain      string
	TLD         string
	Registrar   string
	ExpiresAt   *time.Time
	AutoRenew   bool
	Privacy     bool
	Nameservers string
	CreatedAt   time.Time
}

func ListDomainRegistrations(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DomainRegistrationDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       d.domain, d.tld, d.registrar,
		       d.expires_at, d.auto_renew, d.privacy,
		       d.nameservers,
		       r.created_at
		FROM workspace_domain_registrations d
		JOIN resources r ON r.id = d.resource_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDomainRegistrations: %w", err)
	}
	defer rows.Close()
	var out []DomainRegistrationDB
	for rows.Next() {
		var v DomainRegistrationDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status,
			&v.Domain, &v.TLD, &v.Registrar,
			&v.ExpiresAt, &v.AutoRenew, &v.Privacy,
			&v.Nameservers,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDomainRegistration(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, domain, tld string,
	autoRenew, privacy bool,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'domain_registration',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDomainRegistration resource: %w", err)
	}
	ns := "ns1.cloud.example.com, ns2.cloud.example.com"
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_domain_registrations
		    (resource_id, domain, tld, auto_renew, privacy, nameservers)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, domain, tld, autoRenew, privacy, ns)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDomainRegistration domain row: %w", err)
	}
	return resourceID, nil
}
