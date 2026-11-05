package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PricingPlan struct {
	ID        int64
	TenantID  int64
	Name      string
	Tier      string
	Status    string
	Price     float64
	Currency  string
	Interval  string
	Features  string
	MaxUsers  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ListPricingPlans(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, tier, status string) ([]PricingPlan, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(name) LIKE $%d OR lower(features) LIKE $%d)", n, n))
	}
	if tier != "" {
		args = append(args, tier)
		conds = append(conds, fmt.Sprintf("tier = $%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, name, tier, status, price, currency, interval, features, max_users, created_at, updated_at
		FROM pricing_plans
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListPricingPlans: %w", err)
	}
	defer rows.Close()

	var out []PricingPlan
	for rows.Next() {
		var p PricingPlan
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Tier, &p.Status,
			&p.Price, &p.Currency, &p.Interval, &p.Features, &p.MaxUsers,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListPricingPlans scan: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func CreatePricingPlan(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, tier, currency, interval, features string, price float64, maxUsers int) (PricingPlan, error) {
	var p PricingPlan
	err := pool.QueryRow(ctx, `
		INSERT INTO pricing_plans (tenant_id, name, tier, price, currency, interval, features, max_users)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, name, tier, status, price, currency, interval, features, max_users, created_at, updated_at`,
		tenantID, name, tier, price, currency, interval, features, maxUsers,
	).Scan(&p.ID, &p.TenantID, &p.Name, &p.Tier, &p.Status,
		&p.Price, &p.Currency, &p.Interval, &p.Features, &p.MaxUsers,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return PricingPlan{}, fmt.Errorf("db.CreatePricingPlan: %w", err)
	}
	return p, nil
}

func GetPricingPlan(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) (PricingPlan, error) {
	var p PricingPlan
	err := pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, tier, status, price, currency, interval, features, max_users, created_at, updated_at
		FROM pricing_plans
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`,
		tenantID, id,
	).Scan(&p.ID, &p.TenantID, &p.Name, &p.Tier, &p.Status,
		&p.Price, &p.Currency, &p.Interval, &p.Features, &p.MaxUsers,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return PricingPlan{}, fmt.Errorf("db.GetPricingPlan: %w", err)
	}
	return p, nil
}

func DeletePricingPlan(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE pricing_plans SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeletePricingPlan: %w", err)
	}
	return nil
}
