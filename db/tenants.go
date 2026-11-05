package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRow represents one active row in the `tenants` table.
type TenantRow struct {
	ID         int64
	Slug       string
	Name       string
	Domain     string
	Plan       string
	OwnerName  string
	OwnerEmail string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// LookupTenantIDBySlug resolves a tenant slug to its numeric ID.
func LookupTenantIDBySlug(ctx context.Context, pool *pgxpool.Pool, slug string) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx,
		`SELECT id FROM tenants WHERE slug=$1 AND deleted_at IS NULL`,
		slug).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("db.LookupTenantIDBySlug: %w", err)
	}
	return id, nil
}

// ListTenants returns all non-deleted tenants ordered by creation time.
func ListTenants(ctx context.Context, pool *pgxpool.Pool) ([]TenantRow, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, slug, name, COALESCE(domain, ''), plan,
		       COALESCE(owner_name, ''), COALESCE(owner_email, ''),
		       created_at, updated_at
		FROM tenants
		WHERE deleted_at IS NULL
		ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenants: %w", err)
	}
	defer rows.Close()

	var out []TenantRow
	for rows.Next() {
		var t TenantRow
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.Domain, &t.Plan,
			&t.OwnerName, &t.OwnerEmail, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListTenants scan: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetTenantByID fetches a single non-deleted tenant by numeric ID.
func GetTenantByID(ctx context.Context, pool *pgxpool.Pool, id int64) (TenantRow, bool, error) {
	var t TenantRow
	err := pool.QueryRow(ctx, `
		SELECT id, slug, name, COALESCE(domain, ''), plan,
		       COALESCE(owner_name, ''), COALESCE(owner_email, ''),
		       created_at, updated_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&t.ID, &t.Slug, &t.Name, &t.Domain, &t.Plan,
		&t.OwnerName, &t.OwnerEmail, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantRow{}, false, nil
	}
	if err != nil {
		return TenantRow{}, false, fmt.Errorf("db.GetTenantByID: %w", err)
	}
	return t, true, nil
}

// GetTenantBySlug fetches a single non-deleted tenant by slug.
func GetTenantBySlug(ctx context.Context, pool *pgxpool.Pool, slug string) (TenantRow, bool, error) {
	var t TenantRow
	err := pool.QueryRow(ctx, `
		SELECT id, slug, name, COALESCE(domain, ''), plan,
		       COALESCE(owner_name, ''), COALESCE(owner_email, ''),
		       created_at, updated_at
		FROM tenants
		WHERE slug = $1 AND deleted_at IS NULL`, slug,
	).Scan(&t.ID, &t.Slug, &t.Name, &t.Domain, &t.Plan,
		&t.OwnerName, &t.OwnerEmail, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantRow{}, false, nil
	}
	if err != nil {
		return TenantRow{}, false, fmt.Errorf("db.GetTenantBySlug: %w", err)
	}
	return t, true, nil
}

// TenantSlugExists checks whether another live tenant already holds slug.
func TenantSlugExists(ctx context.Context, pool *pgxpool.Pool, slug string, excludeID int64) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM tenants
			WHERE slug = $1 AND id <> $2 AND deleted_at IS NULL
		)`, slug, excludeID).Scan(&exists)
	return exists, err
}

// CreateTenant inserts a new tenant row in Postgres.
func CreateTenant(ctx context.Context, pool *pgxpool.Pool, name, slug, domain string) (TenantRow, error) {
	var domVal any
	if strings.TrimSpace(domain) != "" {
		domVal = strings.TrimSpace(domain)
	}
	var t TenantRow
	err := pool.QueryRow(ctx, `
		INSERT INTO tenants (name, slug, domain, plan)
		VALUES ($1, $2, $3, 'pro')
		RETURNING id, slug, name, COALESCE(domain, ''), plan,
		          COALESCE(owner_name, ''), COALESCE(owner_email, ''),
		          created_at, updated_at`,
		strings.TrimSpace(name), strings.TrimSpace(slug), domVal,
	).Scan(&t.ID, &t.Slug, &t.Name, &t.Domain, &t.Plan,
		&t.OwnerName, &t.OwnerEmail, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return TenantRow{}, fmt.Errorf("db.CreateTenant: %w", err)
	}
	return t, nil
}

// UpdateTenant updates name, slug, and domain for an existing tenant.
func UpdateTenant(ctx context.Context, pool *pgxpool.Pool, id int64, name, slug, domain string) (TenantRow, bool, error) {
	var domVal any
	if strings.TrimSpace(domain) != "" {
		domVal = strings.TrimSpace(domain)
	}
	var t TenantRow
	err := pool.QueryRow(ctx, `
		UPDATE tenants
		SET name = $2, slug = $3, domain = $4, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, slug, name, COALESCE(domain, ''), plan,
		          COALESCE(owner_name, ''), COALESCE(owner_email, ''),
		          created_at, updated_at`,
		id, strings.TrimSpace(name), strings.TrimSpace(slug), domVal,
	).Scan(&t.ID, &t.Slug, &t.Name, &t.Domain, &t.Plan,
		&t.OwnerName, &t.OwnerEmail, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantRow{}, false, nil
	}
	if err != nil {
		return TenantRow{}, false, fmt.Errorf("db.UpdateTenant: %w", err)
	}
	return t, true, nil
}

// DeleteTenant soft-deletes a tenant by numeric ID.
func DeleteTenant(ctx context.Context, pool *pgxpool.Pool, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenants
		SET deleted_at = now(), domain = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("db.DeleteTenant: %w", err)
	}
	return nil
}

