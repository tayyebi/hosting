package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Application Services: Media — Video Encoding ──────────────────────────────

type AppsMediaEncodingDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	InputFormat  string
	Profile      string
	Duration     string
	Progress     int
	OutputSize   string
	ETA          string
	EncodeStatus string
	CreatedAt    time.Time
}

func ListAppsMediaEncodings(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsMediaEncodingDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       e.input_format, e.profile, e.duration,
		       e.progress, e.output_size, e.eta, e.encode_status,
		       r.created_at
		FROM workspace_apps_media_encoding e
		JOIN resources r ON r.id = e.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsMediaEncodings: %w", err)
	}
	defer rows.Close()
	var out []AppsMediaEncodingDB
	for rows.Next() {
		var v AppsMediaEncodingDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.InputFormat, &v.Profile, &v.Duration,
			&v.Progress, &v.OutputSize, &v.ETA, &v.EncodeStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsMediaEncoding(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, inputFormat, profile string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_media_encoding',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsMediaEncoding resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_media_encoding (resource_id, input_format, profile)
		VALUES ($1,$2,$3)`,
		resourceID, inputFormat, profile)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsMediaEncoding row: %w", err)
	}
	return resourceID, nil
}

// ── Application Services: Media — Live Streaming ──────────────────────────────

type AppsMediaLivestreamDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	SourceProtocol string
	OutputProtocol string
	Profile        string
	Viewers        int
	Bitrate        string
	Latency        string
	StreamStatus   string
	CreatedAt      time.Time
}

func ListAppsMediaLivestreams(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsMediaLivestreamDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       l.source_protocol, l.output_protocol, l.profile,
		       l.viewers, l.bitrate, l.latency, l.stream_status,
		       r.created_at
		FROM workspace_apps_media_livestream l
		JOIN resources r ON r.id = l.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsMediaLivestreams: %w", err)
	}
	defer rows.Close()
	var out []AppsMediaLivestreamDB
	for rows.Next() {
		var v AppsMediaLivestreamDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.SourceProtocol, &v.OutputProtocol, &v.Profile,
			&v.Viewers, &v.Bitrate, &v.Latency, &v.StreamStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsMediaLivestream(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, sourceProtocol, outputProtocol, profile string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_media_livestream',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsMediaLivestream resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_media_livestream (resource_id, source_protocol, output_protocol, profile)
		VALUES ($1,$2,$3,$4)`,
		resourceID, sourceProtocol, outputProtocol, profile)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsMediaLivestream row: %w", err)
	}
	return resourceID, nil
}

// ── Application Services: Media — Media Storage ───────────────────────────────

type AppsMediaStorageDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	AssetType   string
	Size        string
	Codec       string
	Plays7d     int
	Bandwidth   string
	CDNHit      string
	AssetStatus string
	CreatedAt   time.Time
}

func ListAppsMediaStorageAssets(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AppsMediaStorageDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.asset_type, s.size, s.codec,
		       s.plays_7d, s.bandwidth, s.cdn_hit, s.asset_status,
		       r.created_at
		FROM workspace_apps_media_storage s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAppsMediaStorageAssets: %w", err)
	}
	defer rows.Close()
	var out []AppsMediaStorageDB
	for rows.Next() {
		var v AppsMediaStorageDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.AssetType, &v.Size, &v.Codec,
			&v.Plays7d, &v.Bandwidth, &v.CDNHit, &v.AssetStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAppsMediaStorageAsset(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, assetType, codec string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'apps_media_storage',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAppsMediaStorageAsset resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_apps_media_storage (resource_id, asset_type, codec)
		VALUES ($1,$2,$3)`,
		resourceID, assetType, codec)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAppsMediaStorageAsset row: %w", err)
	}
	return resourceID, nil
}

// ── Multi-Cloud & Hybrid: On-Premises Extension ───────────────────────────────

type MulticloudOnpremSiteDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Location   string
	Hosts      int
	VCPUs      int
	Workloads  int
	LinkType   string
	Bandwidth  string
	Latency    string
	SiteStatus string
	CreatedAt  time.Time
}

func ListMulticloudOnpremSites(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudOnpremSiteDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.location, s.hosts, s.vcpus,
		       s.workloads, s.link_type, s.bandwidth, s.latency, s.site_status,
		       r.created_at
		FROM workspace_multicloud_onprem_sites s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudOnpremSites: %w", err)
	}
	defer rows.Close()
	var out []MulticloudOnpremSiteDB
	for rows.Next() {
		var v MulticloudOnpremSiteDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Location, &v.Hosts, &v.VCPUs,
			&v.Workloads, &v.LinkType, &v.Bandwidth, &v.Latency, &v.SiteStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudOnpremSite(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, location, linkType, bandwidth string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_onprem_site',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudOnpremSite resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_onprem_sites (resource_id, location, link_type, bandwidth)
		VALUES ($1,$2,$3,$4)`,
		resourceID, location, linkType, bandwidth)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudOnpremSite row: %w", err)
	}
	return resourceID, nil
}

// ── Multi-Cloud & Hybrid: Unified Control Plane ───────────────────────────────

type MulticloudControlClusterDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	Provider      string
	ClusterRegion string
	Nodes         int
	Pods          int
	Policies      string
	ConfigDrift   string
	ClusterStatus string
	CreatedAt     time.Time
}

func ListMulticloudControlClusters(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudControlClusterDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       c.provider, c.region, c.nodes,
		       c.pods, c.policies, c.config_drift, c.cluster_status,
		       r.created_at
		FROM workspace_multicloud_control_clusters c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudControlClusters: %w", err)
	}
	defer rows.Close()
	var out []MulticloudControlClusterDB
	for rows.Next() {
		var v MulticloudControlClusterDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Provider, &v.ClusterRegion, &v.Nodes,
			&v.Pods, &v.Policies, &v.ConfigDrift, &v.ClusterStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudControlCluster(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, provider, region string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_control_cluster',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudControlCluster resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_control_clusters (resource_id, provider, region)
		VALUES ($1,$2,$3)`,
		resourceID, provider, region)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudControlCluster row: %w", err)
	}
	return resourceID, nil
}
