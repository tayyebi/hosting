package handlers

import (
	"context"
	"net/http"
	"time"

	"platform/db"
)

// ── GET /api/sysadmin/logs ──────────────────────────────────

type apiLogEntry struct {
	ID        int64             `json:"id"`
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Service   string            `json:"service"`
	Message   string            `json:"message"`
	TraceID   string            `json:"trace_id"`
	Extra     map[string]string `json:"extra,omitempty"`
}

type apiLogBucket struct {
	Label string `json:"label"`
	Error int    `json:"error"`
	Warn  int    `json:"warn"`
	Info  int    `json:"info"`
	Debug int    `json:"debug"`
	Total int    `json:"total"`
}

type apiLogsResponse struct {
	Entries    []apiLogEntry  `json:"entries"`
	Buckets    []apiLogBucket `json:"buckets"`
	HistoMax   int            `json:"histo_max"`
	ErrorCount int            `json:"error_count"`
	WarnCount  int            `json:"warn_count"`
	InfoCount  int            `json:"info_count"`
	DebugCount int            `json:"debug_count"`
	Total      int            `json:"total"`
	Services   []string       `json:"services"`
	Levels     []string       `json:"levels"`
}

func apiSysadminLogsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	q := r.URL.Query()
	filterLevel := q.Get("level")
	filterSvc := q.Get("service")
	filterSearch := q.Get("q")
	filterRange := q.Get("range")
	since := rangeToDuration(filterRange)

	resp := apiLogsResponse{
		Levels: []string{"error", "warn", "info", "debug"},
	}

	if Pools == nil {
		apiOK(w, resp)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Buckets
	lookbackMins := int(since.Minutes())
	if lookbackMins < 1 {
		lookbackMins = 60
	}
	if dbBuckets, histoMax, err := db.SystemLogBuckets(ctx, Pools.Admin, 30, lookbackMins); err == nil {
		resp.Buckets = make([]apiLogBucket, len(dbBuckets))
		for i, b := range dbBuckets {
			resp.Buckets[i] = apiLogBucket{b.Label, b.Error, b.Warn, b.Info, b.Debug, b.Total}
		}
		resp.HistoMax = histoMax
	}

	// Stats
	if stats, err := db.SystemLogStats(ctx, Pools.Admin, since); err == nil {
		resp.ErrorCount = stats["error"] + stats["fatal"]
		resp.WarnCount = stats["warn"]
		resp.InfoCount = stats["info"]
		resp.DebugCount = stats["debug"]
		resp.Total = resp.ErrorCount + resp.WarnCount + resp.InfoCount + resp.DebugCount
	}

	// Services
	if svcs, err := db.SystemLogServices(ctx, Pools.Admin); err == nil {
		resp.Services = svcs
	}

	// Entries
	entries, err := db.ListSystemLogs(ctx, Pools.Admin, db.SystemLogOpts{
		Level:   filterLevel,
		Service: filterSvc,
		Search:  filterSearch,
		Since:   since,
		Limit:   150,
	})
	if err == nil {
		resp.Entries = make([]apiLogEntry, len(entries))
		for i, e := range entries {
			h := dbLogToHandler(e)
			resp.Entries[i] = apiLogEntry{
				ID:        h.ID,
				Timestamp: h.Timestamp,
				Level:     string(h.Level),
				Service:   h.Service,
				Message:   h.Message,
				TraceID:   h.TraceID,
				Extra:     h.Extra,
			}
		}
	}

	apiOK(w, resp)
}
