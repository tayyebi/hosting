package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// ════════════════════════════════════════════════════════════════
// Workspace Application Services API — Messaging: Event Streaming
// apps/messaging/streaming
// ════════════════════════════════════════════════════════════════

type EventStream struct {
	ID          string
	Name        string
	Engine      string
	Partitions  int
	Replication int
	Throughput  string
	Consumers   string
	Retention   string
	Status      string
	Updated     string
}

type eventStreamStoreType struct {
	mu      sync.RWMutex
	streams map[string][]EventStream
}

var eventStreamStore = &eventStreamStoreType{
	streams: map[string][]EventStream{},
}

func (s *eventStreamStoreType) list(slug string) []EventStream {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.streams[slug]
	out := make([]EventStream, len(src))
	copy(out, src)
	return out
}

func (s *eventStreamStoreType) getByID(slug, id string) (EventStream, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, es := range s.streams[slug] {
		if es.ID == id {
			return es, true
		}
	}
	return EventStream{}, false
}

func (s *eventStreamStoreType) add(slug string, es EventStream) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams[slug] = append(s.streams[slug], es)
}

func (s *eventStreamStoreType) delete(slug, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.streams[slug]
	for i, es := range list {
		if es.ID == id {
			s.streams[slug] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

type apiEventStream struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Engine      string `json:"engine"`
	Partitions  int    `json:"partitions"`
	Replication int    `json:"replication"`
	Throughput  string `json:"throughput"`
	Consumers   string `json:"consumers"`
	Retention   string `json:"retention"`
	Status      string `json:"status"`
	Updated     string `json:"updated"`
}

type apiEventStreamList struct {
	Streams []apiEventStream `json:"streams"`
	Total   int              `json:"total"`
}

func toAPIEventStream(es EventStream) apiEventStream {
	return apiEventStream{
		ID: es.ID, Name: es.Name, Engine: es.Engine,
		Partitions: es.Partitions, Replication: es.Replication,
		Throughput: es.Throughput, Consumers: es.Consumers,
		Retention: es.Retention, Status: es.Status, Updated: es.Updated,
	}
}

func filterEventStreams(all []EventStream, search, status string) []EventStream {
	var out []EventStream
	q := strings.ToLower(search)
	for _, es := range all {
		if status != "" && es.Status != status {
			continue
		}
		if q != "" {
			hay := strings.ToLower(es.Name + " " + es.Engine + " " + es.Consumers)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, es)
	}
	return out
}

func apiWSAppsMessagingStreamingListHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	streams := filterEventStreams(eventStreamStore.list(slug), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	out := make([]apiEventStream, len(streams))
	for i, es := range streams {
		out[i] = toAPIEventStream(es)
	}
	apiOK(w, apiEventStreamList{Streams: out, Total: len(out)})
}

type apiCreateEventStreamRequest struct {
	Name        string `json:"name"`
	Engine      string `json:"engine"`
	Partitions  int    `json:"partitions"`
	Replication int    `json:"replication"`
	Retention   string `json:"retention"`
}

func apiWSAppsMessagingStreamingCreateHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	var req apiCreateEventStreamRequest
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
		engine = "Apache Kafka"
	}
	partitions := req.Partitions
	if partitions <= 0 {
		partitions = 8
	}
	replication := req.Replication
	if replication <= 0 {
		replication = 3
	}
	retention := strings.TrimSpace(req.Retention)
	if retention == "" {
		retention = "7 days"
	}
	stream := EventStream{
		ID: newID(), Name: name, Engine: engine,
		Partitions: partitions, Replication: replication,
		Throughput: "0 events/s", Consumers: "",
		Retention: retention, Status: "active", Updated: "2026-04-17",
	}
	eventStreamStore.add(slug, stream)
	apiCreated(w, toAPIEventStream(stream))
}

func apiWSAppsMessagingStreamingGetHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	stream, found := eventStreamStore.getByID(slug, r.PathValue("sid"))
	if !found {
		apiError(w, http.StatusNotFound, "Event stream not found.")
		return
	}
	apiOK(w, toAPIEventStream(stream))
}

func apiWSAppsMessagingStreamingDeleteHandler(w http.ResponseWriter, r *http.Request) {
	slug, _, ok := wsAuth(w, r)
	if !ok {
		return
	}
	if _, found := eventStreamStore.getByID(slug, r.PathValue("sid")); !found {
		apiError(w, http.StatusNotFound, "Event stream not found.")
		return
	}
	eventStreamStore.delete(slug, r.PathValue("sid"))
	apiOK(w, map[string]any{"ok": true, "message": "Event stream deleted."})
}
