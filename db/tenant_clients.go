package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantClient is one row from tenant_clients joined with its workspace names.
type TenantClient struct {
	ID         int64
	TenantID   int64
	Name       string
	Email      string
	Company    string
	Status     string
	CreatedAt  time.Time
	Workspaces []string
}

// ListTenantClients returns all non-deleted clients for the given tenant,
// with optional search (name/email/company) and status filter.
// Workspaces slice is populated for each client.
func ListTenantClients(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status string) ([]TenantClient, error) {
	args := []any{tenantID}
	conds := []string{"c.tenant_id = $1", "c.deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(c.name) LIKE $%d OR lower(c.email) LIKE $%d OR lower(c.company) LIKE $%d)",
			n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("c.status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT c.id, c.tenant_id, c.name, c.email, c.company, c.status, c.created_at,
		       COALESCE(array_agg(cw.name ORDER BY cw.created_at) FILTER (WHERE cw.id IS NOT NULL), '{}')
		FROM tenant_clients c
		LEFT JOIN tenant_client_workspaces cw ON cw.client_id = c.id
		WHERE %s
		GROUP BY c.id
		ORDER BY c.created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListTenantClients: %w", err)
	}
	defer rows.Close()

	var out []TenantClient
	for rows.Next() {
		var c TenantClient
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Company,
			&c.Status, &c.CreatedAt, &c.Workspaces); err != nil {
			return nil, fmt.Errorf("db.ListTenantClients scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetTenantClient returns a single client by ID, scoped to the tenant.
func GetTenantClient(ctx context.Context, pool *pgxpool.Pool, tenantID, clientID int64) (TenantClient, error) {
	var c TenantClient
	err := pool.QueryRow(ctx, `
		SELECT c.id, c.tenant_id, c.name, c.email, c.company, c.status, c.created_at,
		       COALESCE(array_agg(cw.name ORDER BY cw.created_at) FILTER (WHERE cw.id IS NOT NULL), '{}')
		FROM tenant_clients c
		LEFT JOIN tenant_client_workspaces cw ON cw.client_id = c.id
		WHERE c.tenant_id = $1 AND c.id = $2 AND c.deleted_at IS NULL
		GROUP BY c.id`, tenantID, clientID,
	).Scan(&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Company,
		&c.Status, &c.CreatedAt, &c.Workspaces)
	if err != nil {
		return TenantClient{}, fmt.Errorf("db.GetTenantClient: %w", err)
	}
	return c, nil
}

// CreateTenantClient inserts a new client and a default workspace for them.
func CreateTenantClient(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, email, company string) (TenantClient, error) {
	var c TenantClient
	err := pool.QueryRow(ctx, `
		INSERT INTO tenant_clients (tenant_id, name, email, company)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, name, email, company, status, created_at`,
		tenantID, name, email, company,
	).Scan(&c.ID, &c.TenantID, &c.Name, &c.Email, &c.Company, &c.Status, &c.CreatedAt)
	if err != nil {
		return TenantClient{}, fmt.Errorf("db.CreateTenantClient: %w", err)
	}

	// Insert a default workspace for the new client.
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_client_workspaces (client_id, name) VALUES ($1, $2)
		 ON CONFLICT (client_id, name) DO NOTHING`,
		c.ID, "default")
	if err != nil {
		return TenantClient{}, fmt.Errorf("db.CreateTenantClient (default workspace): %w", err)
	}
	c.Workspaces = []string{"default"}
	return c, nil
}

// UpdateTenantClient updates name, email, and company for a client.
func UpdateTenantClient(ctx context.Context, pool *pgxpool.Pool, tenantID, clientID int64, name, email, company string) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenant_clients
		SET name=$3, email=$4, company=$5, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, clientID, name, email, company)
	if err != nil {
		return fmt.Errorf("db.UpdateTenantClient: %w", err)
	}
	return nil
}

// SetTenantClientStatus sets the status field (active / suspended).
func SetTenantClientStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, clientID int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenant_clients
		SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, clientID, status)
	if err != nil {
		return fmt.Errorf("db.SetTenantClientStatus: %w", err)
	}
	return nil
}

// DeleteTenantClient soft-deletes a client by setting deleted_at.
func DeleteTenantClient(ctx context.Context, pool *pgxpool.Pool, tenantID, clientID int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE tenant_clients
		SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, clientID)
	if err != nil {
		return fmt.Errorf("db.DeleteTenantClient: %w", err)
	}
	return nil
}

// CreateClientWorkspace adds a named workspace to a client.
func CreateClientWorkspace(ctx context.Context, pool *pgxpool.Pool, clientID int64, name string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO tenant_client_workspaces (client_id, name) VALUES ($1, $2)
		 ON CONFLICT (client_id, name) DO NOTHING`,
		clientID, name)
	if err != nil {
		return fmt.Errorf("db.CreateClientWorkspace: %w", err)
	}
	return nil
}

// TenantClientCount returns the number of non-deleted clients for a tenant.
func TenantClientCount(ctx context.Context, pool *pgxpool.Pool, tenantID int64) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT count(*) FROM tenant_clients WHERE tenant_id=$1 AND deleted_at IS NULL`,
		tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.TenantClientCount: %w", err)
	}
	return count, nil
}
