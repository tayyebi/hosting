package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IDSEvent struct {
	ID        int64
	TenantID  int64
	Severity  string
	EventType string
	SourceIP  string
	Target    string
	Protocol  string
	Action    string
	Status    string
	RuleID    string
	Details   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ListIDSEvents(ctx context.Context, pool *pgxpool.Pool, tenantID int64, search, severity, status string) ([]IDSEvent, error) {
	args := []any{tenantID}
	conds := []string{"tenant_id = $1", "deleted_at IS NULL"}

	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"(lower(event_type) LIKE $%d OR lower(source_ip) LIKE $%d OR lower(target) LIKE $%d OR lower(rule_id) LIKE $%d)", n, n, n, n))
	}
	if severity != "" {
		args = append(args, severity)
		conds = append(conds, fmt.Sprintf("severity = $%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	q := fmt.Sprintf(`
		SELECT id, tenant_id, severity, event_type, source_ip, target, protocol, action, status, rule_id, details, created_at, updated_at
		FROM ids_events
		WHERE %s
		ORDER BY created_at DESC`, strings.Join(conds, " AND "))

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListIDSEvents: %w", err)
	}
	defer rows.Close()

	var out []IDSEvent
	for rows.Next() {
		var e IDSEvent
		if err := rows.Scan(&e.ID, &e.TenantID, &e.Severity, &e.EventType, &e.SourceIP,
			&e.Target, &e.Protocol, &e.Action, &e.Status, &e.RuleID, &e.Details,
			&e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db.ListIDSEvents scan: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func UpdateIDSEventStatus(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE ids_events SET status=$3, updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id, status)
	if err != nil {
		return fmt.Errorf("db.UpdateIDSEventStatus: %w", err)
	}
	return nil
}

func DeleteIDSEvent(ctx context.Context, pool *pgxpool.Pool, tenantID, id int64) error {
	_, err := pool.Exec(ctx, `
		UPDATE ids_events SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
		tenantID, id)
	if err != nil {
		return fmt.Errorf("db.DeleteIDSEvent: %w", err)
	}
	return nil
}

func CreateIDSEvent(ctx context.Context, pool *pgxpool.Pool, tenantID int64, severity, eventType, sourceIP, target, protocol, action, ruleID, details string) (IDSEvent, error) {
	var e IDSEvent
	err := pool.QueryRow(ctx, `
		INSERT INTO ids_events (tenant_id, severity, event_type, source_ip, target, protocol, action, rule_id, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, severity, event_type, source_ip, target, protocol, action, status, rule_id, details, created_at, updated_at`,
		tenantID, severity, eventType, sourceIP, target, protocol, action, ruleID, details,
	).Scan(&e.ID, &e.TenantID, &e.Severity, &e.EventType, &e.SourceIP, &e.Target,
		&e.Protocol, &e.Action, &e.Status, &e.RuleID, &e.Details, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return IDSEvent{}, fmt.Errorf("db.CreateIDSEvent: %w", err)
	}
	return e, nil
}
