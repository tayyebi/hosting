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
	items: map[string][]HostedSite{},
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
