package events

import events "github.com/gkorsten/TodoCQRS/internal/eventbus"

const (
	TODO_DB_UPDATED   events.Event = "todoDbUpdated"     //Tell the projection service to update the views cache
	TODO_VIEW_UPDATED events.Event = "todoViewUpdated" //Tell the SSE service to broadcast the updated views
)