package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

var serverStart = time.Now()

// HealthzHandler returns a simple health check response. It is intentionally
// unauthenticated so that load balancers and uptime monitors can probe it.
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	type response struct {
		Status  string `json:"status"`
		Uptime  int64  `json:"uptime_seconds"`
		DB      string `json:"db"`
	}

	resp := response{
		Status: "ok",
		Uptime: int64(time.Since(serverStart).Seconds()),
		DB:     "ok",
	}

	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := Pools.Admin.Ping(ctx); err != nil {
			resp.Status = "degraded"
			resp.DB = "unreachable"
		}
	} else {
		resp.DB = "not connected"
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Prevent caching of health check responses.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if resp.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}
