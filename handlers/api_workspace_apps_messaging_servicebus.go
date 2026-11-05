package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Messaging: Service Bus
// apps/messaging/servicebus
// ════════════════════════════════════════════════════════════════

type ServiceBus struct {
	ID            string
	Name          string
	Tier          string
	Namespaces    int
	Topics        int
	Subscriptions int
	Throughput    string
	Status        string
	Updated       string
}

type serviceBusStoreType struct {
	mu    sync.RWMutex
	buses map[string][]ServiceBus
}

var serviceBusStore = &serviceBusStoreType{
	buses: map[string][]ServiceBus{
		"buyvm": {
			{ID: newID(), Name: "production-bus", Tier: "Premium", Namespaces: 4, Topics: 28, Subscriptions: 96, Throughput: "18,000 msg/s", Status: "active", Updated: "2026-04-16"},
			{ID: newID(), Name: "staging-bus", Tier: "Standard", Namespaces: 2, Topics: 28, Subscriptions: 56, Throughput: "2,400 msg/s", Status: "active", Updated: "2026-04-14"},
			{ID: newID(), Name: "integration-bus", Tier: "Premium", Namespaces: 3, Topics: 15, Subscriptions: 42, Throughput: "5,600 msg/s", Status: "active", Updated: "2026-04-12"},
			{ID: newID(), Name: "analytics-bus", Tier: "Standard", Namespaces: 1, Topics: 8, Subscriptions: 24, Throughput: "9,200 msg/s", Status: "active", Updated: "2026-04-10"},
			{ID: newID(), Name: "legacy-esb", Tier: "Basic", Namespaces: 1, Topics: 5, Subscriptions: 10, Throughput: "320 msg/s", Status: "deprecated", Updated: "2025-09-01"},
		},
		"hosting": {
			{ID: newID(), Name: "platform-bus", Tier: "Premium", Namespaces: 3, Topics: 22, Subscriptions: 68, Throughput: "14,000 msg/s", Status: "active", Updated: "2026-04-13"},
			{ID: newID(), Name: "partner-integration-bus", Tier: "Standard", Namespaces: 2, Topics: 12, Subscriptions: 36, Throughput: "3,800 msg/s", Status: "active", Updated: "2026-04-08"},
			{ID: newID(), Name: "internal-events-bus", Tier: "Standard", Namespaces: 1, Topics: 18, Subscriptions: 45, Throughput: "7,200 msg/s", Status: "active", Updated: "2026-04-04"},
		},
	},
}

func (s *serviceBusStoreType) list(slug string) []ServiceBus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.buses[slug]
	out := make([]ServiceBus, len(src))
	copy(out, src)
	return out
}

func (s *serviceBusStoreType) getByID(slug, id string) (ServiceBus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.buses[slug] {
		if b.ID == id {
			return b, true
		}
	}
	return ServiceBus{}, false
}

func (s *serviceBusStoreType) add(slug string, b ServiceBus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buses[slug] = append(s.buses[slug], b)
}

func (s *serviceBusStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.buses[slug]
	for i, b := range list {
		if b.ID == id {
			s.buses[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiServiceBus struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Tier          string `json:"tier"`
	Namespaces    int    `json:"namespaces"`
	Topics        int    `json:"topics"`
	Subscriptions int    `json:"subscriptions"`
	Throughput    string `json:"throughput"`
	Status        string `json:"status"`
	Updated       string `json:"updated"`
}

type apiServiceBusList struct {
	Buses []apiServiceBus `json:"buses"`
	Total int             `json:"total"`
}

func toAPIServiceBus(b ServiceBus) apiServiceBus {
	return apiServiceBus{
		ID: b.ID, Name: b.Name, Tier: b.Tier,
		Namespaces: b.Namespaces, Topics: b.Topics,
		Subscriptions: b.Subscriptions, Throughput: b.Throughput,
		Status: b.Status, Updated: b.Updated,
	}
}

func filterServiceBuses(all []ServiceBus, search, status string) []ServiceBus {
	var out []ServiceBus
	q := strings.ToLower(search)
	for _, b := range all {
		if status != "" && b.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(b.Name + " " + b.Tier)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, b)
	}
	return out
}

func apiWSAppsMessagingServiceBusListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	buses := filterServiceBuses(serviceBusStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiServiceBus, len(buses))
	for i, b := range buses {
		out[i] = toAPIServiceBus(b)
	}
	apiOK(w, apiServiceBusList{Buses: out, Total: len(out)})
}

type apiCreateServiceBusRequest struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
}

func apiWSAppsMessagingServiceBusCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateServiceBusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	tier := strings.TrimSpace(req.Tier)
	if tier == "" {
		tier = "Standard"
	}
	bus := ServiceBus{
		ID: newID(), Name: name, Tier: tier,
		Namespaces: 1, Topics: 0, Subscriptions: 0,
		Throughput: "0 msg/s", Status: "active", Updated: "2026-04-17",
	}
	serviceBusStore.add(slug, bus)
	apiCreated(w, toAPIServiceBus(bus))
}

func apiWSAppsMessagingServiceBusGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	bus, found := serviceBusStore.getByID(slug, r.PathValue("bid"))
	if !found {
		apiError(w, http.StatusNotFound, "Service bus not found.")
		return
	}
	apiOK(w, toAPIServiceBus(bus))
}

func apiWSAppsMessagingServiceBusDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := serviceBusStore.getByID(slug, r.PathValue("bid")); !found {
		apiError(w, http.StatusNotFound, "Service bus not found.")
		return
	}
	serviceBusStore.delete(slug, r.PathValue("bid"))
	apiOK(w, map[string]any{"ok": true, "message": "Service bus deleted."})
}
