package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SystemLog is one row from the system_logs table.
type SystemLog struct {
	ID        int64
	Level     string
	Service   string
	Message   string
	TraceID   string
	Fields    map[string]string
	CreatedAt time.Time
}

// SystemLogBucket is one time-window bar in the histogram.
type SystemLogBucket struct {
	Label string
	Error int
	Warn  int
	Info  int
	Debug int
	Total int
}

// SystemLogOpts controls filtering for ListSystemLogs.
type SystemLogOpts struct {
	Level   string
	Service string
	Search  string
	Since   time.Duration // look-back window, default 1h
	Limit   int
}

// ListSystemLogs returns log entries from system_logs with optional filters,
// ordered newest-first.
func ListSystemLogs(ctx context.Context, pool *pgxpool.Pool, opts SystemLogOpts) ([]SystemLog, error) {
	if opts.Since == 0 {
		opts.Since = time.Hour
	}
	if opts.Limit <= 0 || opts.Limit > 500 {
		opts.Limit = 150
	}

	args := []any{time.Now().Add(-opts.Since)}
	conds := []string{"created_at >= $1"}

	if opts.Level != "" {
		args = append(args, opts.Level)
		conds = append(conds, fmt.Sprintf("level = $%d", len(args)))
	}
	if opts.Service != "" {
		args = append(args, opts.Service)
		conds = append(conds, fmt.Sprintf("service = $%d", len(args)))
	}
	if opts.Search != "" {
		args = append(args, "%"+strings.ToLower(opts.Search)+"%")
		n := len(args)
		conds = append(conds, fmt.Sprintf("(lower(message) LIKE $%d OR lower(service) LIKE $%d)", n, n))
	}

	q := fmt.Sprintf(`
		SELECT id, level, service, message, COALESCE(trace_id,''), fields, created_at
		FROM system_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT %d`, strings.Join(conds, " AND "), opts.Limit)

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db.ListSystemLogs: %w", err)
	}
	defer rows.Close()

	var out []SystemLog
	for rows.Next() {
		var l SystemLog
		var fields map[string]any
		if err := rows.Scan(&l.ID, &l.Level, &l.Service, &l.Message, &l.TraceID, &fields, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("db.ListSystemLogs scan: %w", err)
		}
		l.Fields = make(map[string]string, len(fields))
		for k, v := range fields {
			l.Fields[k] = fmt.Sprintf("%v", v)
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// SystemLogStats returns row counts grouped by log level for the given window.
func SystemLogStats(ctx context.Context, pool *pgxpool.Pool, since time.Duration) (map[string]int, error) {
	if since == 0 {
		since = time.Hour
	}
	rows, err := pool.Query(ctx, `
		SELECT level, count(*)::int
		FROM system_logs
		WHERE created_at >= $1
		GROUP BY level`, time.Now().Add(-since))
	if err != nil {
		return nil, fmt.Errorf("db.SystemLogStats: %w", err)
	}
	defer rows.Close()

	m := map[string]int{}
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		m[level] = count
	}
	return m, rows.Err()
}

// SystemLogServices returns distinct service names present in the table.
func SystemLogServices(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT DISTINCT service FROM system_logs ORDER BY service`)
	if err != nil {
		return nil, fmt.Errorf("db.SystemLogServices: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// SystemLogBuckets aggregates log counts into nbuckets equal time windows
// covering the last lookbackMins minutes. Returns buckets oldest-first.
func SystemLogBuckets(ctx context.Context, pool *pgxpool.Pool, nbuckets, lookbackMins int) ([]SystemLogBucket, int, error) {
	bucketSecs := (lookbackMins * 60) / nbuckets

	// Pull raw (bucket_index, level, count) from DB.
	rows, err := pool.Query(ctx, `
		SELECT
			floor(extract(epoch FROM (now() - created_at)) / $1)::int AS bkt,
			level,
			count(*)::int AS cnt
		FROM system_logs
		WHERE created_at >= now() - ($2 * interval '1 minute')
		GROUP BY 1, 2`,
		bucketSecs, lookbackMins)
	if err != nil {
		return nil, 0, fmt.Errorf("db.SystemLogBuckets: %w", err)
	}
	defer rows.Close()

	type cell struct{ bkt int; level string; cnt int }
	var cells []cell
	for rows.Next() {
		var c cell
		if err := rows.Scan(&c.bkt, &c.level, &c.cnt); err != nil {
			return nil, 0, err
		}
		cells = append(cells, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Build buckets slice (oldest-first, bkt index 0 = most recent).
	now := time.Now()
	buckets := make([]SystemLogBucket, nbuckets)
	for i := range buckets {
		// label: time at start of this bucket (oldest bucket first in the slice)
		age := time.Duration(nbuckets-i) * time.Duration(bucketSecs) * time.Second
		buckets[i].Label = now.Add(-age).Format("15:04")
	}

	for _, c := range cells {
		if c.bkt < 0 || c.bkt >= nbuckets {
			continue
		}
		idx := nbuckets - 1 - c.bkt // bkt=0 (newest) → last slot
		switch c.level {
		case "error", "fatal":
			buckets[idx].Error += c.cnt
		case "warn":
			buckets[idx].Warn += c.cnt
		case "info":
			buckets[idx].Info += c.cnt
		case "debug":
			buckets[idx].Debug += c.cnt
		}
		buckets[idx].Total += c.cnt
	}

	histoMax := 1
	for _, b := range buckets {
		if b.Total > histoMax {
			histoMax = b.Total
		}
	}
	return buckets, histoMax, nil
}

// WriteSystemLog inserts one row into system_logs. Fire-and-forget safe —
// errors are swallowed by the caller.
func WriteSystemLog(ctx context.Context, pool *pgxpool.Pool, level, service, message string, fields map[string]any) error {
	if fields == nil {
		fields = map[string]any{}
	}
	_, err := pool.Exec(ctx,
		`INSERT INTO system_logs (level, service, message, fields) VALUES ($1,$2,$3,$4)`,
		level, service, message, fields)
	if err != nil {
		return fmt.Errorf("db.WriteSystemLog: %w", err)
	}
	return nil
}
