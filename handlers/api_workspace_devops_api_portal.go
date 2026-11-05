package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — API Management: Developer Portal
// devops/api/portal
// ════════════════════════════════════════════════════════════════

type DevPortal struct {
	ID          string
	Name        string
	URL         string
	APIs        string
	Subscribers string
	Theme       string
	Status      string
	Updated     string
}

type devPortalStoreType struct {
	mu      sync.RWMutex
	portals map[string][]DevPortal
}

var devPortalStore = &devPortalStoreType{
	portals: map[string][]DevPortal{
		"buyvm": {
			{ID: newID(), Name: "Public Developer Portal", URL: "developers.buyvm.net", APIs: "24", Subscribers: "1,842", Theme: "default", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "Partner Integration Hub", URL: "partners.buyvm.net", APIs: "12", Subscribers: "156", Theme: "partner", Status: "active", Updated: "2026-04-10"},
			{ID: newID(), Name: "Internal API Docs", URL: "internal-api.buyvm.net", APIs: "48", Subscribers: "320", Theme: "internal", Status: "active", Updated: "2026-04-14"},
			{ID: newID(), Name: "Legacy Portal", URL: "old-dev.buyvm.net", APIs: "8", Subscribers: "45", Theme: "classic", Status: "deprecated", Updated: "2025-08-01"},
		},
		"hosting": {
			{ID: newID(), Name: "Hosting Developer Hub", URL: "developers.hosting.com", APIs: "18", Subscribers: "2,310", Theme: "default", Status: "active", Updated: "2026-04-12"},
			{ID: newID(), Name: "Enterprise API Portal", URL: "enterprise-api.hosting.com", APIs: "32", Subscribers: "89", Theme: "enterprise", Status: "active", Updated: "2026-04-05"},
			{ID: newID(), Name: "Sandbox Portal", URL: "sandbox.hosting.com", APIs: "18", Subscribers: "540", Theme: "sandbox", Status: "active", Updated: "2026-03-28"},
		},
	},
}

func (s *devPortalStoreType) list(slug string) []DevPortal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.portals[slug]
	out := make([]DevPortal, len(src))
	copy(out, src)
	return out
}

func (s *devPortalStoreType) getByID(slug, id string) (DevPortal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.portals[slug] {
		if p.ID == id {
			return p, true
		}
	}
	return DevPortal{}, false
}

func (s *devPortalStoreType) add(slug string, p DevPortal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.portals[slug] = append(s.portals[slug], p)
}

func (s *devPortalStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.portals[slug]
	for i, p := range list {
		if p.ID == id {
			s.portals[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiDevPortal struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	APIs        string `json:"apis"`
	Subscribers string `json:"subscribers"`
	Theme       string `json:"theme"`
	Status      string `json:"status"`
	Updated     string `json:"updated"`
}

type apiDevPortalList struct {
	Portals []apiDevPortal `json:"portals"`
	Total   int            `json:"total"`
}

func toAPIDevPortal(p DevPortal) apiDevPortal {
	return apiDevPortal{
		ID: p.ID, Name: p.Name, URL: p.URL,
		APIs: p.APIs, Subscribers: p.Subscribers, Theme: p.Theme,
		Status: p.Status, Updated: p.Updated,
	}
}

func filterDevPortals(all []DevPortal, search, status string) []DevPortal {
	var out []DevPortal
	q := strings.ToLower(search)
	for _, p := range all {
		if status != "" && p.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(p.Name + " " + p.URL + " " + p.Theme)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func apiWSDevOpsAPIPortalListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	portals := filterDevPortals(devPortalStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiDevPortal, len(portals))
	for i, p := range portals {
		out[i] = toAPIDevPortal(p)
	}
	apiOK(w, apiDevPortalList{Portals: out, Total: len(out)})
}

type apiCreateDevPortalRequest struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Theme string `json:"theme"`
}

func apiWSDevOpsAPIPortalCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateDevPortalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	portalURL := strings.TrimSpace(req.URL)
	if portalURL == "" {
		portalURL = "portal.example.com"
	}
	theme := strings.TrimSpace(req.Theme)
	if theme == "" {
		theme = "default"
	}
	portal := DevPortal{
		ID: newID(), Name: name, URL: portalURL,
		APIs: "0", Subscribers: "0", Theme: theme,
		Status: "active", Updated: "2026-04-17",
	}
	devPortalStore.add(slug, portal)
	apiCreated(w, toAPIDevPortal(portal))
}

func apiWSDevOpsAPIPortalGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	portal, found := devPortalStore.getByID(slug, r.PathValue("pid"))
	if !found {
		apiError(w, http.StatusNotFound, "Developer portal not found.")
		return
	}
	apiOK(w, toAPIDevPortal(portal))
}

func apiWSDevOpsAPIPortalDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := devPortalStore.getByID(slug, r.PathValue("pid")); !found {
		apiError(w, http.StatusNotFound, "Developer portal not found.")
		return
	}
	devPortalStore.delete(slug, r.PathValue("pid"))
	apiOK(w, map[string]any{"ok": true, "message": "Developer portal deleted."})
}
