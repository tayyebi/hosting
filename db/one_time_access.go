package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OneTimeAccess struct {
	ID        int64
	TenantID  int64
	Label     string
	Token     string
	Target    string
	GrantedTo string
	GrantedBy string
	Status    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func ListOneTimeAccess(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, status string) ([]OneTimeAccess, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(label) LIKE $%d OR lower(granted_to) LIKE $%d OR lower(target) LIKE $%d)", n, n, n))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, label, token, target, granted_to, granted_by, status, expires_at, created_at
		FROM one_time_access
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListOneTimeAccess: %w", err)
	}
	defer rows.Close()

	var out []OneTimeAccess
	for rows.Next() {
		var r OneTimeAccess
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Label, &r.Token, &r.Target,
			&r.GrantedTo, &r.GrantedBy, &r.Status, &r.ExpiresAt, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("db.ListOneTimeAccess scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func RevokeOneTimeAccess(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE one_time_access SET status='revoked'
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL AND status='active'`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.RevokeOneTimeAccess: %w", err)
	}
	return nil
}

func DeleteOneTimeAccess(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE one_time_access SET deleted_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteOneTimeAccess: %w", err)
	}
	return nil
}

func CreateOneTimeAccess(ctx context.Context, pool *pgxpool.Pool, tenantID int64, label, token, target, grantedTo, grantedBy string, ttlHours int) (OneTimeAccess, error) {
	var r OneTimeAccess
	err := pool.QueryRow(ctx, `
		INSERT INTO one_time_access (tenant_id, label, token, target, granted_to, granted_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, now() + ($7 || ' hours')::interval)
		RETURNING id, tenant_id, label, token, target, granted_to, granted_by, status, expires_at, created_at`,
		tenantID, label, token, target, grantedTo, grantedBy, ttlHours,
	).Scan(&r.ID, &r.TenantID, &r.Label, &r.Token, &r.Target,
		&r.GrantedTo, &r.GrantedBy, &r.Status, &r.ExpiresAt, &r.CreatedAt)
	if err != nil {
		return OneTimeAccess{}, fmt.Errorf("db.CreateOneTimeAccess: %w", err)
	}
	return r, nil
}
