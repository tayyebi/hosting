package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ObjectBucket is a row from workspace_object_buckets joined with resources.
type ObjectBucket struct {
	ResourceID        int64
	Name              string
	Status            string
	Region            string
	StorageClass      string
	SizeGB            float64
	ObjectCount       int64
	VersioningEnabled bool
	AccessLevel       string
	ArchiveTier       string
	RetentionPolicy   string
	LastAccessedAt    *time.Time
	DataFormat        string
	PartitionKey      string
	RecordCount       int64
	LastWriteAt       *time.Time
	CreatedAt         time.Time
}

// ListObjectBuckets returns all object bucket resources in a workspace filtered by storage class.
func ListObjectBuckets(ctx context.Context, pool *pgxpool.Pool, workspaceID int64, storageClass string) ([]ObjectBucket, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       ob.storage_class, ob.size_gb, ob.object_count,
		       ob.versioning_enabled, ob.access_level,
		       ob.archive_tier, ob.retention_policy, ob.last_accessed_at,
		       ob.data_format, ob.partition_key, ob.record_count, ob.last_write_at,
		       r.created_at
		FROM workspace_object_buckets ob
		JOIN resources r ON r.id = ob.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND ob.storage_class = $2
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID, storageClass)
	if err != nil {
		return nil, fmt.Errorf("db.ListObjectBuckets: %w", err)
	}
	defer rows.Close()
	var out []ObjectBucket
	for rows.Next() {
		var ob ObjectBucket
		if err := rows.Scan(
			&ob.ResourceID, &ob.Name, &ob.Status, &ob.Region,
			&ob.StorageClass, &ob.SizeGB, &ob.ObjectCount,
			&ob.VersioningEnabled, &ob.AccessLevel,
			&ob.ArchiveTier, &ob.RetentionPolicy, &ob.LastAccessedAt,
			&ob.DataFormat, &ob.PartitionKey, &ob.RecordCount, &ob.LastWriteAt,
			&ob.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, ob)
	}
	return out, rows.Err()
}

// CreateObjectBucket inserts a new object bucket resource.
func CreateObjectBucket(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, storageClass, region, accessLevel, archiveTier, retentionPolicy, dataFormat, partitionKey string,
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
		VALUES ($1,$2,'object_bucket',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateObjectBucket resource: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_object_buckets
		    (resource_id, storage_class, access_level, archive_tier, retention_policy, data_format, partition_key)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, storageClass, accessLevel, archiveTier, retentionPolicy, dataFormat, partitionKey)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateObjectBucket bucket row: %w", err)
	}

	return resourceID, nil
}
