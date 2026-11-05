package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Messaging: Message Queues
// apps/messaging/queues
// ════════════════════════════════════════════════════════════════

type MessageQueue struct {
	ID         string
	Name       string
	Engine     string
	Type       string
	Messages   string
	Throughput string
	Consumers  int
	Retention  string
	Status     string
	Updated    string
}

type messageQueueStoreType struct {
	mu     sync.RWMutex
	queues map[string][]MessageQueue
}

var messageQueueStore = &messageQueueStoreType{
	queues: map[string][]MessageQueue{},
}

func (s *messageQueueStoreType) list(slug string) []MessageQueue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.queues[slug]
	out := make([]MessageQueue, len(src))
	copy(out, src)
	return out
}

func (s *messageQueueStoreType) getByID(slug, id string) (MessageQueue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, q := range s.queues[slug] {
		if q.ID == id {
			return q, true
		}
	}
	return MessageQueue{}, false
}

func (s *messageQueueStoreType) add(slug string, q MessageQueue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queues[slug] = append(s.queues[slug], q)
}

func (s *messageQueueStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.queues[slug]
	for i, q := range list {
		if q.ID == id {
			s.queues[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiMessageQueue struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Engine     string `json:"engine"`
	Type       string `json:"type"`
	Messages   string `json:"messages"`
	Throughput string `json:"throughput"`
	Consumers  int    `json:"consumers"`
	Retention  string `json:"retention"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
}

type apiMessageQueueList struct {
	Queues []apiMessageQueue `json:"queues"`
	Total  int               `json:"total"`
}

func toAPIMessageQueue(q MessageQueue) apiMessageQueue {
	return apiMessageQueue{
		ID: q.ID, Name: q.Name, Engine: q.Engine,
		Type: q.Type, Messages: q.Messages, Throughput: q.Throughput,
		Consumers: q.Consumers, Retention: q.Retention,
		Status: q.Status, Updated: q.Updated,
	}
}

func filterMessageQueues(all []MessageQueue, search, status string) []MessageQueue {
	var out []MessageQueue
	q := strings.ToLower(search)
	for _, mq := range all {
		if status != "" && mq.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(mq.Name + " " + mq.Engine + " " + mq.Type)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, mq)
	}
	return out
}

func apiWSAppsMessagingQueuesListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	queues := filterMessageQueues(messageQueueStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiMessageQueue, len(queues))
	for i, mq := range queues {
		out[i] = toAPIMessageQueue(mq)
	}
	apiOK(w, apiMessageQueueList{Queues: out, Total: len(out)})
}

type apiCreateMessageQueueRequest struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Type      string `json:"type"`
	Retention string `json:"retention"`
}

func apiWSAppsMessagingQueuesCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateMessageQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "Invalid JSON body.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apiError(w, http.StatusBadRequest, "Name is required.")
		return
	}
	engine := strings.TrimSpace(req.Engine)
	if engine == "" {
		engine = "RabbitMQ"
	}
	typ := strings.TrimSpace(req.Type)
	if typ == "" {
		typ = "Work queue"
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "7 days"
	}
	queue := MessageQueue{
		ID: newID(), Name: name, Engine: engine,
		Type: typ, Messages: "0 pending", Throughput: "0 msg/s",
		Consumers: 0, Retention: retention,
		Status: "active", Updated: "2026-04-17",
	}
	messageQueueStore.add(slug, queue)
	apiCreated(w, toAPIMessageQueue(queue))
}

func apiWSAppsMessagingQueuesGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	queue, found := messageQueueStore.getByID(slug, r.PathValue("qid"))
	if !found {
		apiError(w, http.StatusNotFound, "Message queue not found.")
		return
	}
	apiOK(w, toAPIMessageQueue(queue))
}

func apiWSAppsMessagingQueuesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := messageQueueStore.getByID(slug, r.PathValue("qid")); !found {
		apiError(w, http.StatusNotFound, "Message queue not found.")
		return
	}
	messageQueueStore.delete(slug, r.PathValue("qid"))
	apiOK(w, map[string]any{"ok": true, "message": "Message queue deleted."})
}
