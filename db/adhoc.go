package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdhocResult is what the SQL console renders.
type AdhocResult struct {
	SQL      string
	Columns  []string
	Rows     [][]string
	RowCount int
	Duration string
	Error    string
}

// RunAdhoc executes a sysadmin-supplied SQL string against the readonly
// pool with defensive guardrails:
//
//   - statement_timeout = 5s, lock_timeout = 1s (per-tx)
//   - the connection is platform_readonly (no INSERT/UPDATE/DELETE/DDL grants)
//   - multi-statement input is rejected (no `;` outside string literals)
//   - SELECT results are capped at 1000 rows client-side
//   - every invocation is recorded in audit_log
//
// actorUserID is used for the audit_log entry; it may be 0 for rare
// pre-login edge cases (not expected from sysadmin flows).
func RunAdhoc(ctx context.Context, ro *pgxpool.Pool, admin *pgxpool.Pool, actorUserID int64, sql string) *AdhocResult {
	started := time.Now()
	res := &AdhocResult{SQL: sql}

	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		res.Error = "empty query"
		_ = writeAudit(ctx, admin, actorUserID, sql, 0, 0, res.Error)
		return res
	}
	if hasMultipleStatements(trimmed) {
		res.Error = "multiple statements are not allowed in the ad-hoc console"
		_ = writeAudit(ctx, admin, actorUserID, sql, 0, 0, res.Error)
		return res
	}

	// Run in a tx so SET LOCAL is scoped and the readonly role can't be
	// tricked into committing anything even if it somehow tried.
	tx, err := ro.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		res.Error = "connect: " + err.Error()
		_ = writeAudit(ctx, admin, actorUserID, sql, 0, time.Since(started), res.Error)
		return res
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SET LOCAL statement_timeout = '5s'`); err != nil {
		res.Error = "set statement_timeout: " + err.Error()
		_ = writeAudit(ctx, admin, actorUserID, sql, 0, time.Since(started), res.Error)
		return res
	}
	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '1s'`); err != nil {
		res.Error = "set lock_timeout: " + err.Error()
		_ = writeAudit(ctx, admin, actorUserID, sql, 0, time.Since(started), res.Error)
		return res
	}
	// Sysadmin console — bypass RLS for visibility across tenants.
	if _, err := tx.Exec(ctx,
		`SELECT set_config('app.current_role', 'sysadmin', true)`); err != nil {
		res.Error = "set role: " + err.Error()
		_ = writeAudit(ctx, admin, actorUserID, sql, 0, time.Since(started), res.Error)
		return res
	}

	rows, err := tx.Query(ctx, trimmed)
	if err != nil {
		res.Error = humanisePgError(err)
		res.Duration = humanDuration(time.Since(started))
		_ = writeAudit(ctx, admin, actorUserID, sql, 0, time.Since(started), res.Error)
		return res
	}
	defer rows.Close()

	for _, fd := range rows.FieldDescriptions() {
		res.Columns = append(res.Columns, string(fd.Name))
	}

	const maxRows = 1000
	for rows.Next() {
		if len(res.Rows) >= maxRows {
			// Drain silently to honour the cap without erroring the whole query.
			break
		}
		vals, err := rows.Values()
		if err != nil {
			res.Error = "scan: " + err.Error()
			_ = writeAudit(ctx, admin, actorUserID, sql, len(res.Rows), time.Since(started), res.Error)
			return res
		}
		cells := make([]string, len(vals))
		for i, v := range vals {
			cells[i] = renderCell(v)
		}
		res.Rows = append(res.Rows, cells)
	}
	if err := rows.Err(); err != nil {
		res.Error = humanisePgError(err)
		res.Duration = humanDuration(time.Since(started))
		_ = writeAudit(ctx, admin, actorUserID, sql, len(res.Rows), time.Since(started), res.Error)
		return res
	}

	res.RowCount = len(res.Rows)
	res.Duration = humanDuration(time.Since(started))
	_ = writeAudit(ctx, admin, actorUserID, sql, res.RowCount, time.Since(started), "")
	return res
}

// hasMultipleStatements returns true when `sql` contains a semicolon
// outside of any string/identifier/line-or-block comment. A trailing
// semicolon with nothing after it is allowed.
func hasMultipleStatements(sql string) bool {
	i := 0
	n := len(sql)
	lastSemicolonAt := -1
	for i < n {
		c := sql[i]
		switch {
		case c == '\'' || c == '"':
			// Walk until the matching closer, handling '' / "" escapes.
			q := c
			i++
			for i < n {
				if sql[i] == q {
					if i+1 < n && sql[i+1] == q {
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
		case c == '-' && i+1 < n && sql[i+1] == '-':
			// Line comment.
			for i < n && sql[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < n && sql[i+1] == '*':
			// Block comment.
			i += 2
			for i+1 < n && !(sql[i] == '*' && sql[i+1] == '/') {
				i++
			}
			if i+1 < n {
				i += 2
			}
		case c == ';':
			lastSemicolonAt = i
			i++
		default:
			i++
		}
	}
	if lastSemicolonAt == -1 {
		return false
	}
	// Anything non-whitespace after the last semicolon → multi-statement.
	for j := lastSemicolonAt + 1; j < n; j++ {
		if sql[j] != ' ' && sql[j] != '\t' && sql[j] != '\n' && sql[j] != '\r' {
			return true
		}
	}
	return false
}

func writeAudit(ctx context.Context, admin *pgxpool.Pool, actorUserID int64, sql string, rowCount int, dur time.Duration, errMsg string) error {
	sev := "info"
	if errMsg != "" {
		sev = "warn"
	}
	if admin == nil {
		return errors.New("no admin pool")
	}
	_, err := admin.Exec(ctx, `
        INSERT INTO audit_log (tenant_id, actor_user_id, actor_role, action,
                               severity, details, error, duration_ms)
        VALUES (NULL, NULLIF($1, 0), 'sysadmin', 'sql.adhoc.run',
                $2, jsonb_build_object('sql', $3::text, 'rows', $4::int),
                NULLIF($5, ''), $6)`,
		actorUserID, sev, sql, rowCount, errMsg, int(dur.Milliseconds()))
	return err
}

func humaniseRejection(reason string) string { return reason }

// humanisePgError surfaces readonly-role rejections with a friendlier hint.
func humanisePgError(err error) string {
	msg := err.Error()
	// Common Postgres rejection texts when the readonly role lacks grants.
	switch {
	case strings.Contains(msg, "permission denied"):
		return "permission denied — ad-hoc console runs as platform_readonly; only SELECT is permitted"
	case strings.Contains(msg, "canceling statement due to statement timeout"):
		return "query exceeded 5s statement_timeout"
	case strings.Contains(msg, "canceling statement due to lock timeout"):
		return "query exceeded 1s lock_timeout"
	}
	return msg
}

func humanDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1 {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", ms)
}
