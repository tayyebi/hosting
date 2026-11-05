package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Column describes one column of a table as reported by information_schema.
type Column struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	Default    string
}

// Table describes a physical or partitioned table along with size + row
// estimates from pg_stat_user_tables / pg_total_relation_size.
type Table struct {
	Schema   string
	Name     string
	Columns  []Column
	RowCount int64 // live-tuple estimate, not an exact count
	SizeKB   int64
}

// FullName returns "schema.name".
func (t Table) FullName() string { return t.Schema + "." + t.Name }

// Schema is a logical grouping of tables in Postgres (e.g. "public").
type Schema struct {
	Name   string
	Tables []Table
}

// ListSchemas returns every non-system schema along with its tables,
// columns, live-row estimates, and on-disk size. Intended for the
// sysadmin Database Manager page.
//
// Runs against the admin pool so it can see every table regardless of RLS.
func ListSchemas(ctx context.Context, pool *pgxpool.Pool) ([]Schema, error) {
	// 1) schema + table list (skip TimescaleDB-style partition children —
	//    partition leaves are identifiable by pg_class.relispartition).
	rows, err := pool.Query(ctx, `
        SELECT n.nspname, c.relname,
               COALESCE(s.n_live_tup, 0)::bigint AS rows,
               (pg_total_relation_size(c.oid) / 1024)::bigint AS size_kb
          FROM pg_class c
          JOIN pg_namespace n ON n.oid = c.relnamespace
          LEFT JOIN pg_stat_user_tables s
                 ON s.relid = c.oid
         WHERE c.relkind IN ('r','p')               -- regular + partitioned parents
           AND NOT c.relispartition                  -- hide partition leaves
           AND n.nspname NOT IN ('pg_catalog','information_schema','pg_toast')
           AND n.nspname NOT LIKE 'pg_temp_%'
         ORDER BY n.nspname, c.relname`)
	if err != nil {
		return nil, fmt.Errorf("db: list tables: %w", err)
	}
	defer rows.Close()

	type key struct{ schema, name string }
	tableMap := make(map[key]*Table)
	var order []key

	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.Schema, &t.Name, &t.RowCount, &t.SizeKB); err != nil {
			return nil, fmt.Errorf("db: scan table row: %w", err)
		}
		k := key{t.Schema, t.Name}
		tableMap[k] = &t
		order = append(order, k)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 2) columns for every table in one shot.
	colRows, err := pool.Query(ctx, `
        SELECT c.table_schema, c.table_name,
               c.column_name, c.data_type, c.is_nullable = 'YES',
               COALESCE(c.column_default, ''),
               COALESCE(pk.is_pk, false)
          FROM information_schema.columns c
          LEFT JOIN LATERAL (
              SELECT true AS is_pk
                FROM information_schema.table_constraints tc
                JOIN information_schema.key_column_usage kcu
                  ON kcu.constraint_name = tc.constraint_name
                 AND kcu.table_schema    = tc.table_schema
                 AND kcu.table_name      = tc.table_name
               WHERE tc.constraint_type = 'PRIMARY KEY'
                 AND kcu.table_schema   = c.table_schema
                 AND kcu.table_name     = c.table_name
                 AND kcu.column_name    = c.column_name
               LIMIT 1
          ) pk ON true
         WHERE c.table_schema NOT IN ('pg_catalog','information_schema','pg_toast')
           AND c.table_schema NOT LIKE 'pg_temp_%'
         ORDER BY c.table_schema, c.table_name, c.ordinal_position`)
	if err != nil {
		return nil, fmt.Errorf("db: list columns: %w", err)
	}
	defer colRows.Close()

	for colRows.Next() {
		var schema, table string
		var col Column
		if err := colRows.Scan(&schema, &table, &col.Name, &col.Type, &col.Nullable, &col.Default, &col.PrimaryKey); err != nil {
			return nil, fmt.Errorf("db: scan column row: %w", err)
		}
		if t, ok := tableMap[key{schema, table}]; ok {
			t.Columns = append(t.Columns, col)
		}
	}
	if err := colRows.Err(); err != nil {
		return nil, err
	}

	// 3) assemble ordered []Schema.
	schemasOrder := []string{}
	seen := map[string]bool{}
	schemas := map[string]*Schema{}
	for _, k := range order {
		if !seen[k.schema] {
			seen[k.schema] = true
			schemasOrder = append(schemasOrder, k.schema)
			schemas[k.schema] = &Schema{Name: k.schema}
		}
		schemas[k.schema].Tables = append(schemas[k.schema].Tables, *tableMap[k])
	}
	out := make([]Schema, 0, len(schemasOrder))
	for _, name := range schemasOrder {
		out = append(out, *schemas[name])
	}
	return out, nil
}

// Relationship describes a single foreign key constraint: the source
// table/columns and the target (referenced) table/columns. Cardinality
// is always rendered 1–N in the ERD; Postgres allows compound/unique
// FKs to be 1–1 but detecting that at introspect time isn't worth the
// query complexity for the UI.
type Relationship struct {
	Name        string   // constraint name
	FromSchema  string   // source (child) schema
	FromTable   string   // source table
	FromColumns []string // source column(s), ordered
	ToSchema    string   // target (referenced) schema
	ToTable     string   // target table
	ToColumns   []string // target column(s), ordered
	OnDelete    string   // Postgres confdeltype: a|r|c|n|d
}

