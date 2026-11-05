package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Security: Intrusion Detection ─────────────────────────────────────────────

type SecurityIDSRuleDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Severity       string
	Direction      string
	Matches24h     string
	FalsePositives int
	Action         string
	LastTriggered  string
	CreatedAt      time.Time
}

func ListSecurityIDSRules(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityIDSRuleDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       i.severity, i.direction, i.matches_24h,
		       i.false_positives, i.action, i.last_triggered,
		       r.created_at
		FROM workspace_security_network_ids i
		JOIN resources r ON r.id = i.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityIDSRules: %w", err)
	}
	defer rows.Close()
	var out []SecurityIDSRuleDB
	for rows.Next() {
		var v SecurityIDSRuleDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Severity, &v.Direction, &v.Matches24h,
			&v.FalsePositives, &v.Action, &v.LastTriggered,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityIDSRule(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, severity, direction, action, region string,
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
		VALUES ($1,$2,'security_ids_rule',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityIDSRule resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_network_ids
		    (resource_id, severity, direction, action)
		VALUES ($1,$2,$3,$4)`,
		resourceID, severity, direction, action)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityIDSRule row: %w", err)
	}
	return resourceID, nil
}

// ── Security: Compliance Policy ────────────────────────────────────────────────

type SecurityCompliancePolicyDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Severity       string
	Framework      string
	CompliantCount int
	TotalCount     int
	Drift7d        string
	AutoRemediate  string
	LastEvaluated  string
	CreatedAt      time.Time
}

func ListSecurityCompliancePolicies(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityCompliancePolicyDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       p.severity, p.framework, p.compliant_count, p.total_count,
		       p.drift_7d, p.auto_remediate, p.last_evaluated,
		       r.created_at
		FROM workspace_security_compliance_policy p
		JOIN resources r ON r.id = p.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityCompliancePolicies: %w", err)
	}
	defer rows.Close()
	var out []SecurityCompliancePolicyDB
	for rows.Next() {
		var v SecurityCompliancePolicyDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Severity, &v.Framework, &v.CompliantCount, &v.TotalCount,
			&v.Drift7d, &v.AutoRemediate, &v.LastEvaluated,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityCompliancePolicy(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, severity, framework, autoRemediate string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'security_compliance_policy',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityCompliancePolicy resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_compliance_policy
		    (resource_id, severity, framework, auto_remediate)
		VALUES ($1,$2,$3,$4)`,
		resourceID, severity, framework, autoRemediate)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityCompliancePolicy row: %w", err)
	}
	return resourceID, nil
}

// ── Security: Audit Log Entries ────────────────────────────────────────────────

type SecurityAuditEventDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	Actor       string
	SourceIP    string
	Resource    string
	AuditRegion string
	Severity    string
	AuditStatus string
	CreatedAt   time.Time
}

func ListSecurityAuditEvents(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityAuditEventDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       a.actor, a.source_ip, a.resource,
		       a.region, a.severity, a.audit_status,
		       r.created_at
		FROM workspace_security_compliance_audit a
		JOIN resources r ON r.id = a.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityAuditEvents: %w", err)
	}
	defer rows.Close()
	var out []SecurityAuditEventDB
	for rows.Next() {
		var v SecurityAuditEventDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Actor, &v.SourceIP, &v.Resource,
			&v.AuditRegion, &v.Severity, &v.AuditStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityAuditEvent(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, actor, sourceIP, resource, region, severity string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'security_audit_event',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityAuditEvent resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_compliance_audit
		    (resource_id, actor, source_ip, resource, region, severity)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resourceID, actor, sourceIP, resource, region, severity)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityAuditEvent row: %w", err)
	}
	return resourceID, nil
}

// ── Security: Regulatory Frameworks ──────────────────────────────────────────

type SecurityRegulatoryFrameworkDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Scope          string
	ControlsPassed int
	ControlsTotal  int
	EvidenceItems  string
	LastAssessment string
	Assessor       string
	Renewal        string
	CreatedAt      time.Time
}

