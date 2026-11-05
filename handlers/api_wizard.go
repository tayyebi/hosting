package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"platform/db"
)

// ── GET /api/wizard ───────────────────────────────────────

type apiWizardApp struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Tagline     string   `json:"tagline"`
	Category    string   `json:"category"`
	IconColor   string   `json:"icon_color"`
	IconFG      string   `json:"icon_fg"`
	IconLetter  string   `json:"icon_letter"`
	Version     string   `json:"version"`
	License     string   `json:"license"`
	Publisher   string   `json:"publisher"`
	Rating      float64  `json:"rating"`
	Installs    string   `json:"installs"`
	Tags        []string `json:"tags"`
	Featured    bool     `json:"featured"`
	Verified    bool     `json:"verified"`
	MinRAM      string   `json:"min_ram"`
	MinCPU      string   `json:"min_cpu"`
	MinDisk     string   `json:"min_disk"`
	Description string   `json:"description"`
	Hero        string   `json:"hero"`
	DeployCount int      `json:"deploy_count,omitempty"`
}

type apiWizardCategory struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
	Count int    `json:"count"`
}

type apiWizardResponse struct {
	Categories []apiWizardCategory `json:"categories"`
	Apps       []apiWizardApp      `json:"apps"`
	Featured   []apiWizardApp      `json:"featured"`
	TotalApps  int                 `json:"total_apps"`
}

func toAPIWizardApps(apps []WizardApp) []apiWizardApp {
	out := make([]apiWizardApp, len(apps))
	for i, a := range apps {
		out[i] = apiWizardApp{
			Slug: a.Slug, Name: a.Name, Tagline: a.Tagline, Category: a.Category,
			IconColor: a.IconColor, IconFG: a.IconFG, IconLetter: a.IconLetter,
			Version: a.Version, License: a.License, Publisher: a.Publisher,
			Rating: a.Rating, Installs: a.Installs, Tags: a.Tags,
			Featured: a.Featured, Verified: a.Verified,
			MinRAM: a.MinRAM, MinCPU: a.MinCPU, MinDisk: a.MinDisk,
			Description: a.Description, Hero: a.Hero,
		}
	}
	return out
}

func apiWizardHandler(w http.ResponseWriter, r *http.Request) {
	apps := wizardApps()
	cats := wizardCategories(apps)
	feat := wizardFeatured(apps)

	apiCats := make([]apiWizardCategory, len(cats))
	for i, c := range cats {
		apiCats[i] = apiWizardCategory{ID: c.ID, Label: c.Label, Icon: c.Icon, Count: c.Count}
	}

	apiOK(w, apiWizardResponse{
		Categories: apiCats,
		Apps:       toAPIWizardApps(apps),
		Featured:   toAPIWizardApps(feat),
		TotalApps:  len(apps),
	})
}

// ── GET /api/wizard/{slug} ────────────────────────────────

func apiWizardAppHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	for _, a := range wizardApps() {
		if a.Slug == slug {
			result := toAPIWizardApps([]WizardApp{a})[0]
			if Pools != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
				defer cancel()
				if n, err := db.CountAppDeploys(ctx, Pools.App, slug); err == nil {
					result.DeployCount = n
				}
			}
			apiOK(w, result)
			return
		}
	}
	apiError(w, http.StatusNotFound, "App not found.")
}

// ── POST /api/wizard/{slug}/deploy ────────────────────────

type apiWizardDeployRequest struct {
	Region      string `json:"region"`
	TenantSlug  string `json:"tenant_slug"`
	WorkspaceID int64  `json:"workspace_id"`
}

type apiWizardDeployResponse struct {
	DeployID int64  `json:"deploy_id"`
	AppSlug  string `json:"app_slug"`
	AppName  string `json:"app_name"`
	Status   string `json:"status"`
	Region   string `json:"region"`
	Message  string `json:"message"`
}

func apiWizardDeployHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	var app *WizardApp
	for _, a := range wizardApps() {
		if a.Slug == slug {
			tmp := a
			app = &tmp
			break
		}
	}
	if app == nil {
		apiError(w, http.StatusNotFound, "App not found.")
		return
	}

	var req apiWizardDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = apiWizardDeployRequest{Region: "us-east-1"}
	}
	if req.Region == "" {
		req.Region = "us-east-1"
	}

	deployRecord := db.WizardDeploy{
		AppSlug: app.Slug,
		AppName: app.Name,
		Region:  req.Region,
		Version: app.Version,
	}

	// Try to resolve user and tenant from session.
	if user, ok := getSession(r); ok && user.ID > 0 {
		deployRecord.UserID = &user.ID
	}

	if Pools != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		created, err := db.CreateWizardDeploy(ctx, Pools.App, deployRecord)
		if err == nil {
			apiCreated(w, apiWizardDeployResponse{
				DeployID: created.ID,
				AppSlug:  app.Slug,
				AppName:  app.Name,
				Status:   created.Status,
				Region:   created.Region,
				Message:  "Deploy queued. " + app.Name + " will be ready in a few minutes.",
			})
			return
		}
	}

	// Pools unavailable — return a synthetic response.
	apiCreated(w, apiWizardDeployResponse{
		DeployID: 0,
		AppSlug:  app.Slug,
		AppName:  app.Name,
		Status:   "queued",
		Region:   req.Region,
		Message:  "Deploy queued. " + app.Name + " will be ready in a few minutes.",
	})
}
