package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

// ── GET /api/tenant/{slug}/one-time-access ──────────────────

type apiOneTimeAccess struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Token     string `json:"token"`
	Target    string `json:"target"`
	GrantedTo string `json:"granted_to"`
	GrantedBy string `json:"granted_by"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	Created   string `json:"created"`
}

type apiOneTimeAccessListResponse struct {
	Tokens []apiOneTimeAccess `json:"tokens"`
	Total  int                `json:"total"`
}

func otaToAPI(t OneTimeAccess) apiOneTimeAccess {
	return apiOneTimeAccess{
		ID: t.ID, Label: t.Label, Token: t.Token,
		Target: t.Target, GrantedTo: t.GrantedTo, GrantedBy: t.GrantedBy,
		Status: t.Status, ExpiresAt: t.ExpiresAt, Created: t.Created,
	}
}

func apiTenantOneTimeAccessListHandler(w http.ResponseWriter, r *http.Request) {
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

	search := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	tokens := filterOneTimeAccess(oneTimeAccessStore.list(slug), search, status)

	out := make([]apiOneTimeAccess, len(tokens))
	for i, t := range tokens {
		out[i] = otaToAPI(t)
	}

	apiOK(w, apiOneTimeAccessListResponse{Tokens: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/one-time-access ─────────────────

type apiCreateOneTimeAccessRequest struct {
	Label     string `json:"label"`
	Target    string `json:"target"`
	GrantedTo string `json:"granted_to"`
	ExpiresAt string `json:"expires_at"`
}

func apiTenantOneTimeAccessCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateOneTimeAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	label := strings.TrimSpace(req.Label)
	target := strings.TrimSpace(req.Target)
	grantedTo := strings.TrimSpace(req.GrantedTo)
	expiresAt := strings.TrimSpace(req.ExpiresAt)

	if label == "" || target == "" || grantedTo == "" {
		apiError(w, http.StatusBadRequest, "Label, target, and granted_to are required.")
		return
	}
	if expiresAt == "" {
		expiresAt = "2026-04-18T23:59:00Z"
	}

	// Generate a unique token
	b := make([]byte, 4)
	rand.Read(b)
	token := "ota-" + hex.EncodeToString(b)

	t := OneTimeAccess{
		ID: newID(), Label: label, Token: token, Target: target,
		GrantedTo: grantedTo, GrantedBy: user.Email, Status: "active",
		ExpiresAt: expiresAt, Created: "2026-04-17T00:00:00Z",
	}
	oneTimeAccessStore.add(slug, t)

	apiCreated(w, otaToAPI(t))
}

// ── GET /api/tenant/{slug}/one-time-access/{id} ─────────────

func apiTenantOneTimeAccessGetHandler(w http.ResponseWriter, r *http.Request) {
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

	id := r.PathValue("id")
	t, ok := oneTimeAccessStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "One-time access token not found.")
		return
	}

	apiOK(w, otaToAPI(t))
}

// ── DELETE /api/tenant/{slug}/one-time-access/{id} ──────────

func apiTenantOneTimeAccessDeleteHandler(w http.ResponseWriter, r *http.Request) {
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

	id := r.PathValue("id")
	if _, ok := oneTimeAccessStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "One-time access token not found.")
		return
	}

	oneTimeAccessStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "One-time access token deleted.",
	})
}
