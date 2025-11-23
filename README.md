## Collaborative TODO App 

Stack:
* Golang   : https://go.dev/
* Templ    : https://templ.guide/
* Datastar : https://data-star.dev/

Using the CQRS design pattern with a longstanding SSE connection (managed by Datastar).

### Why?

I wanted to learn about the CQRS pattern. By using Datastar and longstanding SSE connections that pushes 
updates to all the clients that are connected to the same Todo app. 

This then allows mutliple people to connect to the same server, and all see the same view - that gets live updated as people add and 
mark completed items. 

I tried to keep is modular so that any part can be swopped out as the project scales. (eg the Eventbus for Nats, the KVStore for Redis)

### Building

Install make on your system.

* make build - Creates an bin\todo.exe
* make dev -  Launches a Templ watch environment

Run the todo.exe and connect to http://localhost:3010 

Can open multiple browsers and connect to http://localhost:3010/ to see it live updating.