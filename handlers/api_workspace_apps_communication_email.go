package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Communication: Transactional Email
// apps/communication/transactional-email
// ════════════════════════════════════════════════════════════════

type TransactionalEmail struct {
	ID             string
	Name           string
	Domain         string
	SenderAddress  string
	DKIMStatus     string
	SPFStatus      string
	DMARCStatus    string
	DailySent      int
	DailyLimit     int
	DeliveryRate   float64
	BounceRate     float64
	OpenRate       float64
	WebhookURL     string
	APIKeyPrefix   string
	Status         string
	Updated        string
}

type transactionalEmailStoreType struct {
	mu    sync.RWMutex
	items map[string][]TransactionalEmail
}

var transactionalEmailStore = &transactionalEmailStoreType{
	items: map[string][]TransactionalEmail{
		"buyvm": {
			{ID: newID(), Name: "order-notifications", Domain: "mail.buyvm-demo.com", SenderAddress: "orders@mail.buyvm-demo.com", DKIMStatus: "Verified", SPFStatus: "Pass", DMARCStatus: "Pass", DailySent: 1250, DailyLimit: 10000, DeliveryRate: 99.2, BounceRate: 0.3, OpenRate: 42.5, WebhookURL: "https://api.buyvm-demo.com/webhooks/email", APIKeyPrefix: "bvm_sk_prod_", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "password-resets", Domain: "mail.buyvm-demo.com", SenderAddress: "security@mail.buyvm-demo.com", DKIMStatus: "Verified", SPFStatus: "Pass", DMARCStatus: "Pass", DailySent: 85, DailyLimit: 5000, DeliveryRate: 99.8, BounceRate: 0.1, OpenRate: 78.3, WebhookURL: "https://api.buyvm-demo.com/webhooks/email", APIKeyPrefix: "bvm_sk_prod_", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "marketing-campaigns", Domain: "marketing.buyvm-demo.com", SenderAddress: "hello@marketing.buyvm-demo.com", DKIMStatus: "Verified", SPFStatus: "Pass", DMARCStatus: "Pass", DailySent: 4500, DailyLimit: 50000, DeliveryRate: 97.8, BounceRate: 1.2, OpenRate: 22.1, WebhookURL: "https://api.buyvm-demo.com/webhooks/marketing", APIKeyPrefix: "bvm_sk_mkt_", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "dev-testing", Domain: "sandbox.buyvm-demo.com", SenderAddress: "test@sandbox.buyvm-demo.com", DKIMStatus: "Pending", SPFStatus: "Pending", DMARCStatus: "None", DailySent: 12, DailyLimit: 100, DeliveryRate: 95.0, BounceRate: 5.0, OpenRate: 0.0, WebhookURL: "—", APIKeyPrefix: "bvm_sk_test_", Status: "sandbox", Updated: "2026-04-14"},
		},
		"hosting": {
			{ID: newID(), Name: "platform-alerts", Domain: "notify.hosting-demo.com", SenderAddress: "alerts@notify.hosting-demo.com", DKIMStatus: "Verified", SPFStatus: "Pass", DMARCStatus: "Pass", DailySent: 3200, DailyLimit: 25000, DeliveryRate: 99.5, BounceRate: 0.2, OpenRate: 65.0, WebhookURL: "https://api.hosting-demo.com/hooks/email", APIKeyPrefix: "cz_sk_prod_", Status: "active", Updated: "2026-04-17"},
			{ID: newID(), Name: "user-onboarding", Domain: "notify.hosting-demo.com", SenderAddress: "welcome@notify.hosting-demo.com", DKIMStatus: "Verified", SPFStatus: "Pass", DMARCStatus: "Pass", DailySent: 420, DailyLimit: 10000, DeliveryRate: 99.1, BounceRate: 0.5, OpenRate: 55.8, WebhookURL: "https://api.hosting-demo.com/hooks/email", APIKeyPrefix: "cz_sk_prod_", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "billing-invoices", Domain: "billing.hosting-demo.com", SenderAddress: "invoices@billing.hosting-demo.com", DKIMStatus: "Verified", SPFStatus: "Pass", DMARCStatus: "Pass", DailySent: 150, DailyLimit: 5000, DeliveryRate: 99.9, BounceRate: 0.05, OpenRate: 70.2, WebhookURL: "https://api.hosting-demo.com/hooks/billing", APIKeyPrefix: "cz_sk_bill_", Status: "active", Updated: "2026-04-17"},
		},
	},
}

func (s *transactionalEmailStoreType) list(slug string) []TransactionalEmail {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]TransactionalEmail, len(src))
	copy(out, src)
	return out
}

