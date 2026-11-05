package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── AI Batch Inference Jobs ────────────────────────────────────────────────────

type AIBatchJobDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Model      string
	InputPath  string
	OutputPath string
	Records    string
	Processed  string
	Progress   string
	Runtime    string
	Cost       string
	CreatedAt  time.Time
}

func ListAIBatchJobs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIBatchJobDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       b.model, b.input_path, b.output_path,
		       b.records, b.processed, b.progress,
		       b.runtime, b.cost,
		       r.created_at
		FROM workspace_ai_batch_jobs b
		JOIN resources r ON r.id = b.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIBatchJobs: %w", err)
	}
	defer rows.Close()
	var out []AIBatchJobDB
	for rows.Next() {
		var v AIBatchJobDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Model, &v.InputPath, &v.OutputPath,
			&v.Records, &v.Processed, &v.Progress,
			&v.Runtime, &v.Cost,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIBatchJob(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, model, inputPath, outputPath, region string,
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
		VALUES ($1,$2,'ai_batch_job',$3,'queued',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIBatchJob resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_batch_jobs
		    (resource_id, model, input_path, output_path)
		VALUES ($1,$2,$3,$4)`,
		resourceID, model, inputPath, outputPath)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIBatchJob batch row: %w", err)
	}
	return resourceID, nil
}

// ── AI Edge Deployments ────────────────────────────────────────────────────────

type AIEdgeDeploymentDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Model      string
	Runtime    string
	Target     string
	Devices    int
	Version    string
	Latency    string
	CreatedAt  time.Time
}

func ListAIEdgeDeployments(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIEdgeDeploymentDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       e.model, e.runtime, e.target,
		       e.devices, e.version, e.latency,
		       r.created_at
		FROM workspace_ai_edge_deployments e
		JOIN resources r ON r.id = e.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIEdgeDeployments: %w", err)
	}
	defer rows.Close()
	var out []AIEdgeDeploymentDB
	for rows.Next() {
		var v AIEdgeDeploymentDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Model, &v.Runtime, &v.Target,
			&v.Devices, &v.Version, &v.Latency,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIEdgeDeployment(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, model, runtime, target, region string,
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
		VALUES ($1,$2,'ai_edge_deployment',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIEdgeDeployment resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_edge_deployments
		    (resource_id, model, runtime, target)
		VALUES ($1,$2,$3,$4)`,
		resourceID, model, runtime, target)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIEdgeDeployment edge row: %w", err)
	}
	return resourceID, nil
}

// ── AI Vision APIs ─────────────────────────────────────────────────────────────

type AIVisionAPIDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Model      string
	Capability string
	Version    string
	Requests   string
	Latency    string
	Accuracy   string
	Endpoint   string
	APIKey     string
	CreatedAt  time.Time
}

func ListAIVisionAPIs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIVisionAPIDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       v.model, v.capability, v.version,
		       v.requests, v.latency, v.accuracy,
		       v.endpoint, v.api_key,
		       r.created_at
		FROM workspace_ai_vision_apis v
		JOIN resources r ON r.id = v.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIVisionAPIs: %w", err)
	}
	defer rows.Close()
	var out []AIVisionAPIDB
	for rows.Next() {
		var v AIVisionAPIDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Model, &v.Capability, &v.Version,
			&v.Requests, &v.Latency, &v.Accuracy,
			&v.Endpoint, &v.APIKey,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIVisionAPI(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, model, capability, region string,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	endpoint := name + ".vision.api." + region + ".example.com"

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'ai_vision_api',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIVisionAPI resource: %w", err)
	}
	apiKey := fmt.Sprintf("vk_prod_****%04x", resourceID)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_vision_apis
		    (resource_id, model, capability, endpoint, api_key)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, model, capability, endpoint, apiKey)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIVisionAPI vision row: %w", err)
	}
	return resourceID, nil
}

// ── AI Language APIs ───────────────────────────────────────────────────────────

type AILanguageAPIDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Model      string
	Capability string
	Version    string
	Requests   string
	Latency    string
	Languages  string
	Endpoint   string
	APIKey     string
	CreatedAt  time.Time
}

func ListAILanguageAPIs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AILanguageAPIDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       l.model, l.capability, l.version,
		       l.requests, l.latency, l.languages,
		       l.endpoint, l.api_key,
		       r.created_at
		FROM workspace_ai_language_apis l
		JOIN resources r ON r.id = l.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAILanguageAPIs: %w", err)
	}
	defer rows.Close()
	var out []AILanguageAPIDB
	for rows.Next() {
		var v AILanguageAPIDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Model, &v.Capability, &v.Version,
			&v.Requests, &v.Latency, &v.Languages,
			&v.Endpoint, &v.APIKey,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAILanguageAPI(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, model, capability, languages, region string,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	endpoint := name + ".nlp.api." + region + ".example.com"

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'ai_language_api',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAILanguageAPI resource: %w", err)
	}
	apiKey := fmt.Sprintf("vk_prod_****%04x", resourceID)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_language_apis
		    (resource_id, model, capability, languages, endpoint, api_key)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, model, capability, languages, endpoint, apiKey)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAILanguageAPI language row: %w", err)
	}
	return resourceID, nil
}

// ── AI Speech APIs ─────────────────────────────────────────────────────────────

type AISpeechAPIDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Model      string
	Capability string
	Version    string
	Requests   string
	Latency    string
	Languages  string
	SampleRate string
	Endpoint   string
	APIKey     string
	CreatedAt  time.Time
}

func ListAISpeechAPIs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AISpeechAPIDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.model, s.capability, s.version,
		       s.requests, s.latency, s.languages,
		       s.sample_rate, s.endpoint, s.api_key,
		       r.created_at
		FROM workspace_ai_speech_apis s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAISpeechAPIs: %w", err)
	}
	defer rows.Close()
	var out []AISpeechAPIDB
	for rows.Next() {
		var v AISpeechAPIDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Model, &v.Capability, &v.Version,
			&v.Requests, &v.Latency, &v.Languages,
			&v.SampleRate, &v.Endpoint, &v.APIKey,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAISpeechAPI(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, model, capability, languages, sampleRate, region string,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	endpoint := name + ".speech.api." + region + ".example.com"

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'ai_speech_api',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAISpeechAPI resource: %w", err)
	}
	apiKey := fmt.Sprintf("vk_prod_****%04x", resourceID)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_speech_apis
		    (resource_id, model, capability, languages, sample_rate, endpoint, api_key)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, model, capability, languages, sampleRate, endpoint, apiKey)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAISpeechAPI speech row: %w", err)
	}
	return resourceID, nil
}
