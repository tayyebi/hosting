package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestContext carries the identity the database layer needs to set its
// per-transaction GUCs so Row-Level Security policies match on the right
// rows. Zero values mean "unauthenticated / system"; an RLS-enforced
// table will return zero rows in that case (default-deny).
type RequestContext struct {
	UserID   int64
	TenantID int64  // 0 for sysadmin sessions not scoped to a tenant
	Role     string // "sysadmin" | "tenant" | "user" | "system"
}

// WithTx acquires a connection from pool, opens a transaction, sets the
// request's RLS GUCs via SET LOCAL (so they evaporate on commit/rollback),
// and runs fn. On error the tx is rolled back.
//
// fn receives the pgx.Tx and should use it for every subsequent query so
// the SET LOCAL values apply.
func WithTx(ctx context.Context, pool *pgxpool.Pool, rc RequestContext, fn func(pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := applyRLSGUCs(ctx, tx, rc); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// applyRLSGUCs writes the three GUCs read by the policy helpers
// public.current_tenant_id() and public.current_actor_role() and the
// users-table policy that reads app.current_user_id.
func applyRLSGUCs(ctx context.Context, tx pgx.Tx, rc RequestContext) error {
	// set_config() takes text; empty string is honoured (NULL-equivalent
	// for our helpers). Using set_config vs SET LOCAL avoids the
	// awkwardness of SET LOCAL not accepting parameters.
	if _, err := tx.Exec(ctx,
		`SELECT set_config('app.current_user_id',   $1::text, true),
                set_config('app.current_tenant_id', $2::text, true),
                set_config('app.current_role',      $3::text, true)`,
		stringifyID(rc.UserID), stringifyID(rc.TenantID), rc.Role,
	); err != nil {
		return fmt.Errorf("db: set RLS GUCs: %w", err)
	}
	return nil
}

func stringifyID(id int64) string {
	if id <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", id)
}
