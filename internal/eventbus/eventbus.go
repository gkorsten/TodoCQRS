package eventbus

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

//Event is a function handler that receives a value

type Event string 

type EventHanlder func(e Event)string

type registerHandler struct {
	Id string
	EventHandler EventHanlder	

}

type EventBus interface {
	Publish(event Event) error
}

type eventBus struct {
	clients   map[Event][]registerHandler
	broadcast   chan Event
	stop 	chan struct{}
	mutex sync.RWMutex
	isClosing atomic.Bool
	handlerWg sync.WaitGroup
}

//Create a New eventBus, returning the eventBus.
func NewEventBus() *eventBus {
	eb:=eventBus{
		clients:make(map[Event][]registerHandler),
		broadcast: make(chan Event,5),
		stop: make (chan struct{}),
	}
	go eb.run()

	return &eb
}

//Monitor the event chanel and broadcast event to all subscribers
func (e *eventBus) run () {
	slog.Info("Eventbus: Event bus running")
	for {
	select {
		case event,ok := <-e.broadcast:
			slog.Info("Eventbus (run): Event received","Event",event)
			if !ok {
				slog.Info("Eventbus (run): ❌ Broadcast Channel closed ")
				return
			}
			slog.Info("Eventbus (run): Broadcasting event to clients","Event",event)
			e.mutex.RLock()
			for _,client := range e.clients[event] {
				e.handlerWg.Add(1)
				go func(c registerHandler,ev Event) {
					defer e.handlerWg.Done()
					c.EventHandler(ev)
				}(client,event)
			}
			e.mutex.RUnlock()
		
		case <-e.stop :
			slog.Info("Eventbus (run): Stop the eventBus ")
			return
	}
	}
}

//Register a event with a handler, receive back the id of the handler regitration
//which is needed to deregister the handler from the event.
func (e *eventBus) RegisterEvent(event Event,handler EventHanlder ) string{	
	e.mutex.Lock()
	id := string(fmt.Sprintf("%d", time.Now().UnixNano()))
	regHandler := registerHandler{
		Id:id,
		EventHandler: handler,
	}

	e.clients[event] = append(e.clients[event], regHandler)
	e.mutex.Unlock()
	slog.Info("EventBus: New subscription registered","id",id,"Subs",len(e.clients),"Event",event)
	return id
}

//Deregisters the event using the id
func (e *eventBus) Unregister(event Event,id string) error{
	
	e.mutex.Lock()
    defer e.mutex.Unlock()
	handlers, ok := e.clients[event]
    if !ok {
        slog.Info("EventBus: Cannot unregister, event has no subscribers", "Event", event)
        return errors.New("no such event")
    }

    found := false
    newHandlers := make([]registerHandler, 0, len(handlers))

    for _, h := range handlers {
        if h.Id != id {
            newHandlers = append(newHandlers, h)
        } else {
            found = true
        }
    }

    if found {
        e.clients[event] = newHandlers
        
        if len(newHandlers) == 0 {
            delete(e.clients, event)
            slog.Info("EventBus: Event key removed as all handlers were unregistered", "Event", event)
        }
        
        slog.Info("EventBus: Unregistered handler successfully", 
            "Event", event, 
            "RemainingSubscriptions", len(newHandlers), 
            "TotalEventsTracked", len(e.clients))
		return nil
        
    } else {
        slog.Warn("eventBus: Handler not found for unregistration", "Event", event)
        return errors.New("no such handler or id found")
    }

}

//Send events to all the subscribers
func (e *eventBus) Publish(event Event) error{
	if e.isClosing.Load() {
		slog.Info("Eventbus: ❌ event bus is down, cannot publis event")
        return errors.New("event bus is down, cannot publish event")
    }

	e.mutex.RLock()
	if _,ok:=e.clients[event] ;!ok {
		slog.Info("Eventbus: ❌ Error - no such event","Event",event)
		return errors.New("no such event or subscriptions")
	}
	e.mutex.RUnlock()

	slog.Info("Eventbus: Publish event","Event",event)
	e.broadcast <- event
	return nil

} 

func (e *eventBus) wait() {
    e.handlerWg.Wait()
}

//Shutdown the event bus
func (e *eventBus) Stop () {
	slog.Info("Eventbus: Stopping the eventbus")
	e.isClosing.Store(true)
	close(e.stop)
	close(e.broadcast)
	e.handlerWg.Wait()
}