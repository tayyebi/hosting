package handlers

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"platform/db"
)

// ── data structures ───────────────────────────────────────────

type ReplicaHealth string

const (
	ReplicaHealthHealthy      ReplicaHealth = "healthy"
	ReplicaHealthSyncDelayed  ReplicaHealth = "sync-delayed"
	ReplicaHealthOverloaded   ReplicaHealth = "overloaded"
	ReplicaHealthOutOfService ReplicaHealth = "out-of-service"
	ReplicaHealthPending      ReplicaHealth = "pending"
)

func (h ReplicaHealth) Label() string {
	switch h {
	case ReplicaHealthHealthy:
		return "Healthy"
	case ReplicaHealthSyncDelayed:
		return "Sync Delayed"
	case ReplicaHealthOverloaded:
		return "Overloaded"
	case ReplicaHealthOutOfService:
		return "Out of Service"
	default:
		return "Pending"
	}
}

type Replica struct {
	ID      string // replica_id (UUID string)
	Host    string
	Region  string
	Health  ReplicaHealth
	AddedAt time.Time
}

// merchantCode returns a stable platform-wide code for replica on-boarding.
// Reads PLATFORM_MERCHANT_CODE env var; falls back to a fixed demo value.
func getMerchantCode() string {
	if v := os.Getenv("PLATFORM_MERCHANT_CODE"); v != "" {
		return v
	}
	return "plat-demo-a1b2c3d4e5f6g7h8"
}

// ── DB ↔ handler conversion ───────────────────────────────────

func dbReplicaToHandler(r db.Replica) Replica {
	return Replica{
		ID:      r.ReplicaID,
		Host:    r.Host,
		Region:  r.Region,
		Health:  ReplicaHealth(r.Health),
		AddedAt: r.AddedAt,
	}
}

// ── page data ─────────────────────────────────────────────────

type replicasInnerData struct {
	Replicas     []Replica
	MerchantCode string
	FormError    string
}

func listReplicasFromDB(ctx context.Context) ([]Replica, error) {
	if Pools == nil {
		return nil, nil
	}
	rows, err := db.ListReplicas(ctx, Pools.Admin)
	if err != nil {
		return nil, err
	}
	out := make([]Replica, len(rows))
	for i, r := range rows {
		out[i] = dbReplicaToHandler(r)
	}
	return out, nil
}

func renderReplicasPage(w http.ResponseWriter, r *http.Request, formErr string) {
	user, _ := getSession(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	replicas, _ := listReplicasFromDB(ctx)

	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "sysadmin", "replicas_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "replicas-content", replicasInnerData{
		Replicas:     replicas,
		MerchantCode: getMerchantCode(),
		FormError:    formErr,
	}); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shellPaths := []string{
		filepath.Join("templates", "sysadmin", "base.html"),
		filepath.Join("templates", "sysadmin", "shell.html"),
	}
	shellTmpl, err := template.ParseFiles(shellPaths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := SysadminData{
		User:      user,
		Nav:       sysadminNav,
		ActiveNav: "replicas",
		PageTitle: "Replicas",
		Content:   template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "sysadmin-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── handlers ─────────────────────────────────────────────────

func replicasListHandler(w http.ResponseWriter, r *http.Request) {
	renderReplicasPage(w, r, "")
}

func replicasOnboardHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	host := strings.TrimSpace(r.FormValue("host"))
	region := strings.TrimSpace(r.FormValue("region"))
	if host == "" {
		renderReplicasPage(w, r, "Host address is required.")
		return
	}

	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		exists, err := db.ReplicaHostExists(ctx, Pools.Admin, host)
		if err == nil && exists {
			renderReplicasPage(w, r, "A replica with that host is already registered.")
			return
		}
		if _, err := db.AddReplica(ctx, Pools.Admin, host, region); err != nil {
			renderReplicasPage(w, r, "Failed to register replica: "+err.Error())
			return
		}
	}

	http.Redirect(w, r, "/s/replicas/", http.StatusSeeOther)
}

func replicasDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_ = db.DeleteReplica(ctx, Pools.Admin, id)
	}
	http.Redirect(w, r, "/s/replicas/", http.StatusSeeOther)
}

func RegisterReplicaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /s/replicas/", requireRole(RoleSysAdmin, replicasListHandler))
	mux.HandleFunc("POST /s/replicas/", requireRole(RoleSysAdmin, replicasOnboardHandler))
	mux.HandleFunc("POST /s/replicas/{id}/delete", requireRole(RoleSysAdmin, replicasDeleteHandler))
	mux.HandleFunc("POST /s/replicas/{id}/delete/", requireRole(RoleSysAdmin, replicasDeleteHandler))
}
