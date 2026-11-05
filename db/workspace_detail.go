package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkspaceDetail is the DB-resolved workspace record.
type WorkspaceDetail struct {
	ID          int64
	Slug        string
	Name        string
	Description string
	TenantID    int64
	CreatedAt   time.Time
}

// WorkspaceStats holds live counters for the workspace overview page.
type WorkspaceStats struct {
	ResourceCount int
	ActiveCount   int
	MemberCount   int
}

// GetWorkspaceByID resolves a workspace by its integer DB id.
func GetWorkspaceByID(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) (WorkspaceDetail, error) {
	var w WorkspaceDetail
	err := pool.QueryRow(ctx, `
		SELECT id, slug, name, COALESCE(description,''), tenant_id, created_at
		FROM workspaces WHERE id=$1 AND deleted_at IS NULL`, workspaceID,
	).Scan(&w.ID, &w.Slug, &w.Name, &w.Description, &w.TenantID, &w.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return WorkspaceDetail{}, fmt.Errorf("db.GetWorkspaceByID: not found")
		}
		return WorkspaceDetail{}, fmt.Errorf("db.GetWorkspaceByID: %w", err)
	}
	return w, nil
}

// GetWorkspaceBySlug resolves a workspace by tenant + slug.
func GetWorkspaceBySlug(ctx context.Context, pool *pgxpool.Pool, tenantID int64, slug string) (WorkspaceDetail, error) {
	var w WorkspaceDetail
	err := pool.QueryRow(ctx, `
		SELECT id, slug, name, COALESCE(description,''), tenant_id, created_at
		FROM workspaces WHERE tenant_id=$1 AND slug=$2 AND deleted_at IS NULL`, tenantID, slug,
	).Scan(&w.ID, &w.Slug, &w.Name, &w.Description, &w.TenantID, &w.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return WorkspaceDetail{}, fmt.Errorf("db.GetWorkspaceBySlug: not found")
		}
		return WorkspaceDetail{}, fmt.Errorf("db.GetWorkspaceBySlug: %w", err)
	}
	return w, nil
}

// GetWorkspaceStats returns resource and member counts for the overview panel.
func GetWorkspaceStats(ctx context.Context, pool *pgxpool.Pool, workspaceID int64) (WorkspaceStats, error) {
	var s WorkspaceStats
	// Resource counts
	err := pool.QueryRow(ctx, `
		SELECT count(*),
		       count(*) FILTER (WHERE status = 'active')
		FROM resources
		WHERE workspace_id=$1 AND deleted_at IS NULL`, workspaceID,
	).Scan(&s.ResourceCount, &s.ActiveCount)
	if err != nil && err != pgx.ErrNoRows {
		return s, fmt.Errorf("db.GetWorkspaceStats resources: %w", err)
	}
	// Member count
	err = pool.QueryRow(ctx, `
		SELECT count(*) FROM workspace_memberships WHERE workspace_id=$1`, workspaceID,
	).Scan(&s.MemberCount)
	if err != nil && err != pgx.ErrNoRows {
		return s, fmt.Errorf("db.GetWorkspaceStats members: %w", err)
	}
	return s, nil
}