func ListSecurityRegulatoryFrameworks(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityRegulatoryFrameworkDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       f.scope, f.controls_passed, f.controls_total,
		       f.evidence_items, f.last_assessment, f.assessor, f.renewal,
		       r.created_at
		FROM workspace_security_compliance_regulatory f
		JOIN resources r ON r.id = f.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityRegulatoryFrameworks: %w", err)
	}
	defer rows.Close()
	var out []SecurityRegulatoryFrameworkDB
	for rows.Next() {
		var v SecurityRegulatoryFrameworkDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Scope, &v.ControlsPassed, &v.ControlsTotal,
			&v.EvidenceItems, &v.LastAssessment, &v.Assessor, &v.Renewal,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityRegulatoryFramework(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, scope, assessor, renewal string,
	controlsTotal int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'security_regulatory_framework',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityRegulatoryFramework resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_compliance_regulatory
		    (resource_id, scope, controls_total, assessor, renewal)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, scope, controlsTotal, assessor, renewal)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityRegulatoryFramework row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: Source Control Repositories ──────────────────────────────────────

type DevOpsCICDSourceRepoDB struct {
	ResourceID       int64
	Name             string
	Status           string
	Region           string
	DefaultBranch    string
	Language         string
	LastPush         string
	PRsOpen          int
	Coverage         string
	CIStatus         string
	BranchProtection string
	CreatedAt        time.Time
}

func ListDevOpsCICDSourceRepos(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsCICDSourceRepoDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.default_branch, s.language, s.last_push,
		       s.prs_open, s.coverage, s.ci_status, s.branch_protection,
		       r.created_at
		FROM workspace_devops_cicd_source s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsCICDSourceRepos: %w", err)
	}
	defer rows.Close()
	var out []DevOpsCICDSourceRepoDB
	for rows.Next() {
		var v DevOpsCICDSourceRepoDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.DefaultBranch, &v.Language, &v.LastPush,
			&v.PRsOpen, &v.Coverage, &v.CIStatus, &v.BranchProtection,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsCICDSourceRepo(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, defaultBranch, language, coverage string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_cicd_source_repo',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsCICDSourceRepo resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_cicd_source
		    (resource_id, default_branch, language, coverage)
		VALUES ($1,$2,$3,$4)`,
		resourceID, defaultBranch, language, coverage)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsCICDSourceRepo row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: CI/CD Build Pipelines ─────────────────────────────────────────────

type DevOpsCICDBuildPipelineDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Repo       string
	Trigger    string
	Runtime    string
	LastRun    string
	Duration   string
	CacheHit   string
	BuildStatus string
	CreatedAt  time.Time
}

func ListDevOpsCICDBuildPipelines(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsCICDBuildPipelineDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       b.repo, b.trigger, b.runtime, b.last_run,
		       b.duration, b.cache_hit, b.status,
		       r.created_at
		FROM workspace_devops_cicd_build b
		JOIN resources r ON r.id = b.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsCICDBuildPipelines: %w", err)
	}
	defer rows.Close()
	var out []DevOpsCICDBuildPipelineDB
	for rows.Next() {
		var v DevOpsCICDBuildPipelineDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Repo, &v.Trigger, &v.Runtime, &v.LastRun,
			&v.Duration, &v.CacheHit, &v.BuildStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsCICDBuildPipeline(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, repo, trigger, runtime string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_cicd_build_pipeline',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsCICDBuildPipeline resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_cicd_build (resource_id, repo, trigger, runtime)
		VALUES ($1,$2,$3,$4)`,
		resourceID, repo, trigger, runtime)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsCICDBuildPipeline row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: CI/CD Deploy Pipelines ────────────────────────────────────────────

type DevOpsCICDDeployPipelineDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	App          string
	Environment  string
	Strategy     string
	Traffic      string
	ErrorRate    string
	LatencyP99   string
	Duration     string
	DeployStatus string
	CreatedAt    time.Time
}

func ListDevOpsCICDDeployPipelines(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsCICDDeployPipelineDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       d.app, d.environment, d.strategy, d.traffic,
		       d.error_rate, d.latency_p99, d.duration, d.deploy_status,
		       r.created_at
		FROM workspace_devops_cicd_deploy d
		JOIN resources r ON r.id = d.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsCICDDeployPipelines: %w", err)
	}
	defer rows.Close()
	var out []DevOpsCICDDeployPipelineDB
	for rows.Next() {
		var v DevOpsCICDDeployPipelineDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.App, &v.Environment, &v.Strategy, &v.Traffic,
			&v.ErrorRate, &v.LatencyP99, &v.Duration, &v.DeployStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsCICDDeployPipeline(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, app, environment, strategy string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_cicd_deploy_pipeline',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsCICDDeployPipeline resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_cicd_deploy (resource_id, app, environment, strategy)
		VALUES ($1,$2,$3,$4)`,
		resourceID, app, environment, strategy)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsCICDDeployPipeline row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: IaC Templates / Stacks ────────────────────────────────────────────

type DevOpsIaCTemplateDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	Tool          string
	ResourceCount int
	StateBackend  string
	LastApplied   string
	PlanOutput    string
	DriftStatus   string
	CreatedAt     time.Time
}

func ListDevOpsIaCTemplates(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsIaCTemplateDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       t.tool, t.resource_count, t.state_backend,
		       t.last_applied, t.plan_output, t.drift_status,
		       r.created_at
		FROM workspace_devops_iac_templates t
		JOIN resources r ON r.id = t.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsIaCTemplates: %w", err)
	}
	defer rows.Close()
	var out []DevOpsIaCTemplateDB
	for rows.Next() {
		var v DevOpsIaCTemplateDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Tool, &v.ResourceCount, &v.StateBackend,
			&v.LastApplied, &v.PlanOutput, &v.DriftStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsIaCTemplate(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, tool, stateBackend string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_iac_template',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsIaCTemplate resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_iac_templates (resource_id, tool, state_backend)
		VALUES ($1,$2,$3)`,
		resourceID, tool, stateBackend)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsIaCTemplate row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: IaC Config Management ─────────────────────────────────────────────

type DevOpsIaCConfigDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Tool       string
	HostCount  int
	Inventory  string
	LastApply  string
	Changed    string
	Compliance string
	CfgStatus  string
	CreatedAt  time.Time
}

func ListDevOpsIaCConfigs(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsIaCConfigDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       c.tool, c.host_count, c.inventory,
		       c.last_apply, c.changed, c.compliance, c.cfg_status,
		       r.created_at
		FROM workspace_devops_iac_config c
		JOIN resources r ON r.id = c.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsIaCConfigs: %w", err)
	}
	defer rows.Close()
	var out []DevOpsIaCConfigDB
	for rows.Next() {
		var v DevOpsIaCConfigDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Tool, &v.HostCount, &v.Inventory,
			&v.LastApply, &v.Changed, &v.Compliance, &v.CfgStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsIaCConfig(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, tool, scope string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_iac_config',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsIaCConfig resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_iac_config (resource_id, tool, inventory)
		VALUES ($1,$2,$3)`,
		resourceID, tool, scope)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsIaCConfig row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: API Gateway ───────────────────────────────────────────────────────

type DevOpsAPIGatewayDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	APIVersion string
	Protocol   string
	Calls24h   string
	LatencyP95 string
	RateLimit  string
	AuthType   string
	ErrorRate  string
	GwStatus   string
	CreatedAt  time.Time
}

func ListDevOpsAPIGateways(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsAPIGatewayDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       g.api_version, g.protocol, g.calls_24h,
		       g.latency_p95, g.rate_limit, g.auth_type,
		       g.error_rate, g.gw_status,
		       r.created_at
		FROM workspace_devops_api_gateway g
		JOIN resources r ON r.id = g.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsAPIGateways: %w", err)
	}
	defer rows.Close()
	var out []DevOpsAPIGatewayDB
	for rows.Next() {
		var v DevOpsAPIGatewayDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.APIVersion, &v.Protocol, &v.Calls24h,
			&v.LatencyP95, &v.RateLimit, &v.AuthType,
			&v.ErrorRate, &v.GwStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsAPIGateway(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, protocol, authType, rateLimit string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_api_gateway',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsAPIGateway resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_api_gateway (resource_id, protocol, auth_type, rate_limit)
		VALUES ($1,$2,$3,$4)`,
		resourceID, protocol, authType, rateLimit)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsAPIGateway row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: API Developer Portal ──────────────────────────────────────────────

type DevOpsAPIPortalDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Owner        string
	Plan         string
	APIKeys      string
	Calls7d      string
	LastUsed     string
	QuotaUsage   string
	PortalStatus string
	CreatedAt    time.Time
}

func ListDevOpsAPIPortals(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsAPIPortalDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       p.owner, p.plan, p.api_keys,
		       p.calls_7d, p.last_used, p.quota_usage, p.portal_status,
		       r.created_at
		FROM workspace_devops_api_portal p
		JOIN resources r ON r.id = p.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsAPIPortals: %w", err)
	}
	defer rows.Close()
	var out []DevOpsAPIPortalDB
	for rows.Next() {
		var v DevOpsAPIPortalDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Owner, &v.Plan, &v.APIKeys,
			&v.Calls7d, &v.LastUsed, &v.QuotaUsage, &v.PortalStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsAPIPortal(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, owner, plan string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_api_portal',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsAPIPortal resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_api_portal (resource_id, owner, plan)
		VALUES ($1,$2,$3)`,
		resourceID, owner, plan)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsAPIPortal row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: API Analytics ─────────────────────────────────────────────────────

type DevOpsAPIAnalyticsDB struct {
	ResourceID  int64
	Name        string
	Status      string
	Region      string
	APIName     string
	Method      string
	Calls24h    string
	LatencyP50  string
	LatencyP95  string
	LatencyP99  string
	ErrorRate   string
	EpStatus    string
	CreatedAt   time.Time
}

func ListDevOpsAPIAnalytics(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsAPIAnalyticsDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       a.api_name, a.method, a.calls_24h,
		       a.latency_p50, a.latency_p95, a.latency_p99,
		       a.error_rate, a.ep_status,
		       r.created_at
		FROM workspace_devops_api_analytics a
		JOIN resources r ON r.id = a.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsAPIAnalytics: %w", err)
	}
	defer rows.Close()
	var out []DevOpsAPIAnalyticsDB
	for rows.Next() {
		var v DevOpsAPIAnalyticsDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.APIName, &v.Method, &v.Calls24h,
			&v.LatencyP50, &v.LatencyP95, &v.LatencyP99,
			&v.ErrorRate, &v.EpStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsAPIAnalytics(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, apiName, method string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_api_analytics',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsAPIAnalytics resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_api_analytics (resource_id, api_name, method)
		VALUES ($1,$2,$3)`,
		resourceID, apiName, method)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsAPIAnalytics row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: Observability — Logging ──────────────────────────────────────────

type DevOpsObsLoggingDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Source     string
	Format     string
	Volume24h  string
	Lines24h   string
	Retention  string
	Lag        string
	LogStatus  string
	CreatedAt  time.Time
}

func ListDevOpsObsLogging(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsObsLoggingDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       l.source, l.format, l.volume_24h,
		       l.lines_24h, l.retention, l.lag, l.log_status,
		       r.created_at
		FROM workspace_devops_obs_logging l
		JOIN resources r ON r.id = l.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsObsLogging: %w", err)
	}
	defer rows.Close()
	var out []DevOpsObsLoggingDB
	for rows.Next() {
		var v DevOpsObsLoggingDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Source, &v.Format, &v.Volume24h,
			&v.Lines24h, &v.Retention, &v.Lag, &v.LogStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsObsLogging(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, source, format, retention string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_obs_logging',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsObsLogging resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_obs_logging (resource_id, source, format, retention)
		VALUES ($1,$2,$3,$4)`,
		resourceID, source, format, retention)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsObsLogging row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: Observability — Metrics & Dashboards ──────────────────────────────

type DevOpsObsMetricsDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Owner        string
	Panels       int
	DataSource   string
	RefreshRate  string
	LastViewed   string
	AlertsLinked string
	DashStatus   string
	CreatedAt    time.Time
}

func ListDevOpsObsMetrics(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsObsMetricsDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       m.owner, m.panels, m.data_source,
		       m.refresh_rate, m.last_viewed, m.alerts_linked, m.dash_status,
		       r.created_at
		FROM workspace_devops_obs_metrics m
		JOIN resources r ON r.id = m.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsObsMetrics: %w", err)
	}
	defer rows.Close()
	var out []DevOpsObsMetricsDB
	for rows.Next() {
		var v DevOpsObsMetricsDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Owner, &v.Panels, &v.DataSource,
			&v.RefreshRate, &v.LastViewed, &v.AlertsLinked, &v.DashStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsObsMetrics(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, owner, dataSource string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_obs_metrics',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsObsMetrics resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_obs_metrics (resource_id, owner, data_source)
		VALUES ($1,$2,$3)`,
		resourceID, owner, dataSource)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsObsMetrics row: %w", err)
	}
	return resourceID, nil
}

