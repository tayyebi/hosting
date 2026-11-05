package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — CI/CD: Source Control
// devops/cicd/source
// ════════════════════════════════════════════════════════════════

type SourceRepo struct {
	ID       string
	Name     string
	Provider string
	URL      string
	Branch   string
	Status   string
	Commits  string
	Updated  string
}

type sourceRepoStoreType struct {
	mu    sync.RWMutex
	repos map[string][]SourceRepo
}

var sourceRepoStore = &sourceRepoStoreType{
	repos: map[string][]SourceRepo{},
}

func (s *sourceRepoStoreType) list(slug string) []SourceRepo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.repos[slug]
	out := make([]SourceRepo, len(src))
	copy(out, src)
	return out
}

func (s *sourceRepoStoreType) getByID(slug, id string) (SourceRepo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.repos[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return SourceRepo{}, false
}

func (s *sourceRepoStoreType) add(slug string, r SourceRepo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repos[slug] = append(s.repos[slug], r)
}

func (s *sourceRepoStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.repos[slug]
	for i, r := range list {
		if r.ID == id {
			s.repos[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiSourceRepo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	Status   string `json:"status"`
	Commits  string `json:"commits"`
	Updated  string `json:"updated"`
}

type apiSourceRepoList struct {
	Repos []apiSourceRepo `json:"repos"`
	Total int             `json:"total"`
}

func toAPISourceRepo(r SourceRepo) apiSourceRepo {
	return apiSourceRepo{
		ID: r.ID, Name: r.Name, Provider: r.Provider,
		URL: r.URL, Branch: r.Branch, Status: r.Status,
		Commits: r.Commits, Updated: r.Updated,
	}
}

func filterSourceRepos(all []SourceRepo, search, status string) []SourceRepo {
	var out []SourceRepo
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Name + " " + r.Provider + " " + r.URL)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func apiWSDevOpsCICDSourceListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	repos := filterSourceRepos(sourceRepoStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiSourceRepo, len(repos))
	for i, repo := range repos {
		out[i] = toAPISourceRepo(repo)
	}
	apiOK(w, apiSourceRepoList{Repos: out, Total: len(out)})
}

type apiCreateSourceRepoRequest struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	URL      string `json:"url"`
	Branch   string `json:"branch"`
}

func apiWSDevOpsCICDSourceCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateSourceRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "GitHub"
	}
	repoURL := strings.TrimSpace(req.URL)
	if repoURL == "" {
		repoURL = "github.com/org/" + name
	}
	branch := strings.TrimSpace(req.Branch)
	if branch == "" {
		branch = "main"
	}
	repo := SourceRepo{
		ID: newID(), Name: name, Provider: provider,
		URL: repoURL, Branch: branch, Status: "active",
		Commits: "0", Updated: "2026-04-17",
	}
	sourceRepoStore.add(slug, repo)
	apiCreated(w, toAPISourceRepo(repo))
}

func apiWSDevOpsCICDSourceGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	repo, found := sourceRepoStore.getByID(slug, r.PathValue("rid"))
	if !found {
		apiError(w, http.StatusNotFound, "Source repository not found.")
		return
	}
	apiOK(w, toAPISourceRepo(repo))
}

func apiWSDevOpsCICDSourceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := sourceRepoStore.getByID(slug, r.PathValue("rid")); !found {
		apiError(w, http.StatusNotFound, "Source repository not found.")
		return
	}
	sourceRepoStore.delete(slug, r.PathValue("rid"))
	apiOK(w, map[string]any{"ok": true, "message": "Source repository deleted."})
}
