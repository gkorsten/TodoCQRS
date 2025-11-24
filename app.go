package main

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"

	"github.com/gkorsten/TodoCQRS/internal/DB"
	ssehandler "github.com/gkorsten/TodoCQRS/internal/SSEHandler"
	"github.com/gkorsten/TodoCQRS/internal/eventbus"
	"github.com/gkorsten/TodoCQRS/internal/events"
	"github.com/gkorsten/TodoCQRS/internal/kvstore"
	"github.com/gkorsten/TodoCQRS/internal/projections"
	todohandler "github.com/gkorsten/TodoCQRS/internal/todoHandler"
	"github.com/gkorsten/TodoCQRS/pages"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Error: 404 Page not Found : handler for %s not found", html.EscapeString(r.URL.Path))
			return
		}
	pages.MainPage().Render(r.Context(),w)
}

func main() {
	slog.Info("Main: TODO CQRS implementation")
   
	
	// handles the events - passes it on from Control -> Projection side -> SSE Views updates 
	eventBus := eventbus.NewEventBus()
	defer eventBus.Stop()

	//Basic Key,Value Store - used by Projection side to create templ views and cache them
	kvStore:=kvstore.NewKVStore()
	todoDBService:= DB.NewTodoDBService(eventBus)

	//Query side os CQRS
	Projection := projections.NewProjection(eventBus,kvStore,todoDBService)
	sse := ssehandler.NewSSEController(kvStore,Projection) //Handles the 

	  //DB Writer
	todoHandlers:=todohandler.NewTodoHandlers(todoDBService) //Handlers for todo - can be part of DB Serivce ?

	//Register events on Eventbus - probably a better idea is to have two eventbusses for each side
	eventBus.RegisterEvent(events.TODO_DB_UPDATED,Projection.UpdateTodoProjection) //Tell the Projection layer to update the TODO view
	eventBus.RegisterEvent(events.TODO_VIEW_UPDATED,sse.Listen) //Tell the SSE handler to send the new view to all clients 

    //Route Handlers
	http.HandleFunc("POST /addTodo",todoHandlers.AddTodo)
	http.HandleFunc("POST /removeTodoCompleted",todoHandlers.RemoveCompleted)
	http.HandleFunc("GET /toggleCheckbox",todoHandlers.ToggleCheckbox)

	http.HandleFunc("GET /sse",sse.SSEHandler)
	http.HandleFunc("/",mainPage)

	slog.Info("Main: Server running on port 3010")
	err:=http.ListenAndServe(":3010",nil)
	if err!=nil {
		panic(err)
	}
	
}