// ListRelationships returns every foreign-key edge in non-system schemas.
// Runs against the admin pool; used to draw the ERD in the sysadmin UI.
func ListRelationships(ctx context.Context, pool *pgxpool.Pool) ([]Relationship, error) {
	rows, err := pool.Query(ctx, `
        SELECT c.conname,
               fn.nspname AS from_schema, ft.relname AS from_table,
               (SELECT array_agg(a.attname ORDER BY u.ord)
                  FROM unnest(c.conkey) WITH ORDINALITY AS u(attnum, ord)
                  JOIN pg_attribute a ON a.attrelid = ft.oid AND a.attnum = u.attnum) AS from_cols,
               tn.nspname AS to_schema, tt.relname AS to_table,
               (SELECT array_agg(a.attname ORDER BY u.ord)
                  FROM unnest(c.confkey) WITH ORDINALITY AS u(attnum, ord)
                  JOIN pg_attribute a ON a.attrelid = tt.oid AND a.attnum = u.attnum) AS to_cols,
               c.confdeltype::text
          FROM pg_constraint c
          JOIN pg_class     ft ON ft.oid = c.conrelid
          JOIN pg_namespace fn ON fn.oid = ft.relnamespace
          JOIN pg_class     tt ON tt.oid = c.confrelid
          JOIN pg_namespace tn ON tn.oid = tt.relnamespace
         WHERE c.contype = 'f'
           AND fn.nspname NOT IN ('pg_catalog','information_schema','pg_toast')
           AND tn.nspname NOT IN ('pg_catalog','information_schema','pg_toast')
         ORDER BY fn.nspname, ft.relname, c.conname`)
	if err != nil {
		return nil, fmt.Errorf("db: list relationships: %w", err)
	}
	defer rows.Close()

	var out []Relationship
	for rows.Next() {
		var r Relationship
		if err := rows.Scan(&r.Name,
			&r.FromSchema, &r.FromTable, &r.FromColumns,
			&r.ToSchema, &r.ToTable, &r.ToColumns,
			&r.OnDelete); err != nil {
			return nil, fmt.Errorf("db: scan relationship: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// BrowseTable returns up to `limit` rows from schema.name as
// already-stringified cells (what the template expects). Runs against
// admin pool to bypass RLS — this is a sysadmin-only view.
func BrowseTable(ctx context.Context, pool *pgxpool.Pool, schema, name string, limit int) (*Table, [][]string, error) {
	if !identSafe(schema) || !identSafe(name) {
		return nil, nil, fmt.Errorf("db: invalid identifier")
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	// Re-describe so we know the column list (and return Table for the UI).
	schemas, err := ListSchemas(ctx, pool)
	if err != nil {
		return nil, nil, err
	}
	var tbl *Table
	for i := range schemas {
		if schemas[i].Name != schema {
			continue
		}
		for j := range schemas[i].Tables {
			if schemas[i].Tables[j].Name == name {
				tbl = &schemas[i].Tables[j]
				break
			}
		}
	}
	if tbl == nil {
		return nil, nil, fmt.Errorf("db: table %s.%s not found", schema, name)
	}

	q := fmt.Sprintf(`SELECT * FROM %q.%q LIMIT $1`, schema, name)
	rows, err := pool.Query(ctx, q, limit)
	if err != nil {
		return tbl, nil, fmt.Errorf("db: select %s.%s: %w", schema, name, err)
	}
	defer rows.Close()

	out := make([][]string, 0, limit)
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return tbl, nil, err
		}
		cells := make([]string, len(vals))
		for i, v := range vals {
			cells[i] = renderCell(v)
		}
		out = append(out, cells)
	}
	return tbl, out, rows.Err()
}

// DatabaseVersion returns the server_version string so the UI can display
// "PostgreSQL 16.2" etc. without hardcoding.
func DatabaseVersion(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	var full string
	if err := pool.QueryRow(ctx, `SELECT version()`).Scan(&full); err != nil {
		return "", err
	}
	// version() returns "PostgreSQL 16.4 on x86_64-..." — trim to first clause.
	for i := 0; i < len(full); i++ {
		if full[i] == ' ' && i+1 < len(full) {
			for j := i + 1; j < len(full); j++ {
				if full[j] == ' ' {
					return full[:j], nil
				}
			}
		}
	}
	return full, nil
}

// TotalSize returns the on-disk size across every non-system schema,
// formatted as "4.2 MB" / "1.7 GB".
func TotalSize(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	var bytes int64
	err := pool.QueryRow(ctx, `
        SELECT COALESCE(SUM(pg_total_relation_size(c.oid)), 0)::bigint
          FROM pg_class c
          JOIN pg_namespace n ON n.oid = c.relnamespace
         WHERE c.relkind IN ('r','p')
           AND NOT c.relispartition
           AND n.nspname NOT IN ('pg_catalog','information_schema','pg_toast')`).Scan(&bytes)
	if err != nil {
		return "", err
	}
	return humanBytes(bytes), nil
}

// identSafe rejects identifiers that contain anything other than letters,
// digits, underscore, and a single dot (Postgres allows far more but we
// don't want to risk a double-quote escape). Used as an extra guard
// before interpolating into a SELECT.
func identSafe(s string) bool {
	if s == "" || len(s) > 63 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
	}
	return true
}

func renderCell(v any) string {
	if v == nil {
		return "NULL"
	}
	switch x := v.(type) {
	case []byte:
		return string(x)
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func humanBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// Unused ident (silence pgx import check if needed elsewhere).
var _ = pgx.ErrNoRows
