package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── SD-WAN Networks ────────────────────────────────────────────────────────────

type SDWANNetworkDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Topology   string
	Sites      int
	Hubs       int
	Bandwidth  string
	Protocol   string
	Encryption string
	QoS        bool
	Uptime     string
	CreatedAt  time.Time
}

func ListSDWANNetworks(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SDWANNetworkDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.topology, s.sites, s.hubs,
		       s.bandwidth, s.protocol, s.encryption,
		       s.qos, s.uptime,
		       r.created_at
		FROM workspace_sdwan_networks s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSDWANNetworks: %w", err)
	}
	defer rows.Close()
	var out []SDWANNetworkDB
	for rows.Next() {
		var v SDWANNetworkDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Topology, &v.Sites, &v.Hubs,
			&v.Bandwidth, &v.Protocol, &v.Encryption,
			&v.QoS, &v.Uptime,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSDWANNetwork(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, topology, protocol, encryption, region string,
	qos bool,
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
		VALUES ($1,$2,'sdwan_network',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSDWANNetwork resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_sdwan_networks
		    (resource_id, topology, protocol, encryption, qos)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, topology, protocol, encryption, qos)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSDWANNetwork sdwan_networks row: %w", err)
	}
	return resourceID, nil
}

// ── AI Distributed Training ────────────────────────────────────────────────────

type AIDistributedTrainingDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Framework   string
	Nodes       int
	GPUsPerNode int
	GPUType     string
	Dataset     string
	Epochs      int
	Progress    string
	Runtime     string
	Cost        string
	CreatedAt   time.Time
}

func ListAIDistributedTraining(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIDistributedTrainingDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       t.framework, t.nodes, t.gpus_per_node, t.gpu_type,
		       t.dataset, t.epochs, t.progress, t.runtime, t.cost,
		       r.created_at
		FROM workspace_ai_distributed_training t
		JOIN resources r ON r.id = t.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIDistributedTraining: %w", err)
	}
	defer rows.Close()
	var out []AIDistributedTrainingDB
	for rows.Next() {
		var v AIDistributedTrainingDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Framework, &v.Nodes, &v.GPUsPerNode, &v.GPUType,
			&v.Dataset, &v.Epochs, &v.Progress, &v.Runtime, &v.Cost,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIDistributedTraining(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, framework, gpuType, dataset, region string,
	nodes, gpusPerNode, epochs int,
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
		VALUES ($1,$2,'ai_distributed_training',$3,'queued',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIDistributedTraining resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_distributed_training
		    (resource_id, framework, nodes, gpus_per_node, gpu_type, dataset, epochs)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, framework, nodes, gpusPerNode, gpuType, dataset, epochs)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIDistributedTraining training row: %w", err)
	}
	return resourceID, nil
}

// ── AI AutoML Runs ─────────────────────────────────────────────────────────────

type AIAutoMLRunDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	TaskType   string
	Dataset    string
	TargetCol  string
	Trials     int
	BestModel  string
	BestScore  string
	Metric     string
	Runtime    string
	Cost       string
	CreatedAt  time.Time
}

func ListAIAutoMLRuns(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIAutoMLRunDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       a.task_type, a.dataset, a.target_col,
		       a.trials, a.best_model, a.best_score,
		       a.metric, a.runtime, a.cost,
		       r.created_at
		FROM workspace_ai_automl_runs a
		JOIN resources r ON r.id = a.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIAutoMLRuns: %w", err)
	}
	defer rows.Close()
	var out []AIAutoMLRunDB
	for rows.Next() {
		var v AIAutoMLRunDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.TaskType, &v.Dataset, &v.TargetCol,
			&v.Trials, &v.BestModel, &v.BestScore,
			&v.Metric, &v.Runtime, &v.Cost,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIAutoMLRun(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, taskType, dataset, targetCol, metric, region string,
	trials int,
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
		VALUES ($1,$2,'ai_automl_run',$3,'queued',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIAutoMLRun resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_automl_runs
		    (resource_id, task_type, dataset, target_col, trials, metric)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, taskType, dataset, targetCol, trials, metric)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIAutoMLRun automl row: %w", err)
	}
	return resourceID, nil
}

// ── AI Fine-tuning Jobs ────────────────────────────────────────────────────────

type AIFineTuningJobDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	BaseModel    string
	Method       string
	Dataset      string
	Epochs       int
	LearningRate string
	Progress     string
	Runtime      string
	Cost         string
	OutputModel  string
	CreatedAt    time.Time
}

func ListAIFineTuningJobs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIFineTuningJobDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       f.base_model, f.method, f.dataset,
		       f.epochs, f.learning_rate, f.progress,
		       f.runtime, f.cost, f.output_model,
		       r.created_at
		FROM workspace_ai_finetuning_jobs f
		JOIN resources r ON r.id = f.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIFineTuningJobs: %w", err)
	}
	defer rows.Close()
	var out []AIFineTuningJobDB
	for rows.Next() {
		var v AIFineTuningJobDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.BaseModel, &v.Method, &v.Dataset,
			&v.Epochs, &v.LearningRate, &v.Progress,
			&v.Runtime, &v.Cost, &v.OutputModel,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIFineTuningJob(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, baseModel, method, dataset, learningRate, region string,
	epochs int,
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
		VALUES ($1,$2,'ai_finetuning_job',$3,'queued',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIFineTuningJob resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_finetuning_jobs
		    (resource_id, base_model, method, dataset, epochs, learning_rate)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, baseModel, method, dataset, epochs, learningRate)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIFineTuningJob finetuning row: %w", err)
	}
	return resourceID, nil
}

// ── AI Inference Endpoints ─────────────────────────────────────────────────────

type AIInferenceEndpointDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Model        string
	Framework    string
	InstanceType string
	Replicas     int
	Latency      string
	RPS          string
	Endpoint     string
	AutoScale    bool
	CreatedAt    time.Time
}

func ListAIInferenceEndpoints(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIInferenceEndpointDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       e.model, e.framework, e.instance_type,
		       e.replicas, e.latency, e.rps,
		       e.endpoint, e.auto_scale,
		       r.created_at
		FROM workspace_ai_inference_endpoints e
		JOIN resources r ON r.id = e.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIInferenceEndpoints: %w", err)
	}
	defer rows.Close()
	var out []AIInferenceEndpointDB
	for rows.Next() {
		var v AIInferenceEndpointDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Model, &v.Framework, &v.InstanceType,
			&v.Replicas, &v.Latency, &v.RPS,
			&v.Endpoint, &v.AutoScale,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIInferenceEndpoint(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, model, framework, instanceType, region string,
	replicas int,
	autoScale bool,
) (int64, error) {
	var regionID *int64
	if region != "" {
		var rID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM regions WHERE slug=$1`, region).Scan(&rID); err == nil {
			regionID = &rID
		}
	}

	endpointURL := name + ".inf." + region + ".example.com"

	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, region_id, created_by)
		VALUES ($1,$2,'ai_inference_endpoint',$3,'provisioning',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIInferenceEndpoint resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_inference_endpoints
		    (resource_id, model, framework, instance_type, replicas, endpoint, auto_scale)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		resourceID, model, framework, instanceType, replicas, endpointURL, autoScale)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIInferenceEndpoint endpoints row: %w", err)
	}
	return resourceID, nil
}
