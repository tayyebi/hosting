package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Time-series databases ─────────────────────────────────────────────────────

type TimeSeriesDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Engine       string
	Version      string
	Size         string
	RetentionDays int
	StorageGB    int
	MetricsCount int64
	Endpoint     string
	Encrypted    bool
	CreatedAt    time.Time
}

func ListTimeSeriesDbs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]TimeSeriesDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       t.engine, t.version, t.size,
		       t.retention_days, t.storage_gb, t.metrics_count,
		       t.endpoint, t.encrypted,
		       r.created_at
		FROM workspace_timeseries_dbs t
		JOIN resources r ON r.id = t.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListTimeSeriesDbs: %w", err)
	}
	defer rows.Close()
	var out []TimeSeriesDB
	for rows.Next() {
		var v TimeSeriesDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.RetentionDays, &v.StorageGB, &v.MetricsCount,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateTimeSeriesDB(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	retentionDays int,
	encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "8086"
	switch engine {
	case "timescaledb":
		port = "5432"
	case "prometheus":
		port = "9090"
	}
	endpoint := name + ".ts." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'timeseries_db',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateTimeSeriesDB resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_timeseries_dbs
		    (resource_id, engine, version, size, retention_days, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, engine, version, size, retentionDays, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateTimeSeriesDB timeseries row: %w", err)
	}
	return resourceID, nil
}

// ── Search clusters ───────────────────────────────────────────────────────────

type SearchCluster struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Engine     string
	Version    string
	Size       string
	Nodes      int
	StorageGB  int
	Indices    int
	Endpoint   string
	Encrypted  bool
	CreatedAt  time.Time
}

func ListSearchClusters(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SearchCluster, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.engine, s.version, s.size,
		       s.nodes, s.storage_gb, s.indices,
		       s.endpoint, s.encrypted,
		       r.created_at
		FROM workspace_search_clusters s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSearchClusters: %w", err)
	}
	defer rows.Close()
	var out []SearchCluster
	for rows.Next() {
		var v SearchCluster
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.Nodes, &v.StorageGB, &v.Indices,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSearchCluster(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	nodes int,
	encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "9200"
	switch engine {
	case "meilisearch":
		port = "7700"
	case "typesense":
		port = "8108"
	}
	endpoint := name + ".search." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'search_cluster',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSearchCluster resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_search_clusters
		    (resource_id, engine, version, size, nodes, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, engine, version, size, nodes, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSearchCluster search row: %w", err)
	}
	return resourceID, nil
}

// ── VPCs ─────────────────────────────────────────────────────────────────────

type VPC struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	CIDR       string
	DNS        string
	Subnets    int
	Peerings   int
	Gateway    string
	Encrypted  bool
	CreatedAt  time.Time
}

func ListVPCs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]VPC, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       v.cidr, v.dns, v.subnets, v.peerings,
		       v.gateway, v.encrypted,
		       r.created_at
		FROM workspace_vpcs v
		JOIN resources r ON r.id = v.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListVPCs: %w", err)
	}
	defer rows.Close()
	var out []VPC
	for rows.Next() {
		var v VPC
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.CIDR, &v.DNS, &v.Subnets, &v.Peerings,
			&v.Gateway, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateVPC(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, cidr, dns, region string,
	encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	gateway := "igw-" + name

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'vpc',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateVPC resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_vpcs
		    (resource_id, cidr, dns, gateway, encrypted)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, cidr, dns, gateway, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateVPC vpc row: %w", err)
	}
	return resourceID, nil
}

// ── VPC Peerings ─────────────────────────────────────────────────────────────

type VPCPeering struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	LocalVPC    string
	RemoteVPC   string
	RemoteOwner string
	Direction   string
	Bandwidth   string
	Encrypted   bool
	CreatedAt   time.Time
}

func ListVPCPeerings(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]VPCPeering, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       p.local_vpc, p.remote_vpc, p.remote_owner,
		       p.direction, p.bandwidth, p.encrypted,
		       r.created_at
		FROM workspace_vpc_peerings p
		JOIN resources r ON r.id = p.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListVPCPeerings: %w", err)
	}
	defer rows.Close()
	var out []VPCPeering
	for rows.Next() {
		var v VPCPeering
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.LocalVPC, &v.RemoteVPC, &v.RemoteOwner,
			&v.Direction, &v.Bandwidth, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateVPCPeering(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, localVPC, remoteVPC, remoteOwner, direction, bandwidth, region string,
	encrypted bool,
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
		VALUES ($1,$2,'vpc_peering',$3,'pending-acceptance',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateVPCPeering resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_vpc_peerings
		    (resource_id, local_vpc, remote_vpc, remote_owner, direction, bandwidth, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, localVPC, remoteVPC, remoteOwner, direction, bandwidth, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateVPCPeering peering row: %w", err)
	}
	return resourceID, nil
}

// ── Network Security Groups ───────────────────────────────────────────────────

type NetworkSecurityGroup struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	VPCName        string
	InboundRules   int
	OutboundRules  int
	AssociatedNICs int
	DefaultAction  string
	Logging        bool
	CreatedAt      time.Time
}

func ListNetworkSecurityGroups(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]NetworkSecurityGroup, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       n.vpc_name, n.inbound_rules, n.outbound_rules,
		       n.associated_nics, n.default_action, n.logging,
		       r.created_at
		FROM workspace_network_security_groups n
		JOIN resources r ON r.id = n.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListNetworkSecurityGroups: %w", err)
	}
	defer rows.Close()
	var out []NetworkSecurityGroup
	for rows.Next() {
		var v NetworkSecurityGroup
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.VPCName, &v.InboundRules, &v.OutboundRules,
			&v.AssociatedNICs, &v.DefaultAction, &v.Logging,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateNetworkSecurityGroup(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, vpcName, region, defaultAction string,
	logging bool,
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
		VALUES ($1,$2,'nsg',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateNetworkSecurityGroup resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_network_security_groups
		    (resource_id, vpc_name, default_action, logging)
		VALUES ($1,$2,$3,$4)`,
		resourceID, vpcName, defaultAction, logging)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateNetworkSecurityGroup nsg row: %w", err)
	}
	return resourceID, nil
}
