package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── AI Recommendation APIs ─────────────────────────────────────────────────────

type AIRecommendationAPIDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Model      string
	Algorithm  string
	Version    string
	Requests   string
	Latency    string
	Catalog    string
	Endpoint   string
	APIKey     string
	CreatedAt  time.Time
}

func ListAIRecommendationAPIs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIRecommendationAPIDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       a.model, a.algorithm, a.version,
		       a.requests, a.latency, a.catalog,
		       a.endpoint, a.api_key,
		       r.created_at
		FROM workspace_ai_recommendation_apis a
		JOIN resources r ON r.id = a.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIRecommendationAPIs: %w", err)
	}
	defer rows.Close()
	var out []AIRecommendationAPIDB
	for rows.Next() {
		var v AIRecommendationAPIDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Model, &v.Algorithm, &v.Version,
			&v.Requests, &v.Latency, &v.Catalog,
			&v.Endpoint, &v.APIKey,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIRecommendationAPI(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, model, algorithm, region string,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	endpoint := "recommend-" + name + ".api." + region + ".example.com"

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'ai_recommendation_api',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIRecommendationAPI resource: %w", err)
	}
	apiKey := fmt.Sprintf("vk_prod_****%04x", resourceID)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_recommendation_apis
		    (resource_id, model, algorithm, endpoint, api_key)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, model, algorithm, endpoint, apiKey)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIRecommendationAPI api row: %w", err)
	}
	return resourceID, nil
}

// ── AI ETL / ELT Pipelines ─────────────────────────────────────────────────────

type AIETLPipelineDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	PipelineType string
	Source       string
	Destination  string
	Schedule     string
	LastRun      string
	Duration     string
	Records      string
	CreatedAt    time.Time
}

func ListAIETLPipelines(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIETLPipelineDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       p.pipeline_type, p.source, p.destination,
		       p.schedule, p.last_run, p.duration, p.records,
		       r.created_at
		FROM workspace_ai_etl_pipelines p
		JOIN resources r ON r.id = p.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIETLPipelines: %w", err)
	}
	defer rows.Close()
	var out []AIETLPipelineDB
	for rows.Next() {
		var v AIETLPipelineDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.PipelineType, &v.Source, &v.Destination,
			&v.Schedule, &v.LastRun, &v.Duration, &v.Records,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIETLPipeline(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, pipelineType, source, destination, schedule, region string,
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
		VALUES ($1,$2,'ai_etl_pipeline',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIETLPipeline resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_etl_pipelines
		    (resource_id, pipeline_type, source, destination, schedule)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, pipelineType, source, destination, schedule)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIETLPipeline pipeline row: %w", err)
	}
	return resourceID, nil
}

// ── AI Data Streams ────────────────────────────────────────────────────────────

type AIDataStreamDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Engine      string
	Source      string
	Destination string
	Throughput  string
	Partitions  int
	Retention   string
	Lag         string
	CreatedAt   time.Time
}

func ListAIDataStreams(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIDataStreamDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.engine, s.source, s.destination,
		       s.throughput, s.partitions, s.retention, s.lag,
		       r.created_at
		FROM workspace_ai_data_streams s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIDataStreams: %w", err)
	}
	defer rows.Close()
	var out []AIDataStreamDB
	for rows.Next() {
		var v AIDataStreamDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Engine, &v.Source, &v.Destination,
			&v.Throughput, &v.Partitions, &v.Retention, &v.Lag,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIDataStream(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, engine, source, destination, region string,
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
		VALUES ($1,$2,'ai_data_stream',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIDataStream resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_data_streams
		    (resource_id, engine, source, destination)
		VALUES ($1,$2,$3,$4)`,
		resourceID, engine, source, destination)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIDataStream stream row: %w", err)
	}
	return resourceID, nil
}

// ── AI Data Catalog Assets ─────────────────────────────────────────────────────

type AIDataCatalogAssetDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Source       string
	Owner        string
	AssetType    string
	PII          string
	Tags         string
	AssetUpdated string
	Columns      int
	CreatedAt    time.Time
}

func ListAIDataCatalogAssets(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIDataCatalogAssetDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       c.source, c.owner, c.asset_type,
		       c.pii, c.tags, c.asset_updated, c.columns,
		       r.created_at
		FROM workspace_ai_data_catalog_assets c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIDataCatalogAssets: %w", err)
	}
	defer rows.Close()
	var out []AIDataCatalogAssetDB
	for rows.Next() {
		var v AIDataCatalogAssetDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Source, &v.Owner, &v.AssetType,
			&v.PII, &v.Tags, &v.AssetUpdated, &v.Columns,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIDataCatalogAsset(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, source, owner, assetType, pii, tags, region string,
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
		VALUES ($1,$2,'ai_catalog_asset',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIDataCatalogAsset resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_data_catalog_assets
		    (resource_id, source, owner, asset_type, pii, tags)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, source, owner, assetType, pii, tags)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIDataCatalogAsset asset row: %w", err)
	}
	return resourceID, nil
}

// ── AI ML Experiments ──────────────────────────────────────────────────────────

type AIMLExperimentDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	BestRun    string
	Metric     string
	Runs       int
	Framework  string
	Owner      string
	CreatedAt  time.Time
}

func ListAIMLExperiments(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIMLExperimentDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       e.best_run, e.metric, e.runs,
		       e.framework, e.owner,
		       r.created_at
		FROM workspace_ai_ml_experiments e
		JOIN resources r ON r.id = e.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIMLExperiments: %w", err)
	}
	defer rows.Close()
	var out []AIMLExperimentDB
	for rows.Next() {
		var v AIMLExperimentDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.BestRun, &v.Metric, &v.Runs,
			&v.Framework, &v.Owner,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIMLExperiment(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, framework, owner, region string,
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
		VALUES ($1,$2,'ai_ml_experiment',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIMLExperiment resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_ml_experiments
		    (resource_id, framework, owner)
		VALUES ($1,$2,$3)`,
		resourceID, framework, owner)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIMLExperiment experiment row: %w", err)
	}
	return resourceID, nil
}
