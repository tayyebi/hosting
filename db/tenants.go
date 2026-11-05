package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LookupTenantIDBySlug resolves a tenant slug to its numeric ID.
// Returns pgx.ErrNoRows if the tenant does not exist or is deleted.
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
