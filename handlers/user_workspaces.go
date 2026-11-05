package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"platform/db"
)

// UserWorkspacesData is passed to user_workspaces.html.
type UserWorkspacesData struct {
	User        User
	Workspaces  []db.UserWorkspace
	Invitations []db.UserInvitation
}

func userWorkspacesHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := UserWorkspacesData{User: user}

	if Pools != nil && user.ID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if workspaces, err := db.ListUserWorkspaces(ctx, Pools.App, user.ID); err == nil {
			data.Workspaces = workspaces
		}
		if invitations, err := db.ListUserInvitations(ctx, Pools.App, user.Email); err == nil {
			data.Invitations = invitations
		}
	}

	if data.Workspaces == nil {
		data.Workspaces = []db.UserWorkspace{}
	}
	if data.Invitations == nil {
		data.Invitations = []db.UserInvitation{}
	}

	tmpl, err := template.New("user_workspaces.html").Funcs(template.FuncMap{
		"slice": func(s string, a, b int) string {
			if b > len(s) {
				b = len(s)
			}
			return s[a:b]
		},
		"fmtDate": func(t time.Time) string {
			return t.Format("Jan 2006")
		},
		"lenSlice": func(s []db.UserWorkspace) int { return len(s) },
		"wsColor": func(i int) string {
			colors := []string{"accent", "amber", "green", "indigo", "cyan", "pink", "yellow"}
			return colors[i%len(colors)]
		},
	}).ParseFiles(filepath.Join("templates", "user_workspaces.html"))
	if err != nil {
		ServeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tmpl.ExecuteTemplate(w, "user-workspaces", data); err != nil {
		ServeError(w, r, http.StatusInternalServerError, err.Error())
	}
}

type createWorkspaceRequest struct {
	Name       string `json:"name"`
	Desc       string `json:"description"`
	Visibility string `json:"visibility"`
}

func userWorkspacesCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON.")
		return
	}

	if user.ID > 0 {
		key := fmt.Sprintf("%d", user.ID)
		if !workspaceCreateRateLimiter.allow(key) {
			w.Header().Set("Retry-After", "3600")
			apiError(w, http.StatusTooManyRequests, "Workspace creation rate limit exceeded. Try again later.")
			return
		}
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Workspace name is required.")
		return
	}
	if len(name) > 80 {
		apiError(w, http.StatusBadRequest, "Workspace name must be 80 characters or fewer.")
		return
	}
	desc := strings.TrimSpace(req.Desc)
	if len(desc) > 500 {
		apiError(w, http.StatusBadRequest, "Description must be 500 characters or fewer.")
		return
	}

	if Pools != nil && user.ID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		count, err := db.CountUserWorkspaces(ctx, Pools.App, user.ID)
		if err == nil && count >= 10 {
			apiError(w, http.StatusUnprocessableEntity, "Workspace limit reached. You may not create more than 10 workspaces.")
			return
		}

		if exists, err := db.WorkspaceNameExistsForUser(ctx, Pools.App, user.ID, name); err == nil && exists {
			apiError(w, http.StatusUnprocessableEntity, "A workspace with that name already exists.")
			return
		}

		if err := db.CreateUserWorkspace(ctx, Pools.App, user.ID, name, desc); err != nil {
			apiError(w, http.StatusInternalServerError, "Failed to create workspace.")
			return
		}
		go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
			"workspace.created", "info", map[string]any{"name": name},
			"", w.Header().Get("X-Request-ID"), fmt.Sprintf("%d", user.ID))
	}

	apiOK(w, map[string]string{"status": "ok", "name": name})
}

func userWorkspacesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	idStr := r.PathValue("id")
	var wsID int64
	if _, err := fmt.Sscanf(idStr, "%d", &wsID); err != nil || wsID <= 0 {
		apiError(w, http.StatusBadRequest, "Invalid workspace ID.")
		return
	}

	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	deleted, err := db.DeleteUserWorkspace(ctx, Pools.App, user.ID, wsID)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to delete workspace.")
		return
	}
	if !deleted {
		apiError(w, http.StatusNotFound, "Workspace not found or you are not the owner.")
		return
	}

	go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role),
		"workspace.deleted", "warn", map[string]any{"workspace_id": wsID},
		"", r.Header.Get("X-Request-ID"), clientIP(r))

	apiOK(w, map[string]string{"status": "ok"})
}

func RegisterUserWorkspacesRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /u/workspaces/", userWorkspacesHandler)
	mux.HandleFunc("POST /u/workspaces/create/", userWorkspacesCreateHandler)
	mux.HandleFunc("DELETE /u/workspaces/{id}", userWorkspacesDeleteHandler)
}
