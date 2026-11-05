package handlers

import (
	"context"
	"net/http"
	"time"

	"platform/db"
)

// ── GET /api/sysadmin/replicas ─────────────────────────────

type apiReplica struct {
	ID      string `json:"id"`
	Host    string `json:"host"`
	Region  string `json:"region"`
	Health  string `json:"health"`
	Label   string `json:"health_label"`
	AddedAt string `json:"added_at"`
}

type apiReplicasResponse struct {
	Replicas     []apiReplica `json:"replicas"`
	Total        int          `json:"total"`
	MerchantCode string       `json:"merchant_code"`
	Healthy      int          `json:"healthy_count"`
	Degraded     int          `json:"degraded_count"`
}

func apiSysadminReplicasHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok || user.Role != RoleSysAdmin {
		apiError(w, http.StatusForbidden, "Sysadmin access required.")
		return
	}

	resp := apiReplicasResponse{MerchantCode: getMerchantCode()}

	if Pools == nil {
		apiOK(w, resp)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := db.ListReplicas(ctx, Pools.Admin)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to load replicas.")
		return
	}

	var healthy, degraded int
	out := make([]apiReplica, len(rows))
	for i, rp := range rows {
		h := ReplicaHealth(rp.Health)
		out[i] = apiReplica{
			ID:      rp.ReplicaID,
			Host:    rp.Host,
			Region:  rp.Region,
			Health:  rp.Health,
			Label:   h.Label(),
			AddedAt: rp.AddedAt.Format("2006-01-02 15:04:05"),
		}
		if h == ReplicaHealthHealthy {
			healthy++
		} else {
			degraded++
		}
	}

	apiOK(w, apiReplicasResponse{
		Replicas:     out,
		Total:        len(out),
		MerchantCode: getMerchantCode(),
		Healthy:      healthy,
		Degraded:     degraded,
	})
}
