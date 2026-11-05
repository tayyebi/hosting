package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantResource struct {
	ID           int64
	TenantID     int64
	Name         string
	ResourceType string
	Provider     string
	Region       string
	Status       string
	IPAddress    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ListTenantResources(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status string) ([]TenantResource, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(name) LIKE $%d OR lower(resource_type) LIKE $%d OR lower(ip_address) LIKE $%d)", n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, name, resource_type, provider, region, status, ip_address, created_at, updated_at
		FROM tenant_resources
		WHERE %s
		ORDER BY name ASC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenantResources: %w", err)
	}
	defer rows.Close()

	var out []TenantResource
	for rows.Next() {
		var r TenantResource
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Name, &r.ResourceType, &r.Provider,
			&r.Region, &r.Status, &r.IPAddress, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListTenantResources scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func UpdateTenantResourceStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenant_resources SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantResourceStatus: %w", err)
	}
	return nil
}

func DeleteTenantResource(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenant_resources SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteTenantResource: %w", err)
	}
	return nil
}

func CreateTenantResource(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, resourceType, provider, region, ipAddress string) (TenantResource, error) {
	var r TenantResource
	err := pool.QueryRow(ctx, `
		INSERT INTO tenant_resources (tenant_id, name, resource_type, provider, region, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, name, resource_type, provider, region, status, ip_address, created_at, updated_at`,
		tenantID, name, resourceType, provider, region, ipAddress,
	).Scan(&r.ID, &r.TenantID, &r.Name, &r.ResourceType, &r.Provider,
		&r.Region, &r.Status, &r.IPAddress, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return TenantResource{}, fmt.Errorf("db.CreateTenantResource: %w", err)
	}
	return r, nil
}
