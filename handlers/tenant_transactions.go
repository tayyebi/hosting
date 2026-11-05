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

// ── Transaction store ─────────────────────────────────────────

type TenantTxn struct {
	ID          string
	Reference   string
	Client      string
	ClientCo    string
	Description string
	TxnType     string // "charge" | "credit" | "refund" | "metered" | "adjustment"
	Amount      string // formatted: "+$480.00" or "−$120.00"
	AmountClass string // "txn-amount--pos" | "txn-amount--neg" | "txn-amount--credit"
	Status      string // "completed" | "pending" | "failed" | "cancelled"
	DateLabel   string
	Created     string
	Updated     string
}

type tenantTxnStoreType struct {
	mu           sync.RWMutex
	transactions map[string][]TenantTxn
}

var tenantTxnStore = &tenantTxnStoreType{
	transactions: map[string][]TenantTxn{
		"buyvm": {
			{ID: newID(), Reference: "INV-1091", Client: "James Wright", ClientCo: "Acme Inc.", Description: "Monthly — Production cluster · 8 vCPU, 32 GB (Nov)", TxnType: "charge", Amount: "+$480.00", AmountClass: "txn-amount--pos", Status: "completed", DateLabel: "2024-11-03 14:22", Created: "2024-11-03T14:22:00Z", Updated: "2024-11-03T14:22:00Z"},
			{ID: newID(), Reference: "INV-1090", Client: "Alice Monroe", ClientCo: "Monroe LLC", Description: "Monthly — us-east-1 workspace · 4 vCPU, 16 GB (Nov)", TxnType: "charge", Amount: "+$320.00", AmountClass: "txn-amount--pos", Status: "completed", DateLabel: "2024-11-03 12:05", Created: "2024-11-03T12:05:00Z", Updated: "2024-11-03T12:05:00Z"},
			{ID: newID(), Reference: "INV-1089", Client: "Faisal Al-Amin", ClientCo: "Al-Amin Corp", Description: "Annual plan renewal — Business tier · 32 vCPU, 128 GB, 1 TB SSD", TxnType: "charge", Amount: "+$2,160.00", AmountClass: "txn-amount--pos", Status: "pending", DateLabel: "2024-11-02 09:47", Created: "2024-11-02T09:47:00Z", Updated: "2024-11-02T09:47:00Z"},
			{ID: newID(), Reference: "CR-0042", Client: "Lena Bauer", ClientCo: "BauerHosts", Description: "SLA credit — 27 min compute outage, eu-central-1 (Nov 12)", TxnType: "credit", Amount: "−$120.00", AmountClass: "txn-amount--credit", Status: "completed", DateLabel: "2024-11-01 18:30", Created: "2024-11-01T18:30:00Z", Updated: "2024-11-01T18:30:00Z"},
			{ID: newID(), Reference: "RF-0019", Client: "Tom Fischer", ClientCo: "Fischer GmbH", Description: "Refund — cancelled managed backup addon (90-day retention)", TxnType: "refund", Amount: "−$80.00", AmountClass: "txn-amount--neg", Status: "completed", DateLabel: "2024-10-31 11:15", Created: "2024-10-31T11:15:00Z", Updated: "2024-10-31T11:15:00Z"},
			{ID: newID(), Reference: "INV-1088", Client: "Ben Okonkwo", ClientCo: "Okonkwo Tech", Description: "Metered — 842 compute-hours + 2.4 TB egress (Oct)", TxnType: "metered", Amount: "+$120.00", AmountClass: "txn-amount--pos", Status: "completed", DateLabel: "2024-10-30 16:02", Created: "2024-10-30T16:02:00Z", Updated: "2024-10-30T16:02:00Z"},
			{ID: newID(), Reference: "INV-1087", Client: "Chloe Martin", ClientCo: "MartinDev", Description: "Add-on — dedicated IPv4 block /28 (16 IPs)", TxnType: "charge", Amount: "+$40.00", AmountClass: "txn-amount--pos", Status: "failed", DateLabel: "2024-10-28 08:55", Created: "2024-10-28T08:55:00Z", Updated: "2024-10-28T08:55:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Reference: "INV-0221", Client: "NocAdmin", ClientCo: "ABRNOC", Description: "Monthly colocation fee — rack 7A", TxnType: "charge", Amount: "+$850.00", AmountClass: "txn-amount--pos", Status: "completed", DateLabel: "2024-11-01 09:00", Created: "2024-11-01T09:00:00Z", Updated: "2024-11-01T09:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Reference: "INV-2041", Client: "Diego Vega", ClientCo: "Vega Systems", Description: "Kubernetes managed cluster — 10 worker nodes (Nov)", TxnType: "charge", Amount: "+$1,200.00", AmountClass: "txn-amount--pos", Status: "completed", DateLabel: "2024-11-02 10:00", Created: "2024-11-02T10:00:00Z", Updated: "2024-11-02T10:00:00Z"},
			{ID: newID(), Reference: "CR-0011", Client: "Santiago López", ClientCo: "López Cloud", Description: "SLA credit — worker node failure (INC-2026-012)", TxnType: "credit", Amount: "−$240.00", AmountClass: "txn-amount--credit", Status: "pending", DateLabel: "2024-10-30 14:00", Created: "2024-10-30T14:00:00Z", Updated: "2024-10-30T14:00:00Z"},
		},
	},
}

