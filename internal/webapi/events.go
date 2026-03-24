package webapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// EventBroker manages Server-Sent Events connections.
type EventBroker struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func NewEventBroker() *EventBroker {
	return &EventBroker{
		clients: make(map[chan []byte]struct{}),
	}
}

// Emit sends an event to all connected SSE clients.
func (b *EventBroker) Emit(event string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("SSE marshal error: %v", err)
		return
	}

	msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, jsonData))

	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// Client too slow, skip
		}
	}
}

// ServeHTTP handles SSE connections at /api/events.
func (b *EventBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", 500)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan []byte, 64)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
		close(ch)
	}()

	// Send initial ping
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			w.Write(msg)
			flusher.Flush()
		}
	}
}
