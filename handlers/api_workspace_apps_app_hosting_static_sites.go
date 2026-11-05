package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — App Hosting: Static Sites
// apps/app-hosting/static-sites
// ════════════════════════════════════════════════════════════════

type HostedSite struct {
	ID           string
	Name         string
	Domain       string
	Framework    string
	Repository   string
	Branch       string
	BuildCmd     string
	OutputDir    string
	CDNRegions   int
	LastDeploy   string
	DeployStatus string
	SSL          string
	Status       string
	Updated      string
}

type hostedSiteStoreType struct {
	mu    sync.RWMutex
	items map[string][]HostedSite
}

var hostedSiteStore = &hostedSiteStoreType{
	items: map[string][]HostedSite{
		"buyvm": {
			{ID: newID(), Name: "marketing-site", Domain: "www.buyvm-demo.com", Framework: "Next.js", Repository: "github.com/buyvm/marketing", Branch: "main", BuildCmd: "npm run build", OutputDir: "out", CDNRegions: 12, LastDeploy: "2026-04-17 10:15", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "docs-portal", Domain: "docs.buyvm-demo.com", Framework: "Hugo", Repository: "github.com/buyvm/docs", Branch: "main", BuildCmd: "hugo build", OutputDir: "public", CDNRegions: 8, LastDeploy: "2026-04-16 14:30", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "status-dashboard", Domain: "status.buyvm-demo.com", Framework: "Astro", Repository: "github.com/buyvm/status", Branch: "production", BuildCmd: "astro build", OutputDir: "dist", CDNRegions: 15, LastDeploy: "2026-04-17 08:00", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "blog", Domain: "blog.buyvm-demo.com", Framework: "Gatsby", Repository: "github.com/buyvm/blog", Branch: "main", BuildCmd: "gatsby build", OutputDir: "public", CDNRegions: 6, LastDeploy: "2026-04-12 09:45", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-12"},
			{ID: newID(), Name: "landing-page-v2", Domain: "—", Framework: "Vite + React", Repository: "github.com/buyvm/landing-v2", Branch: "develop", BuildCmd: "npm run build", OutputDir: "dist", CDNRegions: 0, LastDeploy: "—", DeployStatus: "Not deployed", SSL: "Pending", Status: "draft", Updated: "2026-04-17"},
		},
		"hosting": {
			{ID: newID(), Name: "corporate-site", Domain: "www.hosting-demo.com", Framework: "Next.js", Repository: "github.com/hosting/corporate", Branch: "main", BuildCmd: "next build && next export", OutputDir: "out", CDNRegions: 20, LastDeploy: "2026-04-17 11:00", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "developer-hub", Domain: "developers.hosting-demo.com", Framework: "Docusaurus", Repository: "github.com/hosting/dev-hub", Branch: "main", BuildCmd: "npm run build", OutputDir: "build", CDNRegions: 14, LastDeploy: "2026-04-16 16:00", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "api-reference", Domain: "api.hosting-demo.com", Framework: "Redoc", Repository: "github.com/hosting/api-docs", Branch: "main", BuildCmd: "redoc-cli build", OutputDir: "dist", CDNRegions: 10, LastDeploy: "2026-04-15 20:30", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-15"},
			{ID: newID(), Name: "event-microsite", Domain: "summit.hosting-demo.com", Framework: "11ty", Repository: "github.com/hosting/summit-2026", Branch: "main", BuildCmd: "eleventy", OutputDir: "_site", CDNRegions: 5, LastDeploy: "2026-04-14 12:00", DeployStatus: "Live", SSL: "Auto-renewed", Status: "active", Updated: "2026-04-14"},
		},
	},
}

func (s *hostedSiteStoreType) list(slug string) []HostedSite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]HostedSite, len(src))
	copy(out, src)
	return out
}

func (s *hostedSiteStoreType) getByID(slug, id string) (HostedSite, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, site := range s.items[slug] {
		if site.ID == id {
			return site, true
		}
	}
	return HostedSite{}, false
}

