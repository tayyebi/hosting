package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Security API — IAM: Multi-factor Authentication
// security/iam/mfa
// ════════════════════════════════════════════════════════════════

type MFAEnrollment struct {
	ID       string
	User     string
	Method   string
	Roles    string
	Status   string
	LastUsed string
	Created  string
	Region   string
}

type mfaEnrollmentStoreType struct {
	mu          sync.RWMutex
	enrollments map[string][]MFAEnrollment
}

var mfaEnrollmentStore = &mfaEnrollmentStoreType{
	enrollments: map[string][]MFAEnrollment{
		"buyvm": {
			{ID: newID(), User: "jane.doe@acme.io", Method: "Passkey + TOTP", Roles: "Admin, Finance", Status: "enrolled", LastUsed: "8 min ago", Created: "2025-08-12", Region: "eu-central-1"},
			{ID: newID(), User: "mike.chen@acme.io", Method: "YubiKey 5", Roles: "DevOps", Status: "enrolled", LastUsed: "28 min ago", Created: "2025-09-01", Region: "eu-central-1"},
			{ID: newID(), User: "temp.contractor@acme.io", Method: "SMS only", Roles: "Viewer", Status: "weak", LastUsed: "3h ago", Created: "2026-01-15", Region: "eu-central-1"},
			{ID: newID(), User: "lena@acme.io", Method: "Passkey + TOTP", Roles: "Admin", Status: "enrolled", LastUsed: "1h ago", Created: "2025-11-20", Region: "eu-central-1"},
		},
		"hosting": {
			{ID: newID(), User: "ops@hosting.io", Method: "Passkey", Roles: "Admin", Status: "enrolled", LastUsed: "2 min ago", Created: "2025-10-01", Region: "ap-south-1"},
			{ID: newID(), User: "dev@hosting.io", Method: "TOTP", Roles: "Developer", Status: "enrolled", LastUsed: "45 min ago", Created: "2025-10-15", Region: "ap-south-1"},
			{ID: newID(), User: "intern@hosting.io", Method: "SMS only", Roles: "Viewer", Status: "weak", LastUsed: "2d ago", Created: "2026-02-01", Region: "ap-south-1"},
		},
	},
}

func (s *mfaEnrollmentStoreType) list(slug string) []MFAEnrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.enrollments[slug]
	out := make([]MFAEnrollment, len(src))
	copy(out, src)
	return out
}

func (s *mfaEnrollmentStoreType) getByID(slug, id string) (MFAEnrollment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.enrollments[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return MFAEnrollment{}, false
}

func (s *mfaEnrollmentStoreType) add(slug string, e MFAEnrollment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enrollments[slug] = append(s.enrollments[slug], e)
}

func (s *mfaEnrollmentStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.enrollments[slug]
	for i, e := range list {
		if e.ID == id {
			s.enrollments[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiMFAEnrollment struct {
	ID       string `json:"id"`
	User     string `json:"user"`
	Method   string `json:"method"`
	Roles    string `json:"roles"`
	Status   string `json:"status"`
	LastUsed string `json:"last_used"`
	Created  string `json:"created"`
	Region   string `json:"region"`
}

type apiMFAEnrollmentList struct {
	Enrollments []apiMFAEnrollment `json:"enrollments"`
	Total       int                `json:"total"`
}

func toAPIMFAEnrollment(e MFAEnrollment) apiMFAEnrollment {
	return apiMFAEnrollment{
		ID: e.ID, User: e.User, Method: e.Method,
		Roles: e.Roles, Status: e.Status,
		LastUsed: e.LastUsed, Created: e.Created,
		Region: e.Region,
	}
}

func filterMFAEnrollments(all []MFAEnrollment, search, status string) []MFAEnrollment {
	var out []MFAEnrollment
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.User + " " + e.Method + " " + e.Roles + " " + e.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSSecIAMMFAListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	enrollments := filterMFAEnrollments(mfaEnrollmentStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMFAEnrollment, len(enrollments))
	for i, e := range enrollments {
		out[i] = toAPIMFAEnrollment(e)
	}
	apiOK(w, apiMFAEnrollmentList{Enrollments: out, Total: len(out)})
}

type apiCreateMFAEnrollmentRequest struct {
	User   string `json:"user"`
	Method string `json:"method"`
	Roles  string `json:"roles"`
	Region string `json:"region"`
}

func apiWSSecIAMMFACreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMFAEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	user := strings.TrimSpace(req.User)
	if user == "" {
		apiError(w, http.StatusBadRequest, "User is required.")
		return
	}
	method := strings.TrimSpace(req.Method)
	if method == "" {
		method = "TOTP"
	}
	roles := strings.TrimSpace(req.Roles)
	if roles == "" {
		apiError(w, http.StatusBadRequest, "Roles is required.")
		return
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "eu-central-1"
	}
	enrollment := MFAEnrollment{
		ID: newID(), User: user, Method: method,
		Roles: roles, Status: "pending",
		LastUsed: "—", Created: "2026-04-17",
		Region: region,
	}
	mfaEnrollmentStore.add(slug, enrollment)
	apiCreated(w, toAPIMFAEnrollment(enrollment))
}

func apiWSSecIAMMFAGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	enrollment, found := mfaEnrollmentStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "MFA enrollment not found.")
		return
	}
	apiOK(w, toAPIMFAEnrollment(enrollment))
}

func apiWSSecIAMMFADeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := mfaEnrollmentStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "MFA enrollment not found.")
		return
	}
	mfaEnrollmentStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "MFA enrollment deleted."})
}
