package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"platform/db"
)

// ── GET /api/sysadmin/database ──────────────────────────────

type apiDBColumn struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	PrimaryKey bool   `json:"primary_key"`
	Default    string `json:"default,omitempty"`
}

type apiDBTable struct {
	Schema   string        `json:"schema"`
	Name     string        `json:"name"`
	FullName string        `json:"full_name"`
	Columns  []apiDBColumn `json:"columns"`
	RowCount int           `json:"row_count"`
	SizeKB   int           `json:"size_kb"`
}

type apiDBSchema struct {
	Name   string       `json:"name"`
	Tables []apiDBTable `json:"tables"`
}

type apiDatabaseResponse struct {
	Schemas     []apiDBSchema `json:"schemas"`
	DBVersion   string        `json:"db_version"`
	TotalTables int           `json:"total_tables"`
	TotalSize   string        `json:"total_size"`
}

func apiSysadminDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database pool is not wired.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	schemas, err := db.ListSchemas(ctx, Pools.Admin)
	if err != nil {
		log.Printf("sysadmin database introspection error: %v", err)
		apiError(w, http.StatusInternalServerError, "Database introspection failed.")
		return
	}
	out := make([]apiDBSchema, len(schemas))
	totalTables := 0
	for i, s := range schemas {
		tables := make([]apiDBTable, len(s.Tables))
		for j, t := range s.Tables {
			cols := make([]apiDBColumn, len(t.Columns))
			for k, c := range t.Columns {
				cols[k] = apiDBColumn{
					Name: c.Name, Type: c.Type, Nullable: c.Nullable,
					PrimaryKey: c.PrimaryKey, Default: c.Default,
				}
			}
			tables[j] = apiDBTable{
				Schema: t.Schema, Name: t.Name, FullName: t.FullName(),
				Columns: cols, RowCount: int(t.RowCount), SizeKB: int(t.SizeKB),
			}
			totalTables++
		}
		out[i] = apiDBSchema{Name: s.Name, Tables: tables}
	}

	version, _ := db.DatabaseVersion(ctx, Pools.Admin)
	totalSize, _ := db.TotalSize(ctx, Pools.Admin)

	apiOK(w, apiDatabaseResponse{
		Schemas:     out,
		DBVersion:   version,
		TotalTables: totalTables,
		TotalSize:   totalSize,
	})
}

// ── GET /api/sysadmin/database/table?name={schema.table} ────

type apiTableBrowseResponse struct {
	Table apiDBTable `json:"table"`
	Rows  [][]string `json:"rows"`
	Total int        `json:"total"`
}

func apiSysadminDatabaseTableHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database pool is not wired.")
		return
	}

	tableName := r.URL.Query().Get("name")
	if tableName == "" {
		apiError(w, http.StatusBadRequest, "Query parameter 'name' is required (e.g. public.tenants).")
		return
	}
	schema, name := splitTableName(tableName)
	if schema == "" || name == "" {
		apiError(w, http.StatusBadRequest, "Invalid table name; expected schema.table.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tbl, rows, err := db.BrowseTable(ctx, Pools.Admin, schema, name, 100)
	if err != nil {
		log.Printf("sysadmin table browse error (%s.%s): %v", schema, name, err)
		apiError(w, http.StatusNotFound, "Table not found or not accessible.")
		return
	}

	cols := make([]apiDBColumn, len(tbl.Columns))
	for i, c := range tbl.Columns {
		cols[i] = apiDBColumn{
			Name: c.Name, Type: c.Type, Nullable: c.Nullable,
			PrimaryKey: c.PrimaryKey, Default: c.Default,
		}
	}
	apiOK(w, apiTableBrowseResponse{
		Table: apiDBTable{
			Schema: tbl.Schema, Name: tbl.Name, FullName: tbl.FullName(),
			Columns: cols, RowCount: int(tbl.RowCount), SizeKB: int(tbl.SizeKB),
		},
		Rows:  rows,
		Total: len(rows),
	})
}

// ── POST /api/sysadmin/database/query ───────────────────────

type apiQueryRequest struct {
	SQL string `json:"sql"`
}

type apiQueryResponse struct {
	SQL      string     `json:"sql"`
	Columns  []string   `json:"columns,omitempty"`
	Rows     [][]string `json:"rows,omitempty"`
	RowCount int        `json:"row_count"`
	Duration string     `json:"duration"`
	Error    string     `json:"error,omitempty"`
}

func apiSysadminDatabaseQueryHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database pool is not wired.")
		return
	}

	var req apiQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	result := db.RunAdhoc(ctx, Pools.Readonly, Pools.Admin, user.ID, req.SQL)
	apiOK(w, apiQueryResponse{
		SQL:      result.SQL,
		Columns:  result.Columns,
		Rows:     result.Rows,
		RowCount: result.RowCount,
		Duration: result.Duration,
		Error:    result.Error,
	})
}

// ── GET /api/sysadmin/database/erd ──────────────────────────
//
// Returns every non-system table together with its column list and
// every foreign-key edge between them. The Database Manager uses this
// to render a schema diagram in the browser.

type apiERDTable struct {
	Schema   string        `json:"schema"`
	Name     string        `json:"name"`
	FullName string        `json:"full_name"`
	Columns  []apiDBColumn `json:"columns"`
}

type apiERDRelationship struct {
	Name        string   `json:"name"`
	FromTable   string   `json:"from_table"` // "schema.name"
	FromColumns []string `json:"from_columns"`
	ToTable     string   `json:"to_table"` // "schema.name"
	ToColumns   []string `json:"to_columns"`
	OnDelete    string   `json:"on_delete"` // a|r|c|n|d
}

type apiERDResponse struct {
	Tables        []apiERDTable        `json:"tables"`
	Relationships []apiERDRelationship `json:"relationships"`
}

func apiSysadminDatabaseERDHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}
	if Pools == nil {
		apiError(w, http.StatusServiceUnavailable, "Database pool is not wired.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	schemas, err := db.ListSchemas(ctx, Pools.Admin)
	if err != nil {
		log.Printf("sysadmin ERD schema introspection error: %v", err)
		apiError(w, http.StatusInternalServerError, "Database introspection failed.")
		return
	}
	rels, err := db.ListRelationships(ctx, Pools.Admin)
	if err != nil {
		log.Printf("sysadmin ERD relationship introspection error: %v", err)
		apiError(w, http.StatusInternalServerError, "Relationship introspection failed.")
		return
	}

	var tables []apiERDTable
	for _, s := range schemas {
		for _, t := range s.Tables {
			cols := make([]apiDBColumn, len(t.Columns))
			for i, c := range t.Columns {
				cols[i] = apiDBColumn{
					Name: c.Name, Type: c.Type, Nullable: c.Nullable,
					PrimaryKey: c.PrimaryKey, Default: c.Default,
				}
			}
			tables = append(tables, apiERDTable{
				Schema:   t.Schema,
				Name:     t.Name,
				FullName: t.FullName(),
				Columns:  cols,
			})
		}
	}

	edges := make([]apiERDRelationship, 0, len(rels))
	for _, r := range rels {
		edges = append(edges, apiERDRelationship{
			Name:        r.Name,
			FromTable:   r.FromSchema + "." + r.FromTable,
			FromColumns: r.FromColumns,
			ToTable:     r.ToSchema + "." + r.ToTable,
			ToColumns:   r.ToColumns,
			OnDelete:    r.OnDelete,
		})
	}

	apiOK(w, apiERDResponse{Tables: tables, Relationships: edges})
}

func splitTableName(s string) (schema, name string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return s[:i], s[i+1:]
		}
	}
	return "", ""
}