func (s *tenantTxnStoreType) list(slug string) []TenantTxn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.transactions[slug]
	out := make([]TenantTxn, len(src))
	copy(out, src)
	return out
}

func (s *tenantTxnStoreType) updateStatus(slug, id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.transactions[slug] {
		if t.ID == id {
			s.transactions[slug][i].Status = status
			s.transactions[slug][i].Updated = time.Now().Format(time.RFC3339)
			return
		}
	}
}

func (s *tenantTxnStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.transactions[slug]
	for i, t := range list {
		if t.ID == id {
			s.transactions[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterTenantTxns(all []TenantTxn, search, txnType, status string) []TenantTxn {
	var out []TenantTxn
	q := strings.ToLower(search)
	for _, t := range all {
		if txnType != "" && t.TxnType != txnType {
			continue
		}
		if status != "" && t.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(t.Reference + " " + t.Client + " " + t.Description)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, t)
	}
	return out
}

func dbTxnToLocal(t db.TenantTransaction) TenantTxn {
	var amount string
	var amountClass string
	switch t.TxnType {
	case "credit":
		amount = fmt.Sprintf("−$%.2f", t.Amount)
		amountClass = "txn-amount--credit"
	case "refund":
		amount = fmt.Sprintf("−$%.2f", t.Amount)
		amountClass = "txn-amount--neg"
	default:
		amount = fmt.Sprintf("+$%.2f", t.Amount)
		amountClass = "txn-amount--pos"
	}
	return TenantTxn{
		ID:          fmt.Sprintf("%d", t.ID),
		Reference:   t.Reference,
		Client:      t.Client,
		ClientCo:    t.ClientCo,
		Description: t.Description,
		TxnType:     t.TxnType,
		Amount:      amount,
		AmountClass: amountClass,
		Status:      t.Status,
		DateLabel:   t.CreatedAt.Format("2006-01-02 15:04"),
		Created:     t.CreatedAt.Format(time.RFC3339),
		Updated:     t.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Page data ─────────────────────────────────────────────────

type transactionsInnerData struct {
	Transactions []TenantTxn
	TenantSlug   string
	Search       string
	TypeFilter   string
	StatusFilter string
}

func renderTransactionsPage(w http.ResponseWriter, r *http.Request, tenant Tenant, user User, d transactionsInnerData) {
	innerTmpl, err := template.ParseFiles(filepath.Join("templates", "tenant", "transactions_content.html"))
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
		ActiveNav: "transactions", PageTitle: "Transactions",
		IsSysAdmin: user.Role == RoleSysAdmin,
		Content:    template.HTML(buf.String()),
	}
	if err := shellTmpl.ExecuteTemplate(w, "tenant-base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ── Handlers ─────────────────────────────────────────────────

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := getSession(r)
	tenant, _ := tenantFromContext(r)
	q := r.URL.Query().Get("q")
	txnType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	var txns []TenantTxn
	if Pools != nil && tenant.DBID > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		dts, err := db.ListTenantTransactions(ctx, Pools.Admin, tenant.DBID, q, txnType, status)
		if err != nil {
			log.Printf("transactionsHandler: ListTenantTransactions: %v", err)
			txns = filterTenantTxns(tenantTxnStore.list(tenant.Slug), q, txnType, status)
		} else {
			txns = make([]TenantTxn, len(dts))
			for i, dt := range dts {
				txns[i] = dbTxnToLocal(dt)
			}
		}
	} else {
		txns = filterTenantTxns(tenantTxnStore.list(tenant.Slug), q, txnType, status)
	}

	renderTransactionsPage(w, r, tenant, user, transactionsInnerData{
		Transactions: txns,
		TenantSlug:   tenant.Slug,
		Search:       q,
		TypeFilter:   txnType,
		StatusFilter: status,
	})
}

func transactionStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")
	newStatus := r.FormValue("status")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.UpdateTenantTransactionStatus(ctx, Pools.Admin, tenant.DBID, id, newStatus); err != nil {
				log.Printf("transactionStatusHandler: %v", err)
			}
		}
	} else {
		tenantTxnStore.updateStatus(tenant.Slug, idStr, newStatus)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/transactions/", http.StatusSeeOther)
}

func transactionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tenant, _ := tenantFromContext(r)
	idStr := r.PathValue("id")

	if Pools != nil && tenant.DBID > 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := db.DeleteTenantTransaction(ctx, Pools.Admin, tenant.DBID, id); err != nil {
				log.Printf("transactionDeleteHandler: %v", err)
			}
		}
	} else {
		tenantTxnStore.delete(tenant.Slug, idStr)
	}
	http.Redirect(w, r, "/t/"+tenant.Slug+"/transactions/", http.StatusSeeOther)
}

func RegisterTenantTransactionsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /t/{slug}/transactions/", tenantMiddleware(transactionsHandler))
	mux.HandleFunc("POST /t/{slug}/transactions/{id}/status/", tenantMiddleware(transactionStatusHandler))
	mux.HandleFunc("POST /t/{slug}/transactions/{id}/delete/", tenantMiddleware(transactionDeleteHandler))
}
