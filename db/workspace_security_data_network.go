package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Security: Encryption at Rest ───────────────────────────────────────────────

type SecurityEncRestDB struct {
	ResourceID   int64
	Name         string
	Status       string
	Region       string
	ResourceType string
	Algorithm    string
	KMSKey       string
	KeyAge       string
	LastVerified string
	Compliance   string
	CreatedAt    time.Time
}

func ListSecurityEncRest(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityEncRestDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       e.resource_type, e.algorithm, e.kms_key,
		       e.key_age, e.last_verified, e.compliance,
		       r.created_at
		FROM workspace_security_data_enc_rest e
		JOIN resources r ON r.id = e.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityEncRest: %w", err)
	}
	defer rows.Close()
	var out []SecurityEncRestDB
	for rows.Next() {
		var v SecurityEncRestDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.ResourceType, &v.Algorithm, &v.KMSKey,
			&v.KeyAge, &v.LastVerified, &v.Compliance,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityEncRest(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, resourceType, algorithm, kmsKey, region string,
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
		VALUES ($1,$2,'security_enc_rest',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityEncRest resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_data_enc_rest
		    (resource_id, resource_type, algorithm, kms_key)
		VALUES ($1,$2,$3,$4)`,
		resourceID, resourceType, algorithm, kmsKey)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityEncRest row: %w", err)
	}
	return resourceID, nil
}

// ── Security: Encryption in Transit ───────────────────────────────────────────

type SecurityEncTransitDB struct {
	ResourceID int64
	Name       string
	Status     string
	Region     string
	Protocol   string
	Cipher     string
	CertIssuer string
	Grade      string
	Expires    string
	HSTS       string
	LastScan   string
	CreatedAt  time.Time
}

func ListSecurityEncTransit(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityEncTransitDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       t.protocol, t.cipher, t.cert_issuer,
		       t.grade, t.expires, t.hsts, t.last_scan,
		       r.created_at
		FROM workspace_security_data_enc_transit t
		JOIN resources r ON r.id = t.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityEncTransit: %w", err)
	}
	defer rows.Close()
	var out []SecurityEncTransitDB
	for rows.Next() {
		var v SecurityEncTransitDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Protocol, &v.Cipher, &v.CertIssuer,
			&v.Grade, &v.Expires, &v.HSTS, &v.LastScan,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityEncTransit(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, protocol, cipher, certIssuer, region string,
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
		VALUES ($1,$2,'security_enc_transit',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityEncTransit resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_data_enc_transit
		    (resource_id, protocol, cipher, cert_issuer)
		VALUES ($1,$2,$3,$4)`,
		resourceID, protocol, cipher, certIssuer)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityEncTransit row: %w", err)
	}
	return resourceID, nil
}

// ── Security: KMS Keys ─────────────────────────────────────────────────────────

type SecurityKMSKeyDB struct {
	ResourceID    int64
	Name          string
	Status        string
	Region        string
	Algorithm     string
	KeyType       string
	AgeDays       string
	Rotation      string
	Operations24h string
	GrantsCount   int
	CreatedAt     time.Time
}

func ListSecurityKMSKeys(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityKMSKeyDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       k.algorithm, k.key_type, k.age_days,
		       k.rotation, k.operations_24h, k.grants_count,
		       r.created_at
		FROM workspace_security_data_kms k
		JOIN resources r ON r.id = k.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityKMSKeys: %w", err)
	}
	defer rows.Close()
	var out []SecurityKMSKeyDB
	for rows.Next() {
		var v SecurityKMSKeyDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Algorithm, &v.KeyType, &v.AgeDays,
			&v.Rotation, &v.Operations24h, &v.GrantsCount,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityKMSKey(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, algorithm, keyType, rotation, region string,
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
		VALUES ($1,$2,'security_kms_key',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityKMSKey resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_data_kms
		    (resource_id, algorithm, key_type, rotation)
		VALUES ($1,$2,$3,$4)`,
		resourceID, algorithm, keyType, rotation)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityKMSKey row: %w", err)
	}
	return resourceID, nil
}

// ── Security: WAF Rules ────────────────────────────────────────────────────────

type SecurityWAFRuleDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Priority       int
	Action         string
	Source         string
	Matches24h     string
	FalsePositives int
	Severity       string
	LastTriggered  string
	CreatedAt      time.Time
}

func ListSecurityWAFRules(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityWAFRuleDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       w.priority, w.action, w.source,
		       w.matches_24h, w.false_positives, w.severity, w.last_triggered,
		       r.created_at
		FROM workspace_security_network_waf w
		JOIN resources r ON r.id = w.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY w.priority ASC, r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityWAFRules: %w", err)
	}
	defer rows.Close()
	var out []SecurityWAFRuleDB
	for rows.Next() {
		var v SecurityWAFRuleDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Priority, &v.Action, &v.Source,
			&v.Matches24h, &v.FalsePositives, &v.Severity, &v.LastTriggered,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityWAFRule(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, action, source, severity string,
	priority int,
) (int64, error) {
	var resourceID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, workspace_id, kind, name, status, created_by)
		VALUES ($1,$2,'security_waf_rule',$3,'active',$4)
		RETURNING id`,
		tenantID, workspaceID, name, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityWAFRule resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_network_waf
		    (resource_id, priority, action, source, severity)
		VALUES ($1,$2,$3,$4,$5)`,
		resourceID, priority, action, source, severity)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityWAFRule row: %w", err)
	}
	return resourceID, nil
}

// ── Security: DDoS Protection Profiles ────────────────────────────────────────

type SecurityDDoSProfileDB struct {
	ResourceID     int64
	Name           string
	Status         string
	Region         string
	Method         string
	Threshold      string
	Attacks30d     string
	LastTriggered  string
	FalsePositives int
	CreatedAt      time.Time
}

func ListSecurityDDoSProfiles(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) ([]SecurityDDoSProfileDB, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.name, r.status,
		       COALESCE(reg.slug, ''),
		       d.method, d.threshold, d.attacks_30d,
		       d.last_triggered, d.false_positives,
		       r.created_at
		FROM workspace_security_network_ddos d
		JOIN resources r ON r.id = d.resource_id
		LEFT JOIN regions reg ON reg.id = r.region_id
		WHERE r.workspace_id = $1
		  AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, fmt.Errorf("db.ListSecurityDDoSProfiles: %w", err)
	}
	defer rows.Close()
	var out []SecurityDDoSProfileDB
	for rows.Next() {
		var v SecurityDDoSProfileDB
		if err := rows.Scan(
			&v.ResourceID, &v.Name, &v.Status, &v.Region,
			&v.Method, &v.Threshold, &v.Attacks30d,
			&v.LastTriggered, &v.FalsePositives,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func CreateSecurityDDoSProfile(ctx context.Context, pool *pgxpool.Pool,
	tenantID, workspaceID, createdBy int64,
	name, method, threshold, region string,
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
		VALUES ($1,$2,'security_ddos_profile',$3,'active',$4,$5)
		RETURNING id`,
		tenantID, workspaceID, name, regionID, createdBy,
	).Scan(&resourceID)
	if err != nil {
		return 0, fmt.Errorf("db.CreateSecurityDDoSProfile resource: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_security_network_ddos
		    (resource_id, method, threshold)
		VALUES ($1,$2,$3)`,
		resourceID, method, threshold)
	if err != nil {
		return resourceID, fmt.Errorf("db.CreateSecurityDDoSProfile row: %w", err)
	}
	return resourceID, nil
}
