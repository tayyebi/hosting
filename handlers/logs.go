package handlers

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"platform/db"
)

// ── Data structures ───────────────────────────────────────────

type LogLevel string

const (
	LevelError LogLevel = "error"
	LevelWarn  LogLevel = "warn"
	LevelInfo  LogLevel = "info"
	LevelDebug LogLevel = "debug"
)

type LogEntry struct {
	ID        int64
	Timestamp string
	Level     LogLevel
	Service   string
	Message   string
	TraceID   string
	Extra     map[string]string
}

// LogBucket is one vertical bar in the histogram (2-minute window).
type LogBucket struct {
	Label string
	Error int
	Warn  int
	Info  int
	Debug int
	Total int
}

type LogsPageData struct {
	SysadminData
	Entries      []LogEntry
	Buckets      []LogBucket
	HistoMax     int
	ErrorCount   int
	WarnCount    int
	InfoCount    int
	DebugCount   int
	Services     []string
	LevelOptions []string
	FilterLevel  string
	FilterSvc    string
	FilterSearch string
	FilterRange  string
}

// rangeToDuration converts a ?range= query parameter to a look-back duration.
func rangeToDuration(r string) time.Duration {
	switch r {
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "6h":
		return 6 * time.Hour
	case "24h":
		return 24 * time.Hour
	default:
		return time.Hour // "1h"
	}
}

// ── Handler ───────────────────────────────────────────────────

func LogsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	q := r.URL.Query()
	filterLevel := q.Get("level")
	filterSvc := q.Get("service")
	filterSearch := q.Get("q")
	filterRange := q.Get("range")
	if filterRange == "" {
		filterRange = "1h"
	}
	since := rangeToDuration(filterRange)

	data := LogsPageData{
		SysadminData: SysadminData{
			User:      user,
			Nav:       sysadminNav,
			ActiveNav: "logs",
			PageTitle: "Logs",
		},
		LevelOptions: []string{"error", "warn", "info", "debug"},
		FilterLevel:  filterLevel,
		FilterSvc:    filterSvc,
		FilterSearch: filterSearch,
		FilterRange:  filterRange,
	}

	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Histogram buckets (30 buckets over the selected range)
		lookbackMins := int(since.Minutes())
		if lookbackMins < 1 {
			lookbackMins = 60
		}
		if dbBuckets, histoMax, err := db.SystemLogBuckets(ctx, Pools.Admin, 30, lookbackMins); err == nil {
			for _, b := range dbBuckets {
				data.Buckets = append(data.Buckets, LogBucket{
					Label: b.Label,
					Error: b.Error,
					Warn:  b.Warn,
					Info:  b.Info,
					Debug: b.Debug,
					Total: b.Total,
				})
			}
			data.HistoMax = histoMax
		}

		// Level stats
		if stats, err := db.SystemLogStats(ctx, Pools.Admin, since); err == nil {
			data.ErrorCount = stats["error"] + stats["fatal"]
			data.WarnCount = stats["warn"]
			data.InfoCount = stats["info"]
			data.DebugCount = stats["debug"]
		}

		// Distinct services for filter dropdown
		if svcs, err := db.SystemLogServices(ctx, Pools.Admin); err == nil {
			data.Services = svcs
		}

		// Filtered log entries
		entries, err := db.ListSystemLogs(ctx, Pools.Admin, db.SystemLogOpts{
			Level:   filterLevel,
			Service: filterSvc,
			Search:  filterSearch,
			Since:   since,
			Limit:   150,
		})
		if err == nil {
			data.Entries = make([]LogEntry, 0, len(entries))
			for _, e := range entries {
				data.Entries = append(data.Entries, dbLogToHandler(e))
			}
		}
	}

	// Render inner content template
	funcMap := template.FuncMap{
		"pct": func(val, max int) int {
			if max == 0 {
				return 0
			}
			p := val * 100 / max
			if p < 4 && val > 0 {
				return 4
			}
			return p
		},
		"substr": func(s string, a, b int) string {
			if b > len(s) {
				b = len(s)
			}
			return s[a:b]
		},
	}
	innerTmpl, err := template.New("logs_content.html").Funcs(funcMap).ParseFiles(filepath.Join("templates", "sysadmin", "logs_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "logs-content", data); err != nil {
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

func dbLogToHandler(e db.SystemLog) LogEntry {
	level := LogLevel(strings.ToLower(e.Level))
	if level == "fatal" {
		level = LevelError
	}
	return LogEntry{
		ID:        e.ID,
		Timestamp: e.CreatedAt.Format("2006-01-02 15:04:05.000"),
		Level:     level,
		Service:   e.Service,
		Message:   e.Message,
		TraceID:   e.TraceID,
		Extra:     e.Fields,
	}
}
