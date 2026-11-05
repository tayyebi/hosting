package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — IaC: Templating
// devops/iac/templates
// ════════════════════════════════════════════════════════════════

type IaCTemplate struct {
	ID       string
	Name     string
	Engine   string
	Provider string
	Type     string
	Status   string
	Version  string
	Updated  string
}

type iacTemplateStoreType struct {
	mu        sync.RWMutex
	templates map[string][]IaCTemplate
}

var iacTemplateStore = &iacTemplateStoreType{
	templates: map[string][]IaCTemplate{},
}

func (s *iacTemplateStoreType) list(slug string) []IaCTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.templates[slug]
	out := make([]IaCTemplate, len(src))
	copy(out, src)
	return out
}

func (s *iacTemplateStoreType) getByID(slug, id string) (IaCTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.templates[slug] {
		if t.ID == id {
			return t, true
		}
	}
	return IaCTemplate{}, false
}

func (s *iacTemplateStoreType) add(slug string, t IaCTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[slug] = append(s.templates[slug], t)
}

func (s *iacTemplateStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.templates[slug]
	for i, t := range list {
		if t.ID == id {
			s.templates[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiIaCTemplate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Engine   string `json:"engine"`
	Provider string `json:"provider"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Version  string `json:"version"`
	Updated  string `json:"updated"`
}

type apiIaCTemplateList struct {
	Templates []apiIaCTemplate `json:"templates"`
	Total     int              `json:"total"`
}

func toAPIIaCTemplate(t IaCTemplate) apiIaCTemplate {
	return apiIaCTemplate{
		ID: t.ID, Name: t.Name, Engine: t.Engine,
		Provider: t.Provider, Type: t.Type, Status: t.Status,
		Version: t.Version, Updated: t.Updated,
	}
}

func filterIaCTemplates(all []IaCTemplate, search, status string) []IaCTemplate {
	var out []IaCTemplate
	q := strings.ToLower(search)
	for _, t := range all {
		if status != "" && t.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(t.Name + " " + t.Engine + " " + t.Provider + " " + t.Type)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, t)
	}
	return out
}

func apiWSDevOpsIaCTemplatesListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	templates := filterIaCTemplates(iacTemplateStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiIaCTemplate, len(templates))
	for i, t := range templates {
		out[i] = toAPIIaCTemplate(t)
	}
	apiOK(w, apiIaCTemplateList{Templates: out, Total: len(out)})
}

type apiCreateIaCTemplateRequest struct {
	Name     string `json:"name"`
	Engine   string `json:"engine"`
	Provider string `json:"provider"`
	Type     string `json:"type"`
}

func apiWSDevOpsIaCTemplatesCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateIaCTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	engine := strings.TrimSpace(req.Engine)
	if engine == "" {
		engine = "Terraform"
	}
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "AWS"
	}
	tplType := strings.TrimSpace(req.Type)
	if tplType == "" {
		tplType = "general"
	}
	tpl := IaCTemplate{
		ID: newID(), Name: name, Engine: engine,
		Provider: provider, Type: tplType, Status: "active",
		Version: "1.0.0", Updated: "2026-04-17",
	}
	iacTemplateStore.add(slug, tpl)
	apiCreated(w, toAPIIaCTemplate(tpl))
}

func apiWSDevOpsIaCTemplatesGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	tpl, found := iacTemplateStore.getByID(slug, r.PathValue("tid"))
	if !found {
		apiError(w, http.StatusNotFound, "IaC template not found.")
		return
	}
	apiOK(w, toAPIIaCTemplate(tpl))
}

func apiWSDevOpsIaCTemplatesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := iacTemplateStore.getByID(slug, r.PathValue("tid")); !found {
		apiError(w, http.StatusNotFound, "IaC template not found.")
		return
	}
	iacTemplateStore.delete(slug, r.PathValue("tid"))
	apiOK(w, map[string]any{"ok": true, "message": "IaC template deleted."})
}
