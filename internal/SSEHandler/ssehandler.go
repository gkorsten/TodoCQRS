package ssehandler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gkorsten/TodoCQRS/internal/eventbus"
	"github.com/gkorsten/TodoCQRS/internal/events"
	"github.com/gkorsten/TodoCQRS/internal/kvstore"
	"github.com/gkorsten/TodoCQRS/internal/projections"

	"github.com/starfederation/datastar-go/datastar"
)

type ClientChannel chan eventbus.Event

type sseController struct {
	clients     map[ClientChannel]bool //Hold a map of client connections
	newClient   chan ClientChannel     // when a new client connects, it needs to be added to clients
	closeClient chan ClientChannel     //disconected cient channel - to remove client from map
	broadcast   chan eventbus.Event    // the broadcast channel that all clients must listen to for events.

	pres     *projections.Projection
	updating bool
	connid   int
	mutex    sync.Mutex
}

// This returns an SSE Handler that uses a fan our approach to broadcast
// changes to longstanding connected SSE Clients
func NewSSEController(st kvstore.Store, pres *projections.Projection) *sseController {
	sse := &sseController{
		clients:     make(map[ClientChannel]bool),
		newClient:   make(chan ClientChannel),
		closeClient: make(chan ClientChannel),
		broadcast:   make(chan eventbus.Event),
		pres:        pres,
		updating:    false,
		connid:      1,
	}
	go sse.run()
	return sse
}

func (s *sseController) run() {
	slog.Info("SsseController: SSE is running")
	for {
		select {
		case nc := <-s.newClient:
			s.clients[nc] = true

			select {
			case nc <- events.VIEW_TODO_UPDATED:
				slog.Info("SsseController: Send the first view to the client")
			default:
				slog.Info("SsseController: ❌ Channel blocked, cant send first view - increase buffer size")
			}

		case nc := <-s.closeClient:
			delete(s.clients, nc)
			close(nc)

		case event := <-s.broadcast:
			for clientChan := range s.clients {
				select {
				case clientChan <- event:
					//success in sending event
				default:
					slog.Info("SsseController: ❌ Client blocked, closing connection (slow of disconnected client)")
					go func(c ClientChannel) {
						s.closeClient <- c
					}(clientChan)
				}
			}
		}
	}
}

// Listen to Events from the eventbus, and send them for broadcasting.
func (s *sseController) Listen(ev eventbus.Event) []byte {
	s.broadcast <- ev
	return []byte{}
}

// Handler to manage long standing SSE connections
func (s *sseController) SSEHandler(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	clientID := s.connid
	s.connid++
	s.mutex.Unlock()

	slog.Info("SSEHandler: ✔️ SSE Connection Opened", "NewID", clientID)

	sse := datastar.NewSSE(w, r)

	clientChan := make(ClientChannel, 1)

	s.newClient <- clientChan

	defer func() {
		s.closeClient <- clientChan
		slog.Info("SSEHandler: Client Disconnected", "Client Id", clientID)
	}()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			slog.Info("SSEHandler: ServeHTTP - Connection Closed")
			return
		case event := <-clientChan:
			slog.Info("SSEHandler: SSE event received", "clientid", clientID, "event", event)

			switch event {
			case events.VIEW_TODO_UPDATED:
				value := s.pres.FetchTodoView(clientID)
				sse.PatchElements(string(value))
				slog.Info("SSEHandler: Todo View sent to client", "clientid", clientID)

			default:
				sse.ExecuteScript("alert('Invalid event received)'")
				slog.Info("SSEHandler: ❌ SSE Event was invalid", "Event", event)
			}
		}
	}
}