func (s *hostedSiteStoreType) add(slug string, site HostedSite) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], site)
}

func (s *hostedSiteStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, site := range list {
		if site.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiHostedSite struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	Framework    string `json:"framework"`
	Repository   string `json:"repository"`
	Branch       string `json:"branch"`
	BuildCmd     string `json:"build_cmd"`
	OutputDir    string `json:"output_dir"`
	CDNRegions   int    `json:"cdn_regions"`
	LastDeploy   string `json:"last_deploy"`
	DeployStatus string `json:"deploy_status"`
	SSL          string `json:"ssl"`
	Status       string `json:"status"`
	Updated      string `json:"updated"`
}

type apiHostedSiteList struct {
	Sites []apiHostedSite `json:"sites"`
	Total int             `json:"total"`
}

func toAPIHostedSite(s HostedSite) apiHostedSite {
	return apiHostedSite{
		ID: s.ID, Name: s.Name, Domain: s.Domain,
		Framework: s.Framework, Repository: s.Repository,
		Branch: s.Branch, BuildCmd: s.BuildCmd,
		OutputDir: s.OutputDir, CDNRegions: s.CDNRegions,
		LastDeploy: s.LastDeploy, DeployStatus: s.DeployStatus,
		SSL: s.SSL, Status: s.Status, Updated: s.Updated,
	}
}

func filterHostedSites(all []HostedSite, search, status string) []HostedSite {
	var out []HostedSite
	q := strings.ToLower(search)
	for _, s := range all {
		if status != "" && s.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(s.Name + " " + s.Domain + " " + s.Framework + " " + s.Repository)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

func apiWSAppsAppHostingStaticSitesListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterHostedSites(hostedSiteStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiHostedSite, len(items))
	for i, s := range items {
		out[i] = toAPIHostedSite(s)
	}
	apiOK(w, apiHostedSiteList{Sites: out, Total: len(out)})
}

type apiCreateHostedSiteRequest struct {
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	Framework  string `json:"framework"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	BuildCmd   string `json:"build_cmd"`
	OutputDir  string `json:"output_dir"`
}

func apiWSAppsAppHostingStaticSitesCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateHostedSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		domain = "—"
	}
	framework := strings.TrimSpace(req.Framework)
	if framework == "" {
		framework = "Static HTML"
	}
	repo := strings.TrimSpace(req.Repository)
	if repo == "" {
		repo = "—"
	}
	branch := strings.TrimSpace(req.Branch)
	if branch == "" {
		branch = "main"
	}
	buildCmd := strings.TrimSpace(req.BuildCmd)
	if buildCmd == "" {
		buildCmd = "npm run build"
	}
	outputDir := strings.TrimSpace(req.OutputDir)
	if outputDir == "" {
		outputDir = "dist"
	}
	site := HostedSite{
		ID: newID(), Name: name, Domain: domain,
		Framework: framework, Repository: repo,
		Branch: branch, BuildCmd: buildCmd,
		OutputDir: outputDir, CDNRegions: 0,
		LastDeploy: "—", DeployStatus: "Not deployed",
		SSL: "Pending", Status: "draft", Updated: "2026-04-17",
	}
	hostedSiteStore.add(slug, site)
	apiCreated(w, toAPIHostedSite(site))
}

func apiWSAppsAppHostingStaticSitesGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	site, found := hostedSiteStore.getByID(slug, r.PathValue("sid"))
	if !found {
		apiError(w, http.StatusNotFound, "Hosted site not found.")
		return
	}
	apiOK(w, toAPIHostedSite(site))
}

func apiWSAppsAppHostingStaticSitesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := hostedSiteStore.getByID(slug, r.PathValue("sid")); !found {
		apiError(w, http.StatusNotFound, "Hosted site not found.")
		return
	}
	hostedSiteStore.delete(slug, r.PathValue("sid"))
	apiOK(w, map[string]any{"ok": true, "message": "Hosted site deleted."})
}
