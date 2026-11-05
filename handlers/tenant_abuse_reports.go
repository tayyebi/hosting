package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"platform/db"
)

// ── Abuse Report store ───────────────────────────────────────

type AbuseReport struct {
	ID      string
	Type    string // "spam" | "phishing" | "malware" | "tos-violation" | "dmca" | "other"
	Source  string
	Target  string
	Status  string // "new" | "investigating" | "resolved" | "dismissed"
	Summary string
	Created string
	Updated string
}

type abuseReportStoreType struct {
	mu      sync.RWMutex
	reports map[string][]AbuseReport // key: tenant slug
}

var abuseReportStore = &abuseReportStoreType{
	reports: map[string][]AbuseReport{
		"buyvm": {
			{ID: newID(), Type: "spam", Source: "abuse@spamhaus.org", Target: "185.22.153.10", Status: "investigating", Summary: "Outbound spam detected from IP range", Created: "2026-04-16T08:00:00Z", Updated: "2026-04-16T14:30:00Z"},
			{ID: newID(), Type: "phishing", Source: "security@google.com", Target: "acme-login.buyvm.net", Status: "new", Summary: "Phishing page mimicking Google login", Created: "2026-04-17T03:00:00Z", Updated: "2026-04-17T03:00:00Z"},
			{ID: newID(), Type: "tos-violation", Source: "legal@buyvm.net", Target: "omar@khalid.dev", Status: "resolved", Summary: "Cryptocurrency mining on shared resources", Created: "2026-04-10T12:00:00Z", Updated: "2026-04-12T16:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Type: "malware", Source: "cert@ncsc.gov.uk", Target: "93.115.20.44", Status: "investigating", Summary: "C2 beacon traffic detected", Created: "2026-04-15T18:00:00Z", Updated: "2026-04-16T09:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Type: "spam", Source: "postmaster@outlook.com", Target: "mail.hosting.com", Status: "new", Summary: "Bulk email campaign violating CAN-SPAM", Created: "2026-04-17T01:00:00Z", Updated: "2026-04-17T01:00:00Z"},
			{ID: newID(), Type: "tos-violation", Source: "compliance@hosting.com", Target: "diego@vega.ar", Status: "dismissed", Summary: "False positive - legitimate load testing", Created: "2026-04-08T10:00:00Z", Updated: "2026-04-09T14:00:00Z"},
			{ID: newID(), Type: "phishing", Source: "abuse@cloudflare.com", Target: "172.16.50.12", Status: "resolved", Summary: "Credential harvesting page on shared hosting", Created: "2026-04-11T20:00:00Z", Updated: "2026-04-13T08:00:00Z"},
		},
	},
}

func (s *abuseReportStoreType) list(slug string) []AbuseReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.reports[slug]
	out := make([]AbuseReport, len(src))
	copy(out, src)
	return out
}

func (s *abuseReportStoreType) getByID(slug, id string) (AbuseReport, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.reports[slug] {
		if r.ID == id {
			return r, true
		}
	}
	return AbuseReport{}, false
}

func (s *abuseReportStoreType) add(slug string, r AbuseReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports[slug] = append(s.reports[slug], r)
}

func (s *abuseReportStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.reports[slug] {
		if r.ID == id {
			s.reports[slug][i].Status = status
			s.reports[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *abuseReportStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.reports[slug]
	for i, r := range list {
		if r.ID == id {
			s.reports[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterAbuseReports(all []AbuseReport, search, status, abuseType string) []AbuseReport {
	var out []AbuseReport
	q := strings.ToLower(search)
	for _, r := range all {
		if status != "" && r.Status != status {
			continue
		}
		if abuseType != "" && r.Type != abuseType {
			continue
		}
		if q != "" {
			hay := strings.ToLower(r.Summary + " " + r.Source + " " + r.Target)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func dbAbuseReportToLocal(r db.AbuseReport) AbuseReport {
	return AbuseReport{
		ID:      fmt.Sprintf("%d", r.ID),
		Type:    r.Type,
		Source:  r.Source,
		Target:  r.Target,
		Status:  r.Status,
		Summary: r.Summary,
		Created: r.CreatedAt.Format(time.RFC3339),
		Updated: r.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type abuseReportsInnerData struct {
	Reports    []AbuseReport
	TenantSlug string
	Search     string
	StatusFilter string
	TypeFilter   string
}

func renderAbuseReportsPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d abuseReportsInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "abuse_reports_content.html"))
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := innerTmpl.ExecuteTemplate(&buf, "tenant-page-content", d); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shellPaths := []string{
		filepath.Join("templates", "tenant", "base.html"),
		filepath.Join("templates", "tenant", "shell.html"),
	}
	shellTmpl, err := template.ParseFiles(shellPaths...)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := TenantShellData{
		User: user, Tenant: tenant,
		Nav: tenantNav(tenant.Slug), NavGroups: tenantNavGroups(tenant.Slug),
		ActiveNav: "abuse-reports", PageTitle: "Abuse Reports",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func abuseReportsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	abuseType := r.URL.Query().Get("type")

	var reports []AbuseReport
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		drs, err := db.ListAbuseReports(ctx, Pools.Admin, tenant.DBID, q, status, abuseType)
		if err != nil {
			log.Printf("abuseReportsHandler: ListAbuseReports: %v", err)
			reports = filterAbuseReports(abuseReportStore.list(tenant.Slug), q, status, abuseType)
		} else {
			reports = make([]AbuseReport, len(drs))
			for i, dr := range drs {
				reports[i] = dbAbuseReportToLocal(dr)
			}
		}
	} else {
		reports = filterAbuseReports(abuseReportStore.list(tenant.Slug), q, status, abuseType)
	}

	renderAbuseReportsPage(w, r, tenant, user, abuseReportsInnerData{
		Reports:      reports,
		TenantSlug:   tenant.Slug,
		Search:       q,
		StatusFilter: status,
		TypeFilter:   abuseType,
	})
}

func abuseReportStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateAbuseReportStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("abuseReportStatusHandler: %v", err)
			}
		}
	} else {
		abuseReportStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/abuse-reports/", http.StatusSeeOther)
}

func abuseReportDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteAbuseReport(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("abuseReportDeleteHandler: %v", err)
			}
		}
	} else {
		abuseReportStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/abuse-reports/", http.StatusSeeOther)
}

func RegisterTenantAbuseReportsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/abuse-reports/", tenantMiddleware(abuseReportsHandler))
	mux.HandleFunc("POST /t/{slug}/abuse-reports/{id}/status/", tenantMiddleware(abuseReportStatusHandler))
	mux.HandleFunc("POST /t/{slug}/abuse-reports/{id}/delete/", tenantMiddleware(abuseReportDeleteHandler))
}
