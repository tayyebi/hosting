package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"platform/db"
)

// ── UI-layer types (what templates/sysadmin/database_content.html expects) ──

type DBColumn struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	Default    string
}

type DBTable struct {
	Schema   string
	Name     string
	Columns  []DBColumn
	RowCount int
	SizeKB   int
}

func (t DBTable) FullName() string { return t.Schema + "." + t.Name }

type DBSchema struct {
	Name   string
	Tables []DBTable
}

type QueryResult struct {
	SQL      string
	Columns  []string
	Rows     [][]string
	RowCount int
	Duration string
	Err      string
	Executed bool
}

type DBPageData struct {
	SysadminData
	Schemas     []DBSchema
	ActiveTable string
	BrowseTable *DBTable
	BrowseRows  [][]string
	Result      *QueryResult
	DBVersion   string
	TotalTables int
	TotalSize   string
	DBError     string
	View        string // "", "erd" — selects the main-panel mode
}

// ── Introspection ↔ UI type conversion ────────────────────────

func toUISchemas(src []db.Schema) []DBSchema {
	out := make([]DBSchema, len(src))
	for i, s := range src {
		tables := make([]DBTable, len(s.Tables))
		for j, t := range s.Tables {
			cols := make([]DBColumn, len(t.Columns))
			for k, c := range t.Columns {
				cols[k] = DBColumn{
					Name:       c.Name,
					Type:       c.Type,
					Nullable:   c.Nullable,
					PrimaryKey: c.PrimaryKey,
					Default:    c.Default,
				}
			}
			tables[j] = DBTable{
				Schema:   t.Schema,
				Name:     t.Name,
				Columns:  cols,
				RowCount: int(t.RowCount),
				SizeKB:   int(t.SizeKB),
			}
		}
		out[i] = DBSchema{Name: s.Name, Tables: tables}
	}
	return out
}

func findTable(schemas []DBSchema, fullName string) *DBTable {
	parts := strings.SplitN(fullName, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	for i := range schemas {
		if schemas[i].Name != parts[0] {
			continue
		}
		for j := range schemas[i].Tables {
			if schemas[i].Tables[j].Name == parts[1] {
				return &schemas[i].Tables[j]
			}
		}
	}
	return nil
}

func countTables(schemas []DBSchema) int {
	n := 0
	for _, s := range schemas {
		n += len(s.Tables)
	}
	return n
}

// ── Handler ───────────────────────────────────────────────────

func DatabasePage(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := DBPageData{
		SysadminData: SysadminData{
			User:      user,
			Nav:       sysadminNav,
			ActiveNav: "database",
			PageTitle: "Database",
		},
	}

	if Pools == nil {
		data.DBError = "database pool is not wired — check server logs and DATABASE_URL*"
		renderDatabasePage(w, data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	schemas, err := db.ListSchemas(ctx, Pools.Admin)
	if err != nil {
		data.DBError = "introspection failed: " + err.Error()
		renderDatabasePage(w, data)
		return
	}
	data.Schemas = toUISchemas(schemas)
	data.TotalTables = countTables(data.Schemas)

	if v, err := db.DatabaseVersion(ctx, Pools.Admin); err == nil {
		data.DBVersion = v
	} else {
		data.DBVersion = "PostgreSQL"
	}
	if sz, err := db.TotalSize(ctx, Pools.Admin); err == nil {
		data.TotalSize = sz
	}

	if view := r.URL.Query().Get("view"); view == "erd" {
		data.View = "erd"
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			sql := r.FormValue("sql")
			result := db.RunAdhoc(ctx, Pools.Readonly, Pools.Admin, user.ID, sql)
			data.Result = &QueryResult{
				SQL:      result.SQL,
				Columns:  result.Columns,
				Rows:     result.Rows,
				RowCount: result.RowCount,
				Duration: result.Duration,
				Err:      result.Error,
				Executed: true,
			}
		}
	} else if tbl := r.URL.Query().Get("table"); tbl != "" {
		data.ActiveTable = tbl
		data.BrowseTable = findTable(data.Schemas, tbl)
		if data.BrowseTable != nil {
			_, rows, err := db.BrowseTable(ctx, Pools.Admin, data.BrowseTable.Schema, data.BrowseTable.Name, 100)
			if err != nil {
				data.DBError = "browse failed: " + err.Error()
			} else {
				data.BrowseRows = rows
			}
		}
	}

	renderDatabasePage(w, data)
}

func renderDatabasePage(w http.ResponseWriter, data DBPageData) {
	funcMap := template.FuncMap{
		"fmtSize": func(kb int) string {
			if kb >= 1024 {
				return fmt.Sprintf("%.1f MB", float64(kb)/1024.0)
			}
			return fmt.Sprintf("%d KB", kb)
		},
		"fmtRows": func(n int) string {
			if n >= 1000 {
				return fmt.Sprintf("%dk", n/1000)
			}
			return fmt.Sprintf("%d", n)
		},
		"rownum": func(i int) int { return i + 1 },
	}

	innerTmpl, err := template.New("database_content.html").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "sysadmin", "database_content.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "database-content", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data.SysadminData.Content = template.HTML(buf.String())

	shellTmpl, err := template.ParseFiles(
		filepath.Join("templates", "sysadmin", "base.html"),
		filepath.Join("templates", "sysadmin", "shell.html"),
	)
	if err != nil {
		http.Error(w, "Shell template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", data.SysadminData); err != nil {
		http.Error(w, "Shell render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func RegisterDatabaseRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/s/database/", requireRole(RoleSysAdmin, DatabasePage))
}