func (s *transactionalEmailStoreType) getByID(slug, id string) (TransactionalEmail, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.items[slug] {
		if e.ID == id {
			return e, true
		}
	}
	return TransactionalEmail{}, false
}

func (s *transactionalEmailStoreType) add(slug string, e TransactionalEmail) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], e)
}

func (s *transactionalEmailStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, e := range list {
		if e.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiTransactionalEmail struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Domain         string  `json:"domain"`
	SenderAddress  string  `json:"sender_address"`
	DKIMStatus     string  `json:"dkim_status"`
	SPFStatus      string  `json:"spf_status"`
	DMARCStatus    string  `json:"dmarc_status"`
	DailySent      int     `json:"daily_sent"`
	DailyLimit     int     `json:"daily_limit"`
	DeliveryRate   float64 `json:"delivery_rate"`
	BounceRate     float64 `json:"bounce_rate"`
	OpenRate       float64 `json:"open_rate"`
	WebhookURL     string  `json:"webhook_url"`
	APIKeyPrefix   string  `json:"api_key_prefix"`
	Status         string  `json:"status"`
	Updated        string  `json:"updated"`
}

type apiTransactionalEmailList struct {
	Services []apiTransactionalEmail `json:"services"`
	Total    int                     `json:"total"`
}

func toAPITransactionalEmail(e TransactionalEmail) apiTransactionalEmail {
	return apiTransactionalEmail{
		ID: e.ID, Name: e.Name, Domain: e.Domain,
		SenderAddress: e.SenderAddress, DKIMStatus: e.DKIMStatus,
		SPFStatus: e.SPFStatus, DMARCStatus: e.DMARCStatus,
		DailySent: e.DailySent, DailyLimit: e.DailyLimit,
		DeliveryRate: e.DeliveryRate, BounceRate: e.BounceRate,
		OpenRate: e.OpenRate, WebhookURL: e.WebhookURL,
		APIKeyPrefix: e.APIKeyPrefix, Status: e.Status,
		Updated: e.Updated,
	}
}

func filterTransactionalEmails(all []TransactionalEmail, search, status string) []TransactionalEmail {
	var out []TransactionalEmail
	q := strings.ToLower(search)
	for _, e := range all {
		if status != "" && e.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(e.Name + " " + e.Domain + " " + e.SenderAddress)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func apiWSAppsCommEmailListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterTransactionalEmails(transactionalEmailStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiTransactionalEmail, len(items))
	for i, e := range items {
		out[i] = toAPITransactionalEmail(e)
	}
	apiOK(w, apiTransactionalEmailList{Services: out, Total: len(out)})
}

type apiCreateTransactionalEmailRequest struct {
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	SenderAddress string `json:"sender_address"`
	WebhookURL    string `json:"webhook_url"`
}

func apiWSAppsCommEmailCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateTransactionalEmailRequest
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
		apiError(w, http.StatusBadRequest, "Domain is required.")
		return
	}
	senderAddr := strings.TrimSpace(req.SenderAddress)
	if senderAddr == "" {
		senderAddr = "noreply@" + domain
	}
	webhookURL := strings.TrimSpace(req.WebhookURL)
	if webhookURL == "" {
		webhookURL = "—"
	}
	svc := TransactionalEmail{
		ID: newID(), Name: name, Domain: domain,
		SenderAddress: senderAddr, DKIMStatus: "Pending",
		SPFStatus: "Pending", DMARCStatus: "None",
		DailySent: 0, DailyLimit: 100,
		DeliveryRate: 0, BounceRate: 0, OpenRate: 0,
		WebhookURL: webhookURL, APIKeyPrefix: "key_",
		Status: "pending_verification", Updated: "2026-04-17",
	}
	transactionalEmailStore.add(slug, svc)
	apiCreated(w, toAPITransactionalEmail(svc))
}

func apiWSAppsCommEmailGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	svc, found := transactionalEmailStore.getByID(slug, r.PathValue("eid"))
	if !found {
		apiError(w, http.StatusNotFound, "Email service not found.")
		return
	}
	apiOK(w, toAPITransactionalEmail(svc))
}

func apiWSAppsCommEmailDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := transactionalEmailStore.getByID(slug, r.PathValue("eid")); !found {
		apiError(w, http.StatusNotFound, "Email service not found.")
		return
	}
	transactionalEmailStore.delete(slug, r.PathValue("eid"))
	apiOK(w, map[string]any{"ok": true, "message": "Email service deleted."})
}
