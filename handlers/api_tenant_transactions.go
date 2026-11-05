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
	transactions: map[string][]Transaction{},
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
