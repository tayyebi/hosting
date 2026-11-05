package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── AI Model Registry ──────────────────────────────────────────────────────────

type AIModelRegistryDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Version    string
	Format     string
	SizeBytes  string
	Stage      string
	RequestsS  string
	P99Latency string
	Owner      string
	LastUsed   string
	CreatedAt  time.Time
}

func ListAIModelRegistry(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIModelRegistryDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       m.version, m.format, m.size_bytes,
		       m.stage, m.requests_s, m.p99_latency,
		       m.owner, m.last_used,
		       r.created_at
		FROM workspace_ai_model_registry m
		JOIN resources r ON r.id = m.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIModelRegistry: %w", err)
	}
	defer rows.Close()
	var out []AIModelRegistryDB
	for rows.Next() {
		var v AIModelRegistryDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Version, &v.Format, &v.SizeBytes,
			&v.Stage, &v.RequestsS, &v.P99Latency,
			&v.Owner, &v.LastUsed,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIModelRegistry(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, version, format, stage, owner, region string,
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
		VALUES ($1,$2,'ai_model_registry',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIModelRegistry resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_model_registry
		    (resource_id, version, format, stage, owner)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, version, format, stage, owner)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIModelRegistry registry row: %w", err)
	}
	return resourceID, nil
}

// ── AI ML Monitors ─────────────────────────────────────────────────────────────

type AIMLMonitorDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	DriftPSI      string
	Confidence    string
	RequestsS     string
	P99Latency    string
	ErrorRate     string
	LastRetrain   string
	MonitorStatus string
	CreatedAt     time.Time
}

func ListAIMLMonitors(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]AIMLMonitorDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       m.drift_psi, m.confidence,
		       m.requests_s, m.p99_latency,
		       m.error_rate, m.last_retrain, m.monitor_status,
		       r.created_at
		FROM workspace_ai_ml_monitors m
		JOIN resources r ON r.id = m.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListAIMLMonitors: %w", err)
	}
	defer rows.Close()
	var out []AIMLMonitorDB
	for rows.Next() {
		var v AIMLMonitorDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.DriftPSI, &v.Confidence,
			&v.RequestsS, &v.P99Latency,
			&v.ErrorRate, &v.LastRetrain, &v.MonitorStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateAIMLMonitor(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, region string,
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
		VALUES ($1,$2,'ai_ml_monitor',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateAIMLMonitor resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_ai_ml_monitors (resource_id)
		VALUES ($1)`,
		resourceID)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateAIMLMonitor monitor row: %w", err)
	}
	return resourceID, nil
}

// ── Security SSO Providers ─────────────────────────────────────────────────────

type SecuritySSOProviderDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	Protocol     string
	Domain       string
	UsersCount   int
	Logins24h    int
	SCIM         string
	CertExpires  string
	LastRotation string
	CreatedAt    time.Time
}

func ListSecuritySSOProviders(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecuritySSOProviderDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       s.protocol, s.domain,
		       s.users_count, s.logins_24h,
		       s.scim, s.cert_expires, s.last_rotation,
		       r.created_at
		FROM workspace_security_sso_providers s
		JOIN resources r ON r.id = s.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecuritySSOProviders: %w", err)
	}
	defer rows.Close()
	var out []SecuritySSOProviderDB
	for rows.Next() {
		var v SecuritySSOProviderDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Protocol, &v.Domain,
			&v.UsersCount, &v.Logins24h,
			&v.SCIM, &v.CertExpires, &v.LastRotation,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecuritySSOProvider(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, protocol, domain, region string,
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
		VALUES ($1,$2,'security_sso_provider',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecuritySSOProvider resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_sso_providers
		    (resource_id, protocol, domain)
		VALUES ($1,$2,$3)`,
		resourceID, protocol, domain)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecuritySSOProvider sso row: %w", err)
	}
	return resourceID, nil
}

// ── Security MFA Users ─────────────────────────────────────────────────────────

type SecurityMFAUserDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	PrimaryMethod string
	BackupMethod  string
	Roles         string
	RecoveryCodes string
	Devices       int
	LastMFA       string
	MFAStatus     string
	CreatedAt     time.Time
}

func ListSecurityMFAUsers(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityMFAUserDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       m.primary_method, m.backup_method,
		       m.roles, m.recovery_codes,
		       m.devices, m.last_mfa, m.mfa_status,
		       r.created_at
		FROM workspace_security_mfa_users m
		JOIN resources r ON r.id = m.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityMFAUsers: %w", err)
	}
	defer rows.Close()
	var out []SecurityMFAUserDB
	for rows.Next() {
		var v SecurityMFAUserDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.PrimaryMethod, &v.BackupMethod,
			&v.Roles, &v.RecoveryCodes,
			&v.Devices, &v.LastMFA, &v.MFAStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityMFAUser(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, primaryMethod, roles, region string,
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
		VALUES ($1,$2,'security_mfa_user',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityMFAUser resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_mfa_users
		    (resource_id, primary_method, roles)
		VALUES ($1,$2,$3)`,
		resourceID, primaryMethod, roles)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityMFAUser mfa row: %w", err)
	}
	return resourceID, nil
}

// ── Security RBAC Roles ────────────────────────────────────────────────────────

type SecurityRBACRoleDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	RoleType     string
	UsersCount   int
	Permissions  string
	Scope        string
	LastModified string
	ReviewStatus string
	CreatedAt    time.Time
}

func ListSecurityRBACRoles(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityRBACRoleDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       rb.role_type, rb.users_count,
		       rb.permissions, rb.scope,
		       rb.last_modified, rb.review_status,
		       r.created_at
		FROM workspace_security_rbac_roles rb
		JOIN resources r ON r.id = rb.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityRBACRoles: %w", err)
	}
	defer rows.Close()
	var out []SecurityRBACRoleDB
	for rows.Next() {
		var v SecurityRBACRoleDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.RoleType, &v.UsersCount,
			&v.Permissions, &v.Scope,
			&v.LastModified, &v.ReviewStatus,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityRBACRole(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, roleType, permissions, scope, region string,
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
		VALUES ($1,$2,'security_rbac_role',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityRBACRole resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_rbac_roles
		    (resource_id, role_type, permissions, scope)
		VALUES ($1,$2,$3,$4)`,
		resourceID, roleType, permissions, scope)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityRBACRole rbac row: %w", err)
	}
	return resourceID, nil
}
