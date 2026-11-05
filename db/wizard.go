package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WizardDeploy is one row from wizard_deploys.
type WizardDeploy struct {
	ID          int64
	AppSlug     string
	AppName     string
	TenantID    *int64
	WorkspaceID *int64
	UserID      *int64
	Status      string
	Region      string
	Version     string
	CreatedAt   time.Time
}

// CreateWizardDeploy inserts a new deploy record and returns the created row.
func CreateWizardDeploy(ctx context.Context, pool *pgxpool.Pool, d WizardDeploy) (WizardDeploy, error) {
	err := pool.QueryRow(ctx, `
		INSERT INTO wizard_deploys (app_slug, app_name, tenant_id, workspace_id, user_id, region, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, app_slug, app_name, tenant_id, workspace_id, user_id, status, region, version, created_at`,
		d.AppSlug, d.AppName, d.TenantID, d.WorkspaceID, d.UserID, d.Region, d.Version,
	).Scan(
		&d.ID, &d.AppSlug, &d.AppName, &d.TenantID, &d.WorkspaceID, &d.UserID,
		&d.Status, &d.Region, &d.Version, &d.CreatedAt,
	)
	if err != nil {
		return d, fmt.Errorf("db.CreateWizardDeploy: %w", err)
	}
	return d, nil
}

// CountAppDeploys returns the total number of deploys for a given app slug.
func CountAppDeploys(ctx context.Context, pool *pgxpool.Pool, appSlug string) (int, error) {
	var n int
	err := pool.QueryRow(ctx, `SELECT count(*) FROM wizard_deploys WHERE app_slug=$1`, appSlug).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("db.CountAppDeploys: %w", err)
	}
	return n, nil
}

// ListRecentDeploys returns the most recent deploy records across all apps.
func ListRecentDeploys(ctx context.Context, pool *pgxpool.Pool, limit int) ([]WizardDeploy, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := pool.Query(ctx, `
		SELECT id, app_slug, app_name, tenant_id, workspace_id, user_id, status, region, version, created_at
		FROM wizard_deploys
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("db.ListRecentDeploys: %w", err)
	}
	defer rows.Close()
	var out []WizardDeploy
	for rows.Next() {
		var d WizardDeploy
		if err := rows.Scan(
			&d.ID, &d.AppSlug, &d.AppName, &d.TenantID, &d.WorkspaceID, &d.UserID,
			&d.Status, &d.Region, &d.Version, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
