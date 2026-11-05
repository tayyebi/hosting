package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ── Transaction store ──────────────────────────────────────

type Transaction struct {
	ID          string
	Type        string // "payment" | "refund" | "credit" | "debit" | "adjustment"
	Status      string // "completed" | "pending" | "failed" | "reversed"
	Amount      string
	Currency    string
	Description string
	Client      string
	InvoiceID   string
	Method      string // "credit_card" | "wire" | "paypal" | "crypto" | "balance"
	Created     string
}

type transactionStoreType struct {
	mu           sync.RWMutex
	transactions map[string][]Transaction
}

var transactionStore = &transactionStoreType{
	transactions: map[string][]Transaction{
		"buyvm": {
			{ID: newID(), Type: "payment", Status: "completed", Amount: "249.00", Currency: "USD", Description: "Monthly hosting — Pro plan", Client: "Acme Corp", InvoiceID: "INV-2026-0412", Method: "credit_card", Created: "2026-04-15T10:23:00Z"},
			{ID: newID(), Type: "payment", Status: "completed", Amount: "1,200.00", Currency: "USD", Description: "Annual SSL certificate bundle", Client: "Globex Inc", InvoiceID: "INV-2026-0408", Method: "wire", Created: "2026-04-12T08:45:00Z"},
			{ID: newID(), Type: "refund", Status: "completed", Amount: "-75.00", Currency: "USD", Description: "Prorated refund — service downtime credit", Client: "Initech Ltd", InvoiceID: "INV-2026-0401", Method: "credit_card", Created: "2026-04-10T14:30:00Z"},
			{ID: newID(), Type: "payment", Status: "pending", Amount: "499.00", Currency: "USD", Description: "Dedicated server — Q2 prepay", Client: "Wayne Enterprises", InvoiceID: "INV-2026-0415", Method: "wire", Created: "2026-04-16T09:00:00Z"},
			{ID: newID(), Type: "credit", Status: "completed", Amount: "50.00", Currency: "USD", Description: "Referral program credit", Client: "Acme Corp", InvoiceID: "", Method: "balance", Created: "2026-04-08T12:00:00Z"},
			{ID: newID(), Type: "payment", Status: "failed", Amount: "149.00", Currency: "USD", Description: "VPS upgrade — card declined", Client: "Stark Industries", InvoiceID: "INV-2026-0410", Method: "credit_card", Created: "2026-04-09T16:20:00Z"},
			{ID: newID(), Type: "debit", Status: "completed", Amount: "-25.00", Currency: "USD", Description: "Overage — bandwidth 150GB extra", Client: "Globex Inc", InvoiceID: "INV-2026-0413", Method: "balance", Created: "2026-04-14T00:00:00Z"},
		},
		"abrnoc": {
			{ID: newID(), Type: "payment", Status: "completed", Amount: "89.00", Currency: "EUR", Description: "VPS Standard — monthly", Client: "Schmidt GmbH", InvoiceID: "INV-2026-0205", Method: "paypal", Created: "2026-04-11T07:30:00Z"},
			{ID: newID(), Type: "payment", Status: "completed", Amount: "320.00", Currency: "EUR", Description: "Managed Kubernetes cluster", Client: "Dupont SA", InvoiceID: "INV-2026-0209", Method: "credit_card", Created: "2026-04-13T11:15:00Z"},
			{ID: newID(), Type: "refund", Status: "pending", Amount: "-45.00", Currency: "EUR", Description: "SLA credit — April incident", Client: "Schmidt GmbH", InvoiceID: "INV-2026-0205", Method: "balance", Created: "2026-04-15T09:00:00Z"},
		},
		"hosting": {
			{ID: newID(), Type: "payment", Status: "completed", Amount: "599.00", Currency: "USD", Description: "GPU instance — ml.g5.xlarge monthly", Client: "DeepNet AI", InvoiceID: "INV-2026-0310", Method: "credit_card", Created: "2026-04-14T06:00:00Z"},
			{ID: newID(), Type: "payment", Status: "completed", Amount: "2,400.00", Currency: "USD", Description: "Enterprise support — annual", Client: "Quantum Labs", InvoiceID: "INV-2026-0312", Method: "wire", Created: "2026-04-10T10:00:00Z"},
			{ID: newID(), Type: "adjustment", Status: "completed", Amount: "-100.00", Currency: "USD", Description: "Billing correction — duplicate charge", Client: "DeepNet AI", InvoiceID: "INV-2026-0308", Method: "balance", Created: "2026-04-07T15:45:00Z"},
		},
	},
}

