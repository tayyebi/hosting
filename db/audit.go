package db

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PruneOldAuditLogs deletes audit_log rows older than retentionDays days.
// Configurable via AUDIT_LOG_RETENTION_DAYS (default 90). Runs as a daily job.
func PruneOldAuditLogs(ctx context.Context, admin *pgxpool.Pool, retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 90
	}
	_, err := admin.Exec(ctx,
		`DELETE FROM audit_log WHERE created_at < now() - ($1 * interval '1 day')`,
		retentionDays)
	return err
}

// InsertAuditEvent writes a single row to audit_log using the admin pool.
// severity is one of "info", "warn", "critical".
// details is an arbitrary map that will be serialised to JSONB.
// requestID and ipAddr are optional — pass "" to omit.
func InsertAuditEvent(ctx context.Context, admin *pgxpool.Pool, userID int64, role, action, severity string, details map[string]any, errMsg, requestID, ipAddr string) {
	if admin == nil {
		return
	}
	var detailsJSON []byte
	if len(details) > 0 {
		detailsJSON, _ = json.Marshal(details)
	}
	_, _ = admin.Exec(ctx, `
		INSERT INTO audit_log (tenant_id, actor_user_id, actor_role, action,
		                       severity, details, error, request_id, ip_address)
		VALUES (NULL, NULLIF($1, 0), $2, $3, $4,
		        $5::jsonb, NULLIF($6, ''), NULLIF($7,''), NULLIF($8,'')::inet)`,
		userID, role, action, severity, detailsJSON, errMsg, requestID, ipAddr)
}