// ── DevOps: Observability — Distributed Tracing ───────────────────────────────

type DevOpsObsTracingDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Spans24h   string
	LatencyP50 string
	LatencyP95 string
	LatencyP99 string
	ErrorRate  string
	Sampling   string
	SvcStatus  string
	CreatedAt  time.Time
}

func ListDevOpsObsTracing(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]DevOpsObsTracingDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       t.spans_24h, t.latency_p50, t.latency_p95,
		       t.latency_p99, t.error_rate, t.sampling, t.svc_status,
		       r.created_at
		FROM workspace_devops_obs_tracing t
		JOIN resources r ON r.id = t.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListDevOpsObsTracing: %w", err)
	}
	defer rows.Close()
	var out []DevOpsObsTracingDB
	for rows.Next() {
		var v DevOpsObsTracingDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Spans24h, &v.LatencyP50, &v.LatencyP95,
			&v.LatencyP99, &v.ErrorRate, &v.Sampling, &v.SvcStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateDevOpsObsTracing(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, sampling string,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'devops_obs_tracing',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateDevOpsObsTracing resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_devops_obs_tracing (resource_id, sampling)
		VALUES ($1,$2)`,
		resourceID, sampling)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateDevOpsObsTracing row: %w", err)
	}
	return resourceID, nil
}
