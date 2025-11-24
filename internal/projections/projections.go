package projections

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/gkorsten/TodoCQRS/internal/DB"
	"github.com/gkorsten/TodoCQRS/internal/eventbus"
	"github.com/gkorsten/TodoCQRS/internal/events"
	"github.com/gkorsten/TodoCQRS/internal/kvstore"
	"github.com/gkorsten/TodoCQRS/pages"
)

type Projection struct {
	eb    eventbus.EventBus
	store kvstore.Store
	db    *DB.TodoDBService
}

func NewProjection(ebv eventbus.EventBus, storev kvstore.Store, db *DB.TodoDBService) *Projection {
	return &Projection{
		eb:    ebv,
		store: storev,
		db:    db,
	}
}

func (p *Projection) UpdateTodoProjection(e eventbus.Event) []byte {
	slog.Info("Projection: Render TODO and store in cache")

	todoitems := p.db.GetTodos()

	buf := new(bytes.Buffer)
	//Render the TMPL Page and store it in a Cache
	err := pages.ShowTodo(todoitems).Render(context.Background(), buf)
	if err == nil {
		p.store.AddItem("TODOPAGE", buf.Bytes())
	} else {
		slog.Info("ERROR buf2string", "error", err.Error())
	}
	p.eb.Publish(events.TODO_VIEW_UPDATED)
	return buf.Bytes()
}

func (p *Projection) FetchTodoView(clientid int) []byte {
	slog.Info("Projection: Fetch TODO View from store")
	value, ok := p.store.Fetch("TODOPAGE")
	if !ok {
		value = p.UpdateTodoProjection(events.TODO_DB_UPDATED)
		slog.Info("Projection: Cache MISS", "Clientid", clientid)
	} else {
		slog.Info("Projection: Cache HIT", "Clientid", clientid)
	}
	return value
}