// SetTenantOwner updates owner_name + owner_email on the tenant and ensures a
// corresponding `users` + `user_tenant_memberships` record exists.
func SetTenantOwner(ctx context.Context, pool *pgxpool.Pool, tenantID int64, ownerName, ownerEmail string) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenants
		SET owner_name = $2, owner_email = $3, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL`,
		tenantID, strings.TrimSpace(ownerName), strings.TrimSpace(ownerEmail))
	if err != nil {
		return fmt.Errorf("db.SetTenantOwner: %w", err)
	}

	userID, err := UpsertUser(ctx, pool, strings.TrimSpace(ownerEmail), strings.TrimSpace(ownerName), "user")
	if err != nil {
		return fmt.Errorf("db.SetTenantOwner upsert user: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO user_tenant_memberships (user_id, tenant_id, role)
		VALUES ($1, $2, 'tenant')
		ON CONFLICT (user_id, tenant_id) DO UPDATE SET role = 'tenant'`,
		userID, tenantID)
	return err
}

// EnsureWorkspaceRow guarantees a `workspaces` table row exists for (tenantID, slug, name)
// and returns its numeric ID so workspace-scoped resource operations always have a valid FK.
func EnsureWorkspaceRow(ctx context.Context, pool *pgxpool.Pool, tenantID int64, slug, name string) (int64, error) {
	if slug == "" {
		slug = "default"
	}
	if name == "" {
		name = slug
	}
	var id int64
	err := pool.QueryRow(ctx, `
		INSERT INTO workspaces (tenant_id, slug, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, slug) DO UPDATE SET
		    deleted_at = NULL,
		    updated_at = now()
		RETURNING id`,
		tenantID, slug, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("db.EnsureWorkspaceRow: %w", err)
	}
	return id, nil
}

// WorkspaceSummary represents a `workspaces` table row scoped to a tenant.
type WorkspaceSummary struct {
	ID   int64
	Slug string
	Name string
}

// ListWorkspacesByTenantID returns all non-deleted workspaces belonging to tenantID.
func ListWorkspacesByTenantID(ctx context.Context, pool *pgxpool.Pool, tenantID int64) ([]WorkspaceSummary, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, slug, name
		FROM workspaces
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY id ASC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("db.ListWorkspacesByTenantID: %w", err)
	}
	defer rows.Close()
	var out []WorkspaceSummary
	for rows.Next() {
		var w WorkspaceSummary
		if err := rows.Scan(&w.ID, &w.Slug, &w.Name); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

