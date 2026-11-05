package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace DevOps API — API Management: Gateway
// devops/api/gateway
// ════════════════════════════════════════════════════════════════

type APIGateway struct {
	ID        string
	Name      string
	Protocol  string
	Upstream  string
	RateLimit string
	Auth      string
	Status    string
	Requests  string
	Updated   string
}

type apiGatewayStoreType struct {
	mu       sync.RWMutex
	gateways map[string][]APIGateway
}

var apiGatewayStore = &apiGatewayStoreType{
	gateways: map[string][]APIGateway{},
}

func (s *apiGatewayStoreType) list(slug string) []APIGateway {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.gateways[slug]
	out := make([]APIGateway, len(src))
	copy(out, src)
	return out
}

func (s *apiGatewayStoreType) getByID(slug, id string) (APIGateway, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, g := range s.gateways[slug] {
		if g.ID == id {
			return g, true
		}
	}
	return APIGateway{}, false
}

func (s *apiGatewayStoreType) add(slug string, g APIGateway) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gateways[slug] = append(s.gateways[slug], g)
}

func (s *apiGatewayStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.gateways[slug]
	for i, g := range list {
		if g.ID == id {
			s.gateways[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiAPIGateway struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Upstream  string `json:"upstream"`
	RateLimit string `json:"rate_limit"`
	Auth      string `json:"auth"`
	Status    string `json:"status"`
	Requests  string `json:"requests"`
	Updated   string `json:"updated"`
}

type apiAPIGatewayList struct {
	Gateways []apiAPIGateway `json:"gateways"`
	Total    int             `json:"total"`
}

func toAPIAPIGateway(g APIGateway) apiAPIGateway {
	return apiAPIGateway{
		ID: g.ID, Name: g.Name, Protocol: g.Protocol,
		Upstream: g.Upstream, RateLimit: g.RateLimit, Auth: g.Auth,
		Status: g.Status, Requests: g.Requests, Updated: g.Updated,
	}
}

func filterAPIGateways(all []APIGateway, search, status string) []APIGateway {
	var out []APIGateway
	q := strings.ToLower(search)
	for _, g := range all {
		if status != "" && g.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(g.Name + " " + g.Protocol + " " + g.Upstream + " " + g.Auth)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, g)
	}
	return out
}

func apiWSDevOpsAPIGatewayListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	gateways := filterAPIGateways(apiGatewayStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiAPIGateway, len(gateways))
	for i, g := range gateways {
		out[i] = toAPIAPIGateway(g)
	}
	apiOK(w, apiAPIGatewayList{Gateways: out, Total: len(out)})
}

type apiCreateAPIGatewayRequest struct {
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Upstream  string `json:"upstream"`
	RateLimit string `json:"rate_limit"`
	Auth      string `json:"auth"`
}

func apiWSDevOpsAPIGatewayCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateAPIGatewayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	protocol := strings.TrimSpace(req.Protocol)
	if protocol == "" {
		protocol = "REST"
	}
	upstream := strings.TrimSpace(req.Upstream)
	if upstream == "" {
		upstream = "service:8080"
	}
	rateLimit := strings.TrimSpace(req.RateLimit)
	if rateLimit == "" {
		rateLimit = "1000 req/min"
	}
	auth := strings.TrimSpace(req.Auth)
	if auth == "" {
		auth = "API Key"
	}
	gw := APIGateway{
		ID: newID(), Name: name, Protocol: protocol,
		Upstream: upstream, RateLimit: rateLimit, Auth: auth,
		Status: "active", Requests: "0/day", Updated: "2026-04-17",
	}
	apiGatewayStore.add(slug, gw)
	apiCreated(w, toAPIAPIGateway(gw))
}

func apiWSDevOpsAPIGatewayGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	gw, found := apiGatewayStore.getByID(slug, r.PathValue("gid"))
	if !found {
		apiError(w, http.StatusNotFound, "API gateway not found.")
		return
	}
	apiOK(w, toAPIAPIGateway(gw))
}

func apiWSDevOpsAPIGatewayDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := apiGatewayStore.getByID(slug, r.PathValue("gid")); !found {
		apiError(w, http.StatusNotFound, "API gateway not found.")
		return
	}
	apiGatewayStore.delete(slug, r.PathValue("gid"))
	apiOK(w, map[string]any{"ok": true, "message": "API gateway deleted."})
}
