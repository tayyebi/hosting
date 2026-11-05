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
	PublishRate      int     // messages per second
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
	items: map[string][]ProMessageBroker{
		"buyvm": {
			{ID: newID(), Name: "orders-broker", Protocol: "AMQP 0.9.1", Engine: "RabbitMQ", Version: "3.13.2", QueueCount: 24, ExchangeCount: 8, ConsumerCount: 12, MessagesReady: 1840, MessagesUnacked: 64, PublishRate: 2400, ConsumeRate: 2380, StorageGB: 4.2, RetentionHours: 72, TLSEnabled: true, ClusterNodes: 3, HighAvailability: true, Region: "us-east-1", CostPerMonth: 180.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "event-bus", Protocol: "AMQP 0.9.1", Engine: "RabbitMQ", Version: "3.13.2", QueueCount: 42, ExchangeCount: 16, ConsumerCount: 28, MessagesReady: 420, MessagesUnacked: 12, PublishRate: 8400, ConsumeRate: 8390, StorageGB: 8.6, RetentionHours: 48, TLSEnabled: true, ClusterNodes: 3, HighAvailability: true, Region: "us-east-1", CostPerMonth: 240.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "iot-telemetry", Protocol: "MQTT 5", Engine: "RabbitMQ", Version: "3.13.2", QueueCount: 128, ExchangeCount: 4, ConsumerCount: 8, MessagesReady: 42000, MessagesUnacked: 860, PublishRate: 24000, ConsumeRate: 23800, StorageGB: 16.4, RetentionHours: 24, TLSEnabled: true, ClusterNodes: 5, HighAvailability: true, Region: "us-east-1", CostPerMonth: 420.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "notification-queue", Protocol: "AMQP 0.9.1", Engine: "RabbitMQ", Version: "3.13.2", QueueCount: 6, ExchangeCount: 2, ConsumerCount: 4, MessagesReady: 86, MessagesUnacked: 2, PublishRate: 640, ConsumeRate: 638, StorageGB: 0.8, RetentionHours: 168, TLSEnabled: true, ClusterNodes: 2, HighAvailability: true, Region: "us-east-1", CostPerMonth: 65.00, Status: "running", Updated: "2026-04-17"},
		},
		"hosting": {
			{ID: newID(), Name: "platform-events", Protocol: "AMQP 0.9.1", Engine: "RabbitMQ", Version: "3.13.2", QueueCount: 186, ExchangeCount: 42, ConsumerCount: 96, MessagesReady: 12400, MessagesUnacked: 480, PublishRate: 64000, ConsumeRate: 63800, StorageGB: 42.0, RetentionHours: 72, TLSEnabled: true, ClusterNodes: 7, HighAvailability: true, Region: "eu-central-1", CostPerMonth: 1200.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "payment-processor", Protocol: "AMQP 0.9.1", Engine: "RabbitMQ", Version: "3.13.2", QueueCount: 12, ExchangeCount: 4, ConsumerCount: 8, MessagesReady: 24, MessagesUnacked: 4, PublishRate: 1200, ConsumeRate: 1200, StorageGB: 2.4, RetentionHours: 720, TLSEnabled: true, ClusterNodes: 3, HighAvailability: true, Region: "eu-central-1", CostPerMonth: 240.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "analytics-ingestion", Protocol: "AMQP 1.0", Engine: "ActiveMQ", Version: "6.1.0", QueueCount: 8, ExchangeCount: 0, ConsumerCount: 16, MessagesReady: 84000, MessagesUnacked: 2400, PublishRate: 120000, ConsumeRate: 118000, StorageGB: 86.0, RetentionHours: 48, TLSEnabled: true, ClusterNodes: 5, HighAvailability: true, Region: "eu-central-1", CostPerMonth: 640.00, Status: "running", Updated: "2026-04-18"},
			{ID: newID(), Name: "staging-broker", Protocol: "AMQP 0.9.1", Engine: "RabbitMQ", Version: "3.13.2", QueueCount: 24, ExchangeCount: 8, ConsumerCount: 6, MessagesReady: 0, MessagesUnacked: 0, PublishRate: 120, ConsumeRate: 120, StorageGB: 0.4, RetentionHours: 24, TLSEnabled: false, ClusterNodes: 1, HighAvailability: false, Region: "eu-central-1", CostPerMonth: 35.00, Status: "running", Updated: "2026-04-16"},
		},
	},
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
		Region: region, CostPerMonth: float64(nodes) * 45.00,
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
