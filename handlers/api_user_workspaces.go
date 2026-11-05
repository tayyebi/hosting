package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"platform/db"
)

// ── GET /api/user/workspaces ──────────────────────────────

type apiWorkspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Tenant      string `json:"tenant_slug"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
}

type apiTeam struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members int      `json:"member_count"`
	Role    string   `json:"role"`
	Tags    []string `json:"tags"`
}

type apiInvitation struct {
	ID        string `json:"id"`
	Workspace string `json:"workspace"`
	InvitedBy string `json:"invited_by"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
}

type apiPermission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Granted  bool   `json:"granted"`
}

type apiUserWorkspacesResponse struct {
	Workspaces  []apiWorkspace  `json:"workspaces"`
	Teams       []apiTeam       `json:"teams"`
	Invitations []apiInvitation `json:"invitations"`
	Permissions []apiPermission `json:"permissions"`
}

func apiUserWorkspacesHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	resp := apiUserWorkspacesResponse{
		Teams: []apiTeam{},
		Permissions: []apiPermission{
			{Resource: "compute/vms", Action: "create", Granted: true},
			{Resource: "compute/vms", Action: "delete", Granted: true},
			{Resource: "storage/buckets", Action: "create", Granted: true},
			{Resource: "storage/buckets", Action: "delete", Granted: false},
			{Resource: "networking/vpc", Action: "create", Granted: true},
			{Resource: "networking/vpc", Action: "delete", Granted: false},
			{Resource: "billing/invoices", Action: "view", Granted: true},
			{Resource: "billing/invoices", Action: "export", Granted: false},
		},
	}

	if Pools != nil && user.ID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if workspaces, err := db.ListUserWorkspaces(ctx, Pools.App, user.ID); err == nil {
			resp.Workspaces = make([]apiWorkspace, len(workspaces))
			for i, ws := range workspaces {
				resp.Workspaces[i] = apiWorkspace{
					ID:          fmt.Sprintf("%d", ws.ID),
					Name:        ws.Name,
					Description: ws.Description,
					Tenant:      ws.TenantSlug,
					Role:        ws.Role,
					CreatedAt:   ws.CreatedAt.Format(time.RFC3339),
				}
			}
		}

		if invitations, err := db.ListUserInvitations(ctx, Pools.App, user.Email); err == nil {
			resp.Invitations = make([]apiInvitation, len(invitations))
			for i, inv := range invitations {
				resp.Invitations[i] = apiInvitation{
					ID:        fmt.Sprintf("%d", inv.ID),
					Workspace: inv.TenantSlug,
					InvitedBy: inv.InviterEmail,
					Role:      inv.Role,
					Status:    "pending",
					ExpiresAt: inv.ExpiresAt.Format(time.RFC3339),
				}
			}
		}
	}

	// Ensure non-nil slices for JSON output
	if resp.Workspaces == nil {
		resp.Workspaces = []apiWorkspace{}
	}
	if resp.Invitations == nil {
		resp.Invitations = []apiInvitation{}
	}

	apiOK(w, resp)
}
