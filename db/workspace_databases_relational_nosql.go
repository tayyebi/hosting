package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── MySQL / PostgreSQL ────────────────────────────────────────────────────

type RelationalDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Engine     string
	Version    string
	Size       string
	StorageGB  int
	Replicas   int
	Endpoint   string
	Encrypted  bool
	CreatedAt  time.Time
}

func ListRelationalDbs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]RelationalDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       rdb.engine, rdb.version, rdb.size,
		       rdb.storage_gb, rdb.replicas, rdb.endpoint, rdb.encrypted,
		       r.created_at
		FROM workspace_relational_dbs rdb
		JOIN resources r ON r.id = rdb.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListRelationalDbs: %w", err)
	}
	defer rows.Close()
	var out []RelationalDB
	for rows.Next() {
		var v RelationalDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.StorageGB, &v.Replicas, &v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateRelationalDB(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	storageGB, replicas int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "5432"
	if engine == "mysql" {
		port = "3306"
	}
	endpoint := name + ".db." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'relational_db',$3,'creating',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateRelationalDB resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_relational_dbs
		    (resource_id, engine, version, size, storage_gb, replicas, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		resourceID, engine, version, size, storageGB, replicas, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateRelationalDB db row: %w", err)
	}
	return resourceID, nil
}

// ── SQL Server ────────────────────────────────────────────────────────────

type SQLServerDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Edition    string
	Version    string
	Size       string
	StorageGB  int
	Replicas   int
	Endpoint   string
	Encrypted  bool
	CreatedAt  time.Time
}

func ListSQLServerDbs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SQLServerDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       sdb.edition, sdb.version, sdb.size,
		       sdb.storage_gb, sdb.replicas, sdb.endpoint, sdb.encrypted,
		       r.created_at
		FROM workspace_sqlserver_dbs sdb
		JOIN resources r ON r.id = sdb.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSQLServerDbs: %w", err)
	}
	defer rows.Close()
	var out []SQLServerDB
	for rows.Next() {
		var v SQLServerDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Edition, &v.Version, &v.Size,
			&v.StorageGB, &v.Replicas, &v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSQLServerDB(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, edition, version, size, region string,
	storageGB, replicas int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	endpoint := name + ".db." + region + ".example.com:1433"

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'sqlserver_db',$3,'creating',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSQLServerDB resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_sqlserver_dbs
		    (resource_id, edition, version, size, storage_gb, replicas, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		resourceID, edition, version, size, storageGB, replicas, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSQLServerDB db row: %w", err)
	}
	return resourceID, nil
}

// ── Cloud-native databases ────────────────────────────────────────────────

type CloudNativeDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Engine         string
	Version        string
	Size           string
	StorageUsedGB  int
	Replicas       int
	AutoScaling    bool
	ServerlessUnits string
	Endpoint       string
	Encrypted      bool
	CreatedAt      time.Time
}

func ListCloudNativeDbs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]CloudNativeDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       cdb.engine, cdb.version, cdb.size,
		       cdb.storage_used_gb, cdb.replicas,
		       cdb.auto_scaling, cdb.serverless_units,
		       cdb.endpoint, cdb.encrypted,
		       r.created_at
		FROM workspace_cloudnative_dbs cdb
		JOIN resources r ON r.id = cdb.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListCloudNativeDbs: %w", err)
	}
	defer rows.Close()
	var out []CloudNativeDB
	for rows.Next() {
		var v CloudNativeDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.StorageUsedGB, &v.Replicas,
			&v.AutoScaling, &v.ServerlessUnits,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateCloudNativeDB(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region, serverlessUnits string,
	replicas int, autoScaling, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "5432"
	if engine == "aurora-mysql-serverless" {
		port = "3306"
	} else if engine == "cloud-spanner" {
		port = "443"
	} else if engine == "cockroachdb" {
		port = "26257"
	}
	endpoint := name + ".cluster." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'cloudnative_db',$3,'creating',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateCloudNativeDB resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_cloudnative_dbs
		    (resource_id, engine, version, size, replicas, auto_scaling, serverless_units, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		resourceID, engine, version, size, replicas, autoScaling, serverlessUnits, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateCloudNativeDB db row: %w", err)
	}
	return resourceID, nil
}

// ── Document databases ────────────────────────────────────────────────────

type DocumentDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	Engine        string
	Version       string
	Size          string
	StorageGB     int
	Collections   int
	DocumentCount int64
	Endpoint      string
	Encrypted     bool
	CreatedAt     time.Time
}

func ListDocumentDbs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DocumentDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       ddb.engine, ddb.version, ddb.size,
		       ddb.storage_gb, ddb.collections, ddb.document_count,
		       ddb.endpoint, ddb.encrypted,
		       r.created_at
		FROM workspace_document_dbs ddb
		JOIN resources r ON r.id = ddb.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDocumentDbs: %w", err)
	}
	defer rows.Close()
	var out []DocumentDB
	for rows.Next() {
		var v DocumentDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.StorageGB, &v.Collections, &v.DocumentCount,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDocumentDB(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	storageGB int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "27017"
	if engine == "couchdb" {
		port = "5984"
	}
	prefix := "mongo"
	if len(engine) >= 5 {
		prefix = engine[:5]
	}
	endpoint := name + "." + prefix + "." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'document_db',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDocumentDB resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_document_dbs
		    (resource_id, engine, version, size, storage_gb, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, engine, version, size, storageGB, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDocumentDB db row: %w", err)
	}
	return resourceID, nil
}

// ── Key-value stores ──────────────────────────────────────────────────────

type KeyValueStore struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Engine      string
	Version     string
	Size        string
	MaxMemoryGB int
	KeyCount    int64
	Endpoint    string
	Encrypted   bool
	CreatedAt   time.Time
}

func ListKeyValueStores(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]KeyValueStore, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       kvs.engine, kvs.version, kvs.size,
		       kvs.max_memory_gb, kvs.key_count,
		       kvs.endpoint, kvs.encrypted,
		       r.created_at
		FROM workspace_keyvalue_stores kvs
		JOIN resources r ON r.id = kvs.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListKeyValueStores: %w", err)
	}
	defer rows.Close()
	var out []KeyValueStore
	for rows.Next() {
		var v KeyValueStore
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Version, &v.Size,
			&v.MaxMemoryGB, &v.KeyCount,
			&v.Endpoint, &v.Encrypted,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateKeyValueStore(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, version, size, region string,
	maxMemoryGB int, encrypted bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}
	port := "6379"
	switch engine {
	case "etcd":
		port = "2379"
	case "consul":
		port = "8500"
	}
	endpoint := name + ".kv." + region + ".example.com:" + port

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'keyvalue_store',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateKeyValueStore resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_keyvalue_stores
		    (resource_id, engine, version, size, max_memory_gb, endpoint, encrypted)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, engine, version, size, maxMemoryGB, endpoint, encrypted)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateKeyValueStore kv row: %w", err)
	}
	return resourceID, nil
}
