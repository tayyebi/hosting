package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantEmployee represents a user's membership in a tenant with employee extras.
type TenantEmployee struct {
	UserID       int64
	MembershipID int64
	Name         string
	Email        string
	Title        string
	Active       bool
	Permissions  map[string]bool
	CreatedAt    time.Time
}

// ListTenantEmployees returns all membership rows for the tenant joined with user data.
func ListTenantEmployees(ctx context.Context, pool *pgxpool.Pool, tenantID int64) ([]TenantEmployee, error) {
	rows, err := pool.Query(ctx, `
		SELECT m.id, u.id, COALESCE(u.name,''), u.email,
		       COALESCE(m.title,''), m.active, m.permissions, m.created_at
		FROM user_tenant_memberships m
		JOIN users u ON u.id = m.user_id
		WHERE m.tenant_id = $1
		ORDER BY m.created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenantEmployees: %w", err)
	}
	defer rows.Close()

	var out []TenantEmployee
	for rows.Next() {
		var e TenantEmployee
		var permRaw []byte
		if err := rows.Scan(&e.MembershipID, &e.UserID, &e.Name, &e.Email,
			&e.Title, &e.Active, &permRaw, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("db.ListTenantEmployees scan: %w", err)
		}
		if len(permRaw) > 0 {
			if err := json.Unmarshal(permRaw, &e.Permissions); err != nil {
				e.Permissions = map[string]bool{}
			}
		}
		if e.Permissions == nil {
			e.Permissions = map[string]bool{}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetTenantEmployee returns a single membership row by membership ID, scoped to tenant.
func GetTenantEmployee(ctx context.Context, pool *pgxpool.Pool, tenantID, membershipID int64) (TenantEmployee, error) {
	var e TenantEmployee
	var permRaw []byte
	err := pool.QueryRow(ctx, `
		SELECT m.id, u.id, COALESCE(u.name,''), u.email,
		       COALESCE(m.title,''), m.active, m.permissions, m.created_at
		FROM user_tenant_memberships m
		JOIN users u ON u.id = m.user_id
		WHERE m.tenant_id = $1 AND m.id = $2`,
		tenantID, membershipID,
	).Scan(&e.MembershipID, &e.UserID, &e.Name, &e.Email,
		&e.Title, &e.Active, &permRaw, &e.CreatedAt)
	if err != nil {
		return TenantEmployee{}, fmt.Errorf("db.GetTenantEmployee: %w", err)
	}
	if len(permRaw) > 0 {
		if err := json.Unmarshal(permRaw, &e.Permissions); err != nil {
			e.Permissions = map[string]bool{}
		}
	}
	if e.Permissions == nil {
		e.Permissions = map[string]bool{}
	}
	return e, nil
}

// CreateTenantEmployee upserts the user record and creates a tenant membership.
func CreateTenantEmployee(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, email, title string) (TenantEmployee, error) {
	userID, err := UpsertUser(ctx, pool, email, name, "user")
	if err != nil {
		return TenantEmployee{}, fmt.Errorf("db.CreateTenantEmployee (upsert user): %w", err)
	}

	var e TenantEmployee
	err = pool.QueryRow(ctx, `
		INSERT INTO user_tenant_memberships (user_id, tenant_id, role, title)
		VALUES ($1, $2, 'member', $3)
		ON CONFLICT (user_id, tenant_id) DO UPDATE
		    SET title = EXCLUDED.title
		RETURNING id, created_at`,
		userID, tenantID, title,
	).Scan(&e.MembershipID, &e.CreatedAt)
	if err != nil {
		return TenantEmployee{}, fmt.Errorf("db.CreateTenantEmployee (membership): %w", err)
	}

	e.UserID = userID
	e.Name = name
	e.Email = email
	e.Title = title
	e.Active = true
	e.Permissions = map[string]bool{}
	return e, nil
}

// UpdateTenantEmployee updates name/email on the users row and title/active on the membership.
func UpdateTenantEmployee(ctx context.Context, pool *pgxpool.Pool, tenantID, membershipID int64, name, email, title string, active bool) error {
	// Resolve user_id from membership.
	var userID int64
	err := pool.QueryRow(ctx,
		`SELECT user_id FROM user_tenant_memberships WHERE id=$1 AND tenant_id=$2`,
		membershipID, tenantID).Scan(&userID)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantEmployee (resolve user): %w", err)
	}

	_, err = pool.Exec(ctx,
		`UPDATE users SET name=$2, email=$3 WHERE id=$1`,
		userID, name, email)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantEmployee (users): %w", err)
	}

	_, err = pool.Exec(ctx,
		`UPDATE user_tenant_memberships SET title=$3, active=$4 WHERE id=$1 AND tenant_id=$2`,
		membershipID, tenantID, title, active)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantEmployee (membership): %w", err)
	}
	return nil
}

// DeleteTenantEmployee removes the membership row (hard delete).
func DeleteTenantEmployee(ctx context.Context, pool *pgxpool.Pool, tenantID, membershipID int64) error {
	_, err := pool.Exec(ctx,
		`DELETE FROM user_tenant_memberships WHERE id=$1 AND tenant_id=$2`,
		membershipID, tenantID)
	if err != nil {
		return fmt.Errorf("db.DeleteTenantEmployee: %w", err)
	}
	return nil
}

// UpdateTenantEmployeePermissions replaces the permissions JSONB for a membership.
func UpdateTenantEmployeePermissions(ctx context.Context, pool *pgxpool.Pool, tenantID, membershipID int64, permissions map[string]bool) error {
	raw, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantEmployeePermissions (marshal): %w", err)
	}
	_, err = pool.Exec(ctx,
		`UPDATE user_tenant_memberships SET permissions=$3 WHERE id=$1 AND tenant_id=$2`,
		membershipID, tenantID, raw)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantEmployeePermissions: %w", err)
	}
	return nil
}

// TenantEmployeeCount returns the number of memberships for a tenant.
func TenantEmployeeCount(ctx context.Context, pool *pgxpool.Pool, tenantID int64) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT count(*) FROM user_tenant_memberships WHERE tenant_id=$1`,
		tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.TenantEmployeeCount: %w", err)
	}
	return count, nil
}
