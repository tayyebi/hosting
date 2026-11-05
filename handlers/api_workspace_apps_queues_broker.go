package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Apps API — Queues: Message Broker (Pro Feature)
// apps/queues/message-broker
// ════════════════════════════════════════════════════════════════

type ProMessageBroker struct {
	ID               string
	Name             string
	Protocol         string // "AMQP 0.9.1", "MQTT 5", "STOMP", "AMQP 1.0"
	Engine           string // "RabbitMQ", "ActiveMQ", "Kafka", "Redis Streams"
	Version          string
	QueueCount       int
	ExchangeCount    int
	ConsumerCount    int
	MessagesReady    int64
	MessagesUnacked  int64
	PublishRate      int // messages per second
	ConsumeRate      int
	StorageGB        float64
	RetentionHours   int
	TLSEnabled       bool
	ClusterNodes     int
	HighAvailability bool
	Region           string
	CostPerMonth     float64
	Status           string // "running", "degraded", "provisioning"
	Updated          string
}

type proMsgBrokerStoreType struct {
	mu    sync.RWMutex
	items map[string][]ProMessageBroker
}

var proMsgBrokerStore = &proMsgBrokerStoreType{
	items: map[string][]ProMessageBroker{},
}

func (s *proMsgBrokerStoreType) list(slug string) []ProMessageBroker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.items[slug]
	out := make([]ProMessageBroker, len(src))
	copy(out, src)
	return out
}

func (s *proMsgBrokerStoreType) getByID(slug, id string) (ProMessageBroker, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.items[slug] {
		if b.ID == id {
			return b, true
		}
	}
	return ProMessageBroker{}, false
}

func (s *proMsgBrokerStoreType) add(slug string, b ProMessageBroker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[slug] = append(s.items[slug], b)
}

func (s *proMsgBrokerStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[slug]
	for i, b := range list {
		if b.ID == id {
			s.items[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiProMessageBroker struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Protocol         string  `json:"protocol"`
	Engine           string  `json:"engine"`
	Version          string  `json:"version"`
	QueueCount       int     `json:"queue_count"`
	ExchangeCount    int     `json:"exchange_count"`
	ConsumerCount    int     `json:"consumer_count"`
	MessagesReady    int64   `json:"messages_ready"`
	MessagesUnacked  int64   `json:"messages_unacked"`
	PublishRate      int     `json:"publish_rate_mps"`
	ConsumeRate      int     `json:"consume_rate_mps"`
	StorageGB        float64 `json:"storage_gb"`
	RetentionHours   int     `json:"retention_hours"`
	TLSEnabled       bool    `json:"tls_enabled"`
	ClusterNodes     int     `json:"cluster_nodes"`
	HighAvailability bool    `json:"high_availability"`
	Region           string  `json:"region"`
	CostPerMonth     float64 `json:"cost_per_month"`
	Status           string  `json:"status"`
	Updated          string  `json:"updated"`
}

type apiProMessageBrokerList struct {
	Brokers []apiProMessageBroker `json:"brokers"`
	Total   int                   `json:"total"`
}

func toAPIProMsgBroker(b ProMessageBroker) apiProMessageBroker {
	return apiProMessageBroker{
		ID: b.ID, Name: b.Name, Protocol: b.Protocol,
		Engine: b.Engine, Version: b.Version,
		QueueCount: b.QueueCount, ExchangeCount: b.ExchangeCount,
		ConsumerCount: b.ConsumerCount, MessagesReady: b.MessagesReady,
		MessagesUnacked: b.MessagesUnacked, PublishRate: b.PublishRate,
		ConsumeRate: b.ConsumeRate, StorageGB: b.StorageGB,
		RetentionHours: b.RetentionHours, TLSEnabled: b.TLSEnabled,
		ClusterNodes: b.ClusterNodes, HighAvailability: b.HighAvailability,
		Region: b.Region, CostPerMonth: b.CostPerMonth,
		Status: b.Status, Updated: b.Updated,
	}
}

func filterProMsgBrokers(all []ProMessageBroker, search, status string) []ProMessageBroker {
	var out []ProMessageBroker
	q := strings.ToLower(search)
	for _, b := range all {
		if status != "" && b.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(b.Name + " " + b.Engine + " " + b.Protocol + " " + b.Region)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, b)
	}
	return out
}

func apiWSAppsQueuesBrokerListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	items := filterProMsgBrokers(proMsgBrokerStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiProMessageBroker, len(items))
	for i, b := range items {
		out[i] = toAPIProMsgBroker(b)
	}
	apiOK(w, apiProMessageBrokerList{Brokers: out, Total: len(out)})
}

type apiCreateProMsgBrokerRequest struct {
	Name             string `json:"name"`
	Protocol         string `json:"protocol"`
	Engine           string `json:"engine"`
	RetentionHours   int    `json:"retention_hours"`
	TLSEnabled       bool   `json:"tls_enabled"`
	ClusterNodes     int    `json:"cluster_nodes"`
	HighAvailability bool   `json:"high_availability"`
	Region           string `json:"region"`
}

func apiWSAppsQueuesBrokerCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateProMsgBrokerRequest
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
		protocol = "AMQP 0.9.1"
	}
	engine := strings.TrimSpace(req.Engine)
	if engine == "" {
		engine = "RabbitMQ"
	}
	retention := req.RetentionHours
	if retention <= 0 {
		retention = 72
	}
	nodes := req.ClusterNodes
	if nodes <= 0 {
		nodes = 1
	}
	region := strings.TrimSpace(req.Region)
	if region == "" {
		region = "us-east-1"
	}
	version := "3.13.2"
	if engine == "ActiveMQ" {
		version = "6.1.0"
	}
	broker := ProMessageBroker{
		ID: newID(), Name: name, Protocol: protocol,
		Engine: engine, Version: version,
		QueueCount: 0, ExchangeCount: 0, ConsumerCount: 0,
		MessagesReady: 0, MessagesUnacked: 0,
		PublishRate: 0, ConsumeRate: 0,
		StorageGB: 0, RetentionHours: retention,
		TLSEnabled: req.TLSEnabled, ClusterNodes: nodes,
		HighAvailability: req.HighAvailability,
		Region:           region, CostPerMonth: float64(nodes) * 45.00,
		Status: "provisioning", Updated: "2026-04-18",
	}
	proMsgBrokerStore.add(slug, broker)
	apiCreated(w, toAPIProMsgBroker(broker))
}

func apiWSAppsQueuesBrokerGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	b, found := proMsgBrokerStore.getByID(slug, r.PathValue("bid"))
	if !found {
		apiError(w, http.StatusNotFound, "Message broker not found.")
		return
	}
	apiOK(w, toAPIProMsgBroker(b))
}

func apiWSAppsQueuesBrokerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := proMsgBrokerStore.getByID(slug, r.PathValue("bid")); !found {
		apiError(w, http.StatusNotFound, "Message broker not found.")
		return
	}
	proMsgBrokerStore.delete(slug, r.PathValue("bid"))
	apiOK(w, map[string]any{"ok": true, "message": "Message broker deleted."})
}
