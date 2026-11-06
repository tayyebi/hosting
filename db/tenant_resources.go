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
	APIURL       string
	AuthMethod   string
	Credential   string
	Tags         string
	CPUCores     int
	RAMMB        int
	DiskGB       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateTenantResourceInput struct {
	TenantID     int64
	Name         string
	ResourceType string
	Provider     string
	Region       string
	Status       string
	IPAddress    string
	APIURL       string
	AuthMethod   string
	Credential   string
	Tags         string
	CPUCores     int
	RAMMB        int
	DiskGB       int
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
		SELECT id, tenant_id, name, resource_type, provider, region, status, ip_address,
		       COALESCE(api_url, ''), COALESCE(auth_method, ''), COALESCE(credential, ''),
		       COALESCE(tags, ''), COALESCE(cpu_cores, 0), COALESCE(ram_mb, 0), COALESCE(disk_gb, 0),
		       created_at, updated_at
		FROM tenant_resources
		WHERE %s
		ORDER BY id ASC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenantResources: %w", err)
	}
	defer rows.Close()

	var out []TenantResource
	for rows.Next() {
		var r TenantResource
		if err := rows.Scan(
			&r.ID, &r.TenantID, &r.Name, &r.ResourceType, &r.Provider,
			&r.Region, &r.Status, &r.IPAddress,
			&r.APIURL, &r.AuthMethod, &r.Credential,
			&r.Tags, &r.CPUCores, &r.RAMMB, &r.DiskGB,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
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
	return CreateTenantResourceFull(ctx, pool, CreateTenantResourceInput{
		TenantID:     tenantID,
		Name:         name,
		ResourceType: resourceType,
		Provider:     provider,
		Region:       region,
		Status:       "online",
		IPAddress:    ipAddress,
	})
}

func CreateTenantResourceFull(ctx context.Context, pool *pgxpool.Pool, in CreateTenantResourceInput) (TenantResource, error) {
	st := strings.TrimSpace(in.Status)
	if st == "" {
		st = "online"
	}
	var r TenantResource
	err := pool.QueryRow(ctx, `
		INSERT INTO tenant_resources (
			tenant_id, name, resource_type, provider, region, status, ip_address,
			api_url, auth_method, credential, tags, cpu_cores, ram_mb, disk_gb
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, tenant_id, name, resource_type, provider, region, status, ip_address,
		          api_url, auth_method, credential, tags, cpu_cores, ram_mb, disk_gb,
		          created_at, updated_at`,
		in.TenantID, strings.TrimSpace(in.Name), strings.TrimSpace(in.ResourceType),
		strings.TrimSpace(in.Provider), strings.TrimSpace(in.Region), st, strings.TrimSpace(in.IPAddress),
		strings.TrimSpace(in.APIURL), strings.TrimSpace(in.AuthMethod), in.Credential,
		strings.TrimSpace(in.Tags), in.CPUCores, in.RAMMB, in.DiskGB,
	).Scan(
		&r.ID, &r.TenantID, &r.Name, &r.ResourceType, &r.Provider,
		&r.Region, &r.Status, &r.IPAddress,
		&r.APIURL, &r.AuthMethod, &r.Credential,
		&r.Tags, &r.CPUCores, &r.RAMMB, &r.DiskGB,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return TenantResource{}, fmt.Errorf("db.CreateTenantResourceFull: %w", err)
	}
	return r, nil
}
