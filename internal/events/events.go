package events

import events "github.com/gkorsten/TodoCQRS/internal/eventbus"

const (
	TODO_UPDATED       events.Event = "todoUpdated"     //Tell the projection service to update the views cache
	VIEW_TODO_UPDATED events.Event = "viewTODOUpdated" //Tell the SSE service to broadcast the updated views
)