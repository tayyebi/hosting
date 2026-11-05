package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Wide-column databases ─────────────────────────────────────────────────────

type WideColumnDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Engine     string
	Version    string
	Nodes      int
	Keyspaces  int
	StorageGB  int
	Endpoint   string
	Encrypted  bool
	CreatedAt  time.Time
}

func ListWideColumnDbs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]WideColumnDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       wcd.engine, wcd.version, wcd.nodes, wcd.keyspaces,
		       wcd.storage_gb, wcd.endpoint, wcd.encrypted,
		       r.created_at
		FROM workspace_widecolumn_dbs wcd
		JOIN resources r ON r.id = wcd.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListWideColumnDbs: %w", err)
	}
	defer rows.Close()
	var out []WideColumnDB
	for rows.Next() {
		var v WideColumnDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Nodes, &v.Keyspaces,
			&v.StorageGB, &v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateWideColumnDB(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, region string,
	nodes int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "9042"
	if engine == "hbase" {
		port = "16010"
	}
	endpoint := name + ".wc." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'widecolumn_db',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateWideColumnDB resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_widecolumn_dbs
		    (resource_id, engine, version, nodes, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, engine, version, nodes, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateWideColumnDB wc row: %w", err)
	}
	return resourceID, nil
}

// ── Graph databases ───────────────────────────────────────────────────────────

type GraphDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Engine     string
	Version    string
	Size       string
	StorageGB  int
	NodeCount  int64
	EdgeCount  int64
	Endpoint   string
	Encrypted  bool
	CreatedAt  time.Time
}

func ListGraphDbs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]GraphDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       gdb.engine, gdb.version, gdb.size,
		       gdb.storage_gb, gdb.node_count, gdb.edge_count,
		       gdb.endpoint, gdb.encrypted,
		       r.created_at
		FROM workspace_graph_dbs gdb
		JOIN resources r ON r.id = gdb.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListGraphDbs: %w", err)
	}
	defer rows.Close()
	var out []GraphDB
	for rows.Next() {
		var v GraphDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.StorageGB, &v.NodeCount, &v.EdgeCount,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateGraphDB(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "7687"
	if engine == "neptune" {
		port = "8182"
	}
	endpoint := name + ".graph." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'graph_db',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateGraphDB resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_graph_dbs
		    (resource_id, engine, version, size, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, engine, version, size, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateGraphDB graph row: %w", err)
	}
	return resourceID, nil
}

// ── In-memory caches ──────────────────────────────────────────────────────────

type MemoryCache struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Engine         string
	Version        string
	Size           string
	Nodes          int
	MaxMemoryGB    int
	EvictionPolicy string
	Endpoint       string
	Encrypted      bool
	CreatedAt      time.Time
}

func ListMemoryCaches(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MemoryCache, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       mc.engine, mc.version, mc.size, mc.nodes,
		       mc.max_memory_gb, mc.eviction_policy,
		       mc.endpoint, mc.encrypted,
		       r.created_at
		FROM workspace_memory_caches mc
		JOIN resources r ON r.id = mc.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMemoryCaches: %w", err)
	}
	defer rows.Close()
	var out []MemoryCache
	for rows.Next() {
		var v MemoryCache
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size, &v.Nodes,
			&v.MaxMemoryGB, &v.EvictionPolicy,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMemoryCache(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region, evictionPolicy string,
	nodes, maxMemoryGB int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "6379"
	if engine == "memcached" {
		port = "11211"
	}
	endpoint := name + ".cache." + region + ".example.com:" + port
	if evictionPolicy == "" {
		evictionPolicy = "allkeys-lru"
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'memory_cache',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMemoryCache resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_memory_caches
		    (resource_id, engine, version, size, nodes, max_memory_gb, eviction_policy, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		resourceID, engine, version, size, nodes, maxMemoryGB, evictionPolicy, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMemoryCache cache row: %w", err)
	}
	return resourceID, nil
}

// ── In-memory session stores ──────────────────────────────────────────────────

type MemorySession struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Engine      string
	Version     string
	Size        string
	Nodes       int
	MaxMemoryGB int
	TTLSeconds  int
	Endpoint    string
	Encrypted   bool
	CreatedAt   time.Time
}

func ListMemorySessions(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MemorySession, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       ms.engine, ms.version, ms.size, ms.nodes,
		       ms.max_memory_gb, ms.ttl_seconds,
		       ms.endpoint, ms.encrypted,
		       r.created_at
		FROM workspace_memory_sessions ms
		JOIN resources r ON r.id = ms.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMemorySessions: %w", err)
	}
	defer rows.Close()
	var out []MemorySession
	for rows.Next() {
		var v MemorySession
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size, &v.Nodes,
			&v.MaxMemoryGB, &v.TTLSeconds,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMemorySession(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	nodes, maxMemoryGB, ttlSeconds int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "6379"
	if engine == "memcached" {
		port = "11211"
	}
	endpoint := name + ".session." + region + ".example.com:" + port
	if ttlSeconds <= 0 {
		ttlSeconds = 1800
	}

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'memory_session',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMemorySession resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_memory_sessions
		    (resource_id, engine, version, size, nodes, max_memory_gb, ttl_seconds, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		resourceID, engine, version, size, nodes, maxMemoryGB, ttlSeconds, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMemorySession session row: %w", err)
	}
	return resourceID, nil
}

// ── Data warehouses ───────────────────────────────────────────────────────────

type DataWarehouse struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Engine     string
	Version    string
	Size       string
	StorageTB  int
	Tables     int
	Endpoint   string
	Encrypted  bool
	CreatedAt  time.Time
}

func ListDataWarehouses(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DataWarehouse, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       dw.engine, dw.version, dw.size,
		       dw.storage_tb, dw.tables,
		       dw.endpoint, dw.encrypted,
		       r.created_at
		FROM workspace_warehouse_dbs dw
		JOIN resources r ON r.id = dw.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDataWarehouses: %w", err)
	}
	defer rows.Close()
	var out []DataWarehouse
	for rows.Next() {
		var v DataWarehouse
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.StorageTB, &v.Tables,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDataWarehouse(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "8443"
	switch engine {
	case "snowflake":
		port = "443"
	case "redshift":
		port = "5439"
	case "bigquery":
		port = "443"
	}
	endpoint := name + ".warehouse." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'warehouse_db',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDataWarehouse resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_warehouse_dbs
		    (resource_id, engine, version, size, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, engine, version, size, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDataWarehouse warehouse row: %w", err)
	}
	return resourceID, nil
}
