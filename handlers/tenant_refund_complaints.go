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

// ── Refund Complaint store ───────────────────────────────────

type RefundComplaint struct {
	ID        string
	Client    string
	Amount    string
	Currency  string
	Reason    string
	Status    string // "pending" | "approved" | "denied" | "escalated"
	InvoiceID string
	Created   string
	Updated   string
}

type refundComplaintStoreType struct {
	mu         sync.RWMutex
	complaints map[string][]RefundComplaint // key: tenant slug
}

var refundComplaintStore = &refundComplaintStoreType{
	complaints: map[string][]RefundComplaint{
		"buyvm": {
			{ID: newID(), Client: "james@acme.io", Amount: "49.99", Currency: "USD", Reason: "Service downtime exceeded SLA", Status: "approved", InvoiceID: "INV-4021", Created: "2026-04-14T10:00:00Z", Updated: "2026-04-15T16:00:00Z"},
			{ID: newID(), Client: "sara@techblue.com", Amount: "120.00", Currency: "USD", Reason: "Duplicate charge on monthly billing", Status: "pending", InvoiceID: "INV-4035", Created: "2026-04-16T09:30:00Z", Updated: "2026-04-16T09:30:00Z"},
			{ID: newID(), Client: "omar@khalid.dev", Amount: "25.00", Currency: "USD", Reason: "Charged after account suspension", Status: "escalated", InvoiceID: "INV-3998", Created: "2026-04-12T14:00:00Z", Updated: "2026-04-15T11:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Client: "priya@infranow.in", Amount: "85.50", Currency: "EUR", Reason: "Prorated refund for cancelled service", Status: "pending", InvoiceID: "INV-2201", Created: "2026-04-16T12:00:00Z", Updated: "2026-04-16T12:00:00Z"},
			{ID: newID(), Client: "tom@fischer.eu", Amount: "200.00", Currency: "EUR", Reason: "Data loss incident compensation", Status: "escalated", InvoiceID: "INV-2180", Created: "2026-04-10T08:00:00Z", Updated: "2026-04-14T17:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Client: "alice@monroe.io", Amount: "75.00", Currency: "USD", Reason: "Migration downtime not communicated", Status: "approved", InvoiceID: "INV-8810", Created: "2026-04-11T15:00:00Z", Updated: "2026-04-13T10:00:00Z"},
			{ID: newID(), Client: "diego@vega.ar", Amount: "310.00", Currency: "USD", Reason: "Overcharged for bandwidth overage", Status: "denied", InvoiceID: "INV-8795", Created: "2026-04-08T09:00:00Z", Updated: "2026-04-10T14:00:00Z"},
			{ID: newID(), Client: "elena@popov.ru", Amount: "15.99", Currency: "USD", Reason: "Add-on feature not delivered", Status: "pending", InvoiceID: "INV-8842", Created: "2026-04-17T07:00:00Z", Updated: "2026-04-17T07:00:00Z"},
		},
	},
}

func (s *refundComplaintStoreType) list(slug string) []RefundComplaint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.complaints[slug]
	out := make([]RefundComplaint, len(src))
	copy(out, src)
	return out
}

func (s *refundComplaintStoreType) getByID(slug, id string) (RefundComplaint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.complaints[slug] {
		if c.ID == id {
			return c, true
		}
	}
	return RefundComplaint{}, false
}

func (s *refundComplaintStoreType) add(slug string, c RefundComplaint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.complaints[slug] = append(s.complaints[slug], c)
}

func (s *refundComplaintStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.complaints[slug] {
		if c.ID == id {
			s.complaints[slug][i].Status = status
			s.complaints[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *refundComplaintStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.complaints[slug]
	for i, c := range list {
		if c.ID == id {
			s.complaints[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterRefundComplaints(all []RefundComplaint, search, status string) []RefundComplaint {
	var out []RefundComplaint
	q := strings.ToLower(search)
	for _, c := range all {
		if status != "" && c.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.Client + " " + c.Reason + " " + c.InvoiceID)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func dbRefundComplaintToLocal(r db.RefundComplaint) RefundComplaint {
	return RefundComplaint{
		ID:        fmt.Sprintf("%d", r.ID),
		Client:    r.Client,
		Amount:    r.Amount,
		Currency:  r.Currency,
		Reason:    r.Reason,
		Status:    r.Status,
		InvoiceID: r.InvoiceID,
		Created:   r.CreatedAt.Format(time.RFC3339),
		Updated:   r.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type refundComplaintsInnerData struct {
	Complaints   []RefundComplaint
	TenantSlug   string
	Search       string
	StatusFilter string
}

func renderRefundComplaintsPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d refundComplaintsInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "refund_complaints_content.html"))
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
		ActiveNav: "refund-complaints", PageTitle: "Refund Complaints",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func refundComplaintsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	var complaints []RefundComplaint
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dcs, err := db.ListRefundComplaints(ctx, Pools.Admin, tenant.DBID, q, status)
		if err != nil {
			log.Printf("refundComplaintsHandler: ListRefundComplaints: %v", err)
			complaints = filterRefundComplaints(refundComplaintStore.list(tenant.Slug), q, status)
		} else {
			complaints = make([]RefundComplaint, len(dcs))
			for i, dc := range dcs {
				complaints[i] = dbRefundComplaintToLocal(dc)
			}
		}
	} else {
		complaints = filterRefundComplaints(refundComplaintStore.list(tenant.Slug), q, status)
	}

	renderRefundComplaintsPage(w, r, tenant, user, refundComplaintsInnerData{
		Complaints:   complaints,
		TenantSlug:   tenant.Slug,
		Search:       q,
		StatusFilter: status,
	})
}

func refundComplaintStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateRefundComplaintStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("refundComplaintStatusHandler: %v", err)
			}
		}
	} else {
		refundComplaintStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/refund-complaints/", http.StatusSeeOther)
}

func refundComplaintDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteRefundComplaint(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("refundComplaintDeleteHandler: %v", err)
			}
		}
	} else {
		refundComplaintStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/refund-complaints/", http.StatusSeeOther)
}

func RegisterTenantRefundComplaintsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/refund-complaints/", tenantMiddleware(refundComplaintsHandler))
	mux.HandleFunc("POST /t/{slug}/refund-complaints/{id}/status/", tenantMiddleware(refundComplaintStatusHandler))
	mux.HandleFunc("POST /t/{slug}/refund-complaints/{id}/delete/", tenantMiddleware(refundComplaintDeleteHandler))
}
