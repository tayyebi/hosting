package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaxRule struct {
	ID         int64
	TenantID   int64
	Name       string
	Region     string
	TaxType    string
	Status     string
	Rate       float64
	AppliesTo  string
	ValidFrom  *time.Time
	ValidUntil *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func ListTaxRules(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, taxType, status string) ([]TaxRule, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(name) LIKE $%d OR lower(region) LIKE $%d OR lower(applies_to) LIKE $%d)", n, n, n))
	}
	if taxType != "" {
		args = append(args, taxType)
		conds = append(conds, fmt.Sprintf("tax_type = $%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, name, region, tax_type, status, rate, applies_to, valid_from, valid_until, created_at, updated_at
		FROM tax_rules
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListTaxRules: %w", err)
	}
	defer rows.Close()

	var out []TaxRule
	for rows.Next() {
		var t TaxRule
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Region, &t.TaxType,
			&t.Status, &t.Rate, &t.AppliesTo, &t.ValidFrom, &t.ValidUntil,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListTaxRules scan: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func CreateTaxRule(ctx context.Context, pool *pgxpool.Pool, tenantID int64, name, region, taxType, appliesTo string, rate float64, validFrom, validUntil *time.Time) (TaxRule, error) {
	var t TaxRule
	err := pool.QueryRow(ctx, `
		INSERT INTO tax_rules (tenant_id, name, region, tax_type, rate, applies_to, valid_from, valid_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, name, region, tax_type, status, rate, applies_to, valid_from, valid_until, created_at, updated_at`,
		tenantID, name, region, taxType, rate, appliesTo, validFrom, validUntil,
	).Scan(&t.ID, &t.TenantID, &t.Name, &t.Region, &t.TaxType,
		&t.Status, &t.Rate, &t.AppliesTo, &t.ValidFrom, &t.ValidUntil,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return TaxRule{}, fmt.Errorf("db.CreateTaxRule: %w", err)
	}
	return t, nil
}

func GetTaxRule(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) (TaxRule, error) {
	var t TaxRule
	err := pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, region, tax_type, status, rate, applies_to, valid_from, valid_until, created_at, updated_at
		FROM tax_rules
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`,
		tenantID, id,
	).Scan(&t.ID, &t.TenantID, &t.Name, &t.Region, &t.TaxType,
		&t.Status, &t.Rate, &t.AppliesTo, &t.ValidFrom, &t.ValidUntil,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return TaxRule{}, fmt.Errorf("db.GetTaxRule: %w", err)
	}
	return t, nil
}

func DeleteTaxRule(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE tax_rules SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteTaxRule: %w", err)
	}
	return nil
}
