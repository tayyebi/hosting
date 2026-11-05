package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Multi-Cloud & Hybrid: Edge Computing ─────────────────────────────────────

type MulticloudEdgeNodeDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Location    string
	Provider    string
	CPUCores    int
	MemoryGB    int
	Workloads   int
	Uptime      string
	Latency     string
	LastContact string
	NodeStatus  string
	CreatedAt   time.Time
}

func ListMulticloudEdgeNodes(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudEdgeNodeDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       e.location, e.provider, e.cpu_cores, e.memory_gb,
		       e.workloads, e.uptime, e.latency, e.last_contact, e.node_status,
		       r.created_at
		FROM workspace_multicloud_edge_nodes e
		JOIN resources r ON r.id = e.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudEdgeNodes: %w", err)
	}
	defer rows.Close()
	var out []MulticloudEdgeNodeDB
	for rows.Next() {
		var v MulticloudEdgeNodeDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Location, &v.Provider, &v.CPUCores, &v.MemoryGB,
			&v.Workloads, &v.Uptime, &v.Latency, &v.LastContact, &v.NodeStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudEdgeNode(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, location, provider string,
	cpuCores, memoryGB int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_edge_node',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudEdgeNode resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_edge_nodes (resource_id, location, provider, cpu_cores, memory_gb)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, location, provider, cpuCores, memoryGB)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudEdgeNode row: %w", err)
	}
	return resourceID, nil
}

// ── Multi-Cloud Management: Cost Management ───────────────────────────────────

type MulticloudCostProfileDB struct {
	ResourceID      int64
	Name            string
	Status          string
	Region          string
	Provider        string
	Account         string
	MonthlyBudget   string
	CurrentSpend    string
	Forecast        string
	SavingsFound    string
	Alerts          int
	OptimizationPct string
	ProfileStatus   string
	CreatedAt       time.Time
}

func ListMulticloudCostProfiles(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudCostProfileDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       c.provider, c.account, c.monthly_budget, c.current_spend,
		       c.forecast, c.savings_found, c.alerts, c.optimization_pct, c.profile_status,
		       r.created_at
		FROM workspace_multicloud_cost_profiles c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudCostProfiles: %w", err)
	}
	defer rows.Close()
	var out []MulticloudCostProfileDB
	for rows.Next() {
		var v MulticloudCostProfileDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Provider, &v.Account, &v.MonthlyBudget, &v.CurrentSpend,
			&v.Forecast, &v.SavingsFound, &v.Alerts, &v.OptimizationPct, &v.ProfileStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudCostProfile(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, provider, account, monthlyBudget string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_cost_profile',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudCostProfile resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_cost_profiles (resource_id, provider, account, monthly_budget)
		VALUES ($1,$2,$3,$4)`,
		resourceID, provider, account, monthlyBudget)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudCostProfile row: %w", err)
	}
	return resourceID, nil
}

// ── Multi-Cloud Management: Policy Governance ─────────────────────────────────

type MulticloudGovernancePolicyDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Category     string
	Scope        string
	Enforcement  string
	Compliant    int
	NonCompliant int
	Exemptions   int
	Severity     string
	LastEval     string
	PolicyStatus string
	CreatedAt    time.Time
}

func ListMulticloudGovernancePolicies(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudGovernancePolicyDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       p.category, p.scope, p.enforcement,
		       p.compliant, p.non_compliant, p.exemptions,
		       p.severity, p.last_eval, p.policy_status,
		       r.created_at
		FROM workspace_multicloud_governance_policies p
		JOIN resources r ON r.id = p.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudGovernancePolicies: %w", err)
	}
	defer rows.Close()
	var out []MulticloudGovernancePolicyDB
	for rows.Next() {
		var v MulticloudGovernancePolicyDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Category, &v.Scope, &v.Enforcement,
			&v.Compliant, &v.NonCompliant, &v.Exemptions,
			&v.Severity, &v.LastEval, &v.PolicyStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudGovernancePolicy(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, category, scope, enforcement, severity string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_governance_policy',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudGovernancePolicy resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_governance_policies
		       (resource_id, category, scope, enforcement, severity)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, category, scope, enforcement, severity)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudGovernancePolicy row: %w", err)
	}
	return resourceID, nil
}

// ── Multi-Cloud Management: Workload Portability ──────────────────────────────

type MulticloudWorkloadProfileDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	SourceProvider string
	TargetProvider string
	WorkloadType   string
	Components     int
	Portability    string
	DriftStatus    string
	LastSync       string
	WorkloadStatus string
	CreatedAt      time.Time
}

func ListMulticloudWorkloadProfiles(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudWorkloadProfileDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       w.source_provider, w.target_provider, w.workload_type,
		       w.components, w.portability, w.drift_status, w.last_sync, w.workload_status,
		       r.created_at
		FROM workspace_multicloud_workload_profiles w
		JOIN resources r ON r.id = w.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudWorkloadProfiles: %w", err)
	}
	defer rows.Close()
	var out []MulticloudWorkloadProfileDB
	for rows.Next() {
		var v MulticloudWorkloadProfileDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.SourceProvider, &v.TargetProvider, &v.WorkloadType,
			&v.Components, &v.Portability, &v.DriftStatus, &v.LastSync, &v.WorkloadStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudWorkloadProfile(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, sourceProvider, targetProvider, workloadType string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_workload_profile',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudWorkloadProfile resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_workload_profiles
		       (resource_id, source_provider, target_provider, workload_type)
		VALUES ($1,$2,$3,$4)`,
		resourceID, sourceProvider, targetProvider, workloadType)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudWorkloadProfile row: %w", err)
	}
	return resourceID, nil
}

// ── Multi-Cloud Migration: Assessment & Discovery ─────────────────────────────

type MulticloudMigrationAssessmentDB struct {
	ResourceID        int64
	Name              string
	Status            string
	Region            string
	SourceEnv         string
	TargetEnv         string
	DiscoveredAssets  int
	ReadyToMigrate    int
	NeedsRemediation  int
	Unsupported       int
	EstimatedCost     string
	Complexity        string
	LastScan          string
	AssessmentStatus  string
	CreatedAt         time.Time
}

func ListMulticloudMigrationAssessments(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]MulticloudMigrationAssessmentDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       a.source_env, a.target_env,
		       a.discovered_assets, a.ready_to_migrate, a.needs_remediation, a.unsupported,
		       a.estimated_cost, a.complexity, a.last_scan, a.assessment_status,
		       r.created_at
		FROM workspace_multicloud_migration_assessments a
		JOIN resources r ON r.id = a.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListMulticloudMigrationAssessments: %w", err)
	}
	defer rows.Close()
	var out []MulticloudMigrationAssessmentDB
	for rows.Next() {
		var v MulticloudMigrationAssessmentDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.SourceEnv, &v.TargetEnv,
			&v.DiscoveredAssets, &v.ReadyToMigrate, &v.NeedsRemediation, &v.Unsupported,
			&v.EstimatedCost, &v.Complexity, &v.LastScan, &v.AssessmentStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateMulticloudMigrationAssessment(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, sourceEnv, targetEnv string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'multicloud_migration_assessment',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateMulticloudMigrationAssessment resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_multicloud_migration_assessments (resource_id, source_env, target_env)
		VALUES ($1,$2,$3)`,
		resourceID, sourceEnv, targetEnv)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateMulticloudMigrationAssessment row: %w", err)
	}
	return resourceID, nil
}
