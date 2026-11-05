package handlers

import (
	"net/http"
)

// ── Workspace overview API types ──────────────────────────────

type apiWSNavLeaf struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

type apiWSNavSection struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Items []apiWSNavLeaf `json:"items"`
}

type apiWSNavCategory struct {
	ID       string            `json:"id"`
	Label    string            `json:"label"`
	Icon     string            `json:"icon"`
	Sections []apiWSNavSection `json:"sections"`
}

type apiWorkspaceOverview struct {
	Workspace    apiWSWorkspace     `json:"workspace"`
	Client       apiWSClient        `json:"client"`
	OtherSpaces  []apiWSWorkspace   `json:"other_workspaces"`
	Nav          []apiWSNavCategory `json:"nav"`
}

type apiWSWorkspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiWSClient struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Status  string `json:"status"`
}

// ── GET /api/tenant/{slug}/workspace/{id} ─────────────────────

func apiTenantWorkspaceOverviewHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}

	slug := r.PathValue("slug")
	if _, found := store.getBySlug(slug); !found {
		apiError(w, http.StatusNotFound, "Tenant not found.")
		return
	}
	if user.Role != RoleSysAdmin && (user.Role != RoleTenant || user.TenantSlug != slug) {
		apiError(w, http.StatusForbidden, "Access denied.")
		return
	}

	wsID := r.PathValue("id")
	client, ws, ok := clientStore.getWorkspace(slug, wsID)
	if !ok {
		apiError(w, http.StatusNotFound, "Workspace not found.")
		return
	}

	base := "/t/" + slug + "/w/" + wsID
	navCats := wsNavCategories(base)

	apiNav := make([]apiWSNavCategory, len(navCats))
	for i, cat := range navCats {
		secs := make([]apiWSNavSection, len(cat.Sections))
		for j, sec := range cat.Sections {
			items := make([]apiWSNavLeaf, len(sec.Items))
			for k, leaf := range sec.Items {
				items[k] = apiWSNavLeaf{ID: leaf.ID, Label: leaf.Label, Path: leaf.Path}
			}
			secs[j] = apiWSNavSection{ID: sec.ID, Label: sec.Label, Items: items}
		}
		apiNav[i] = apiWSNavCategory{ID: cat.ID, Label: cat.Label, Icon: cat.Icon, Sections: secs}
	}

	otherSpaces := make([]apiWSWorkspace, 0, len(client.Workspaces))
	for _, w := range client.Workspaces {
		if w.ID != wsID {
			otherSpaces = append(otherSpaces, apiWSWorkspace{ID: w.ID, Name: w.Name})
		}
	}

	apiOK(w, apiWorkspaceOverview{
		Workspace: apiWSWorkspace{ID: ws.ID, Name: ws.Name},
		Client: apiWSClient{
			ID:      client.ID,
			Name:    client.Name,
			Email:   client.Email,
			Company: client.Company,
			Status:  client.Status,
		},
		OtherSpaces: otherSpaces,
		Nav:         apiNav,
	})
}
