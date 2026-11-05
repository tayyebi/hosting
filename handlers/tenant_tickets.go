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

// ── Ticket store ─────────────────────────────────────────────

type Ticket struct {
	ID       string
	Subject  string
	Client   string
	Status   string // "open" | "in-progress" | "resolved" | "closed"
	Priority string // "low" | "medium" | "high" | "critical"
	Created  string
	Updated  string
}

type ticketStoreType struct {
	mu      sync.RWMutex
	tickets map[string][]Ticket // key: tenant slug
}

var ticketStore = &ticketStoreType{
	tickets: map[string][]Ticket{
		"buyvm": {
			{ID: newID(), Subject: "Server unreachable after reboot", Client: "james@acme.io", Status: "open", Priority: "high", Created: "2026-04-15T10:30:00Z", Updated: "2026-04-15T14:20:00Z"},
			{ID: newID(), Subject: "Billing discrepancy on invoice #4021", Client: "sara@techblue.com", Status: "in-progress", Priority: "medium", Created: "2026-04-14T08:15:00Z", Updated: "2026-04-15T09:00:00Z"},
			{ID: newID(), Subject: "Request for additional IP allocation", Client: "lena@bauerhosts.de", Status: "resolved", Priority: "low", Created: "2026-04-12T16:45:00Z", Updated: "2026-04-13T11:30:00Z"},
			{ID: newID(), Subject: "DDoS attack on VPS node", Client: "omar@khalid.dev", Status: "closed", Priority: "critical", Created: "2026-04-10T02:00:00Z", Updated: "2026-04-11T18:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Subject: "DNS propagation delay", Client: "priya@infranow.in", Status: "open", Priority: "medium", Created: "2026-04-16T09:00:00Z", Updated: "2026-04-16T09:00:00Z"},
			{ID: newID(), Subject: "SSL certificate renewal failed", Client: "tom@fischer.eu", Status: "in-progress", Priority: "high", Created: "2026-04-15T14:00:00Z", Updated: "2026-04-16T10:30:00Z"},
		},
		"hosting": {
			{ID: newID(), Subject: "Kubernetes pod crash loop", Client: "alice@monroe.io", Status: "open", Priority: "critical", Created: "2026-04-17T06:00:00Z", Updated: "2026-04-17T08:45:00Z"},
			{ID: newID(), Subject: "Cannot access control panel", Client: "ben@okonkwo.ng", Status: "open", Priority: "high", Created: "2026-04-16T22:00:00Z", Updated: "2026-04-17T01:15:00Z"},
			{ID: newID(), Subject: "Slow disk I/O on SSD volume", Client: "chloe@martindev.fr", Status: "in-progress", Priority: "medium", Created: "2026-04-15T11:00:00Z", Updated: "2026-04-16T15:00:00Z"},
			{ID: newID(), Subject: "Feature request: automated backups", Client: "elena@popov.ru", Status: "closed", Priority: "low", Created: "2026-04-10T07:00:00Z", Updated: "2026-04-12T09:30:00Z"},
			{ID: newID(), Subject: "Network latency spike EU region", Client: "faisal@alamin.sa", Status: "resolved", Priority: "high", Created: "2026-04-13T19:00:00Z", Updated: "2026-04-14T12:00:00Z"},
		},
	},
}

func (s *ticketStoreType) list(slug string) []Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.tickets[slug]
	out := make([]Ticket, len(src))
	copy(out, src)
	return out
}

func (s *ticketStoreType) getByID(slug, id string) (Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tickets[slug] {
		if t.ID == id {
			return t, true
		}
	}
	return Ticket{}, false
}

func (s *ticketStoreType) add(slug string, t Ticket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[slug] = append(s.tickets[slug], t)
}

func (s *ticketStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.tickets[slug]
	for i, t := range list {
		if t.ID == id {
			s.tickets[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterTickets(all []Ticket, search, status, priority string) []Ticket {
	var out []Ticket
	q := strings.ToLower(search)
	for _, t := range all {
		if status != "" && t.Status != status {
			continue
		}
		if priority != "" && t.Priority != priority {
			continue
		}
		if q != "" {
			hay := strings.ToLower(t.Subject + " " + t.Client)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, t)
	}
	return out
}

// ── DB conversion ─────────────────────────────────────────────

func dbTicketToLocal(t db.Ticket) Ticket {
	return Ticket{
		ID:       fmt.Sprintf("%d", t.ID),
		Subject:  t.Subject,
		Client:   t.ClientEmail,
		Status:   t.Status,
		Priority: t.Priority,
		Created:  t.CreatedAt.Format(time.RFC3339),
		Updated:  t.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type ticketsInnerData struct {
	Tickets        []Ticket
	TenantSlug     string
	Search         string
	StatusFilter   string
	PriorityFilter string
}

func renderTicketsPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d ticketsInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "tickets_content.html"))
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
		ActiveNav: "tickets", PageTitle: "Tickets",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func ticketsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")

	var tickets []Ticket
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dts, err := db.ListTickets(ctx, Pools.Admin, tenant.DBID, q, status, priority)
		if err != nil {
			log.Printf("ticketsHandler: ListTickets: %v", err)
			tickets = filterTickets(ticketStore.list(tenant.Slug), q, status, priority)
		} else {
			tickets = make([]Ticket, len(dts))
			for i, dt := range dts {
				tickets[i] = dbTicketToLocal(dt)
			}
		}
	} else {
		tickets = filterTickets(ticketStore.list(tenant.Slug), q, status, priority)
	}

	renderTicketsPage(w, r, tenant, user, ticketsInnerData{
		Tickets:        tickets,
		TenantSlug:     tenant.Slug,
		Search:         q,
		StatusFilter:   status,
		PriorityFilter: priority,
	})
}

func ticketsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		ticketID, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteTicket(ctx, Pools.Admin, tenant.DBID, ticketID); err != nil {
				log.Printf("ticketsDeleteHandler: DeleteTicket: %v", err)
			}
		}
	} else {
		ticketStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/tickets/", http.StatusSeeOther)
}

func RegisterTenantTicketsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/tickets/", tenantMiddleware(ticketsHandler))
	mux.HandleFunc("POST /t/{slug}/tickets/{id}/delete/", tenantMiddleware(ticketsDeleteHandler))
}