func (s *transactionStoreType) list(slug string) []Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.transactions[slug]
	out := make([]Transaction, len(src))
	copy(out, src)
	return out
}

func (s *transactionStoreType) getByID(slug, id string) (Transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, tx := range s.transactions[slug] {
		if tx.ID == id {
			return tx, true
		}
	}
	return Transaction{}, false
}

func (s *transactionStoreType) add(slug string, tx Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactions[slug] = append(s.transactions[slug], tx)
}

func (s *transactionStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.transactions[slug]
	for i, tx := range list {
		if tx.ID == id {
			s.transactions[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

func filterTransactions(all []Transaction, search, txType, status string) []Transaction {
	var out []Transaction
	q := strings.ToLower(search)
	for _, tx := range all {
		if txType != "" && tx.Type != txType {
			continue
		}
		if status != "" && tx.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(tx.Description + " " + tx.Client + " " + tx.InvoiceID + " " + tx.Amount)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, tx)
	}
	return out
}

// ── GET /api/tenant/{slug}/transactions ────────────────────

type apiTransaction struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	Client      string `json:"client"`
	InvoiceID   string `json:"invoice_id"`
	Method      string `json:"method"`
	Created     string `json:"created"`
}

type apiTransactionsListResponse struct {
	Transactions []apiTransaction `json:"transactions"`
	Total        int              `json:"total"`
}

func apiTenantTransactionsListHandler(w http.ResponseWriter, r *http.Request) {
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
	txType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	txns := filterTransactions(transactionStore.list(slug), search, txType, status)

	out := make([]apiTransaction, len(txns))
	for i, tx := range txns {
		out[i] = apiTransaction{
			ID: tx.ID, Type: tx.Type, Status: tx.Status,
			Amount: tx.Amount, Currency: tx.Currency,
			Description: tx.Description, Client: tx.Client,
			InvoiceID: tx.InvoiceID, Method: tx.Method,
			Created: tx.Created,
		}
	}

	apiOK(w, apiTransactionsListResponse{Transactions: out, Total: len(out)})
}

// ── POST /api/tenant/{slug}/transactions ───────────────────

type apiCreateTransactionRequest struct {
	Type        string `json:"type"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	Client      string `json:"client"`
	InvoiceID   string `json:"invoice_id"`
	Method      string `json:"method"`
}

func apiTenantTransactionsCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req apiCreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}

	amount := strings.TrimSpace(req.Amount)
	description := strings.TrimSpace(req.Description)
	txType := strings.TrimSpace(req.Type)
	currency := strings.TrimSpace(req.Currency)
	client := strings.TrimSpace(req.Client)
	method := strings.TrimSpace(req.Method)

	if amount == "" {
		apiError(w, http.StatusBadRequest, "Amount is required.")
		return
	}
	if txType == "" {
		txType = "payment"
	}
	if currency == "" {
		currency = "USD"
	}

	tx := Transaction{
		ID: newID(), Type: txType, Status: "pending",
		Amount: amount, Currency: currency,
		Description: description, Client: client,
		InvoiceID: strings.TrimSpace(req.InvoiceID), Method: method,
		Created: "2026-04-17T00:00:00Z",
	}
	transactionStore.add(slug, tx)

	apiCreated(w, apiTransaction{
		ID: tx.ID, Type: tx.Type, Status: tx.Status,
		Amount: tx.Amount, Currency: tx.Currency,
		Description: tx.Description, Client: tx.Client,
		InvoiceID: tx.InvoiceID, Method: tx.Method,
		Created: tx.Created,
	})
}

// ── GET /api/tenant/{slug}/transactions/{id} ───────────────

func apiTenantTransactionsGetHandler(w http.ResponseWriter, r *http.Request) {
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
	tx, ok := transactionStore.getByID(slug, id)
	if !ok {
		apiError(w, http.StatusNotFound, "Transaction not found.")
		return
	}

	apiOK(w, apiTransaction{
		ID: tx.ID, Type: tx.Type, Status: tx.Status,
		Amount: tx.Amount, Currency: tx.Currency,
		Description: tx.Description, Client: tx.Client,
		InvoiceID: tx.InvoiceID, Method: tx.Method,
		Created: tx.Created,
	})
}

// ── DELETE /api/tenant/{slug}/transactions/{id} ────────────

func apiTenantTransactionsDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := transactionStore.getByID(slug, id); !ok {
		apiError(w, http.StatusNotFound, "Transaction not found.")
		return
	}

	transactionStore.delete(slug, id)
	apiOK(w, map[string]any{
		"ok":      true,
		"message": "Transaction deleted.",
	})
}
