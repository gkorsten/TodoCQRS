package DB

import (
	"log/slog"
	"sync"

	"github.com/gkorsten/TodoCQRS/internal/eventbus"
	e "github.com/gkorsten/TodoCQRS/internal/events"
	"github.com/gkorsten/TodoCQRS/pages"
)

//This handles all the todo.mockDB Transactions
type mockDB struct {
	TODODB map[string][]pages.TodoItem
	mut sync.RWMutex
}

func newTodoDB () mockDB {
	 return mockDB {
		TODODB:make(map[string][]pages.TodoItem),
	 }
}

type TodoDBService struct {
	eventBus eventbus.EventBus
	id int
	mutex sync.RWMutex
	mockDB mockDB
}

func NewTodoDBService(eb eventbus.EventBus) *TodoDBService {
	return &TodoDBService{
		eventBus: eb,
		id:       0,
		mockDB: newTodoDB(),
	}
}

//Add a new todo to the DB
func (todo *TodoDBService) AddTodo (td pages.TodoItem) {
	todo.mutex.Lock()
	todo.id++
	td.Id=todo.id
	todo.mutex.Unlock()

	todo.mockDB.mut.Lock()
	todo.mockDB.TODODB["TODO"] = append(todo.mockDB.TODODB["TODO"],td)
	todo.mockDB.mut.Unlock()
	slog.Info("todoDBService: TODO DB Updated - Added a todo")
	todo.eventBus.Publish(e.TODO_DB_UPDATED)
}

//Get all Todos from the DB
func (todo *TodoDBService) GetTodos () []pages.TodoItem{
	todo.mutex.RLock()
	defer  todo.mutex.RUnlock()

	val,ok:= todo.mockDB.TODODB["TODO"]
	if !ok {
		val=nil
	}
	return val
}

//Toggle status in the db for a task
func (todo *TodoDBService) ToggleStatus(id int) {
	slog.Info("todoDBService: Toggle checkbox in db","id",id)
	todo.mockDB.mut.Lock()
	defer todo.mockDB.mut.Unlock()

	for idx,v:=range todo.mockDB.TODODB["TODO"] {
		if v.Id==id {
			todo.mockDB.TODODB["TODO"][idx].Done=!todo.mockDB.TODODB["TODO"][idx].Done
			break
		}
	}
	
	todo.eventBus.Publish(e.TODO_DB_UPDATED)
}

//Delete item
func (todo *TodoDBService) Delete(id int) {
	slog.Info("todoDBService: Remove item from DB","id",id)
	todo.mockDB.mut.Lock()
	defer todo.mockDB.mut.Unlock()

	indexToDel :=-1
	for idx,v:=range todo.mockDB.TODODB["TODO"] {
		if v.Id==id {
			indexToDel=idx
			break
		}
	}
	if indexToDel>-1 {
		todo.mockDB.TODODB["TODO"] = append(todo.mockDB.TODODB["TODO"][:indexToDel],todo.mockDB.TODODB["TODO"][indexToDel+1:]...)
	}
	
	todo.eventBus.Publish(e.TODO_DB_UPDATED)
}

//Delete a range of items based on ids []int
func (todo *TodoDBService) DeleteRange(ids []int) {
  slog.Info("todoDBService: Remove Range from DB","ids",ids)
  todo.mockDB.mut.Lock()
  defer todo.mockDB.mut.Unlock()
	for _,id := range ids {
		indexToDel :=-1
		for idx,v:=range todo.mockDB.TODODB["TODO"] {
			if v.Id==id {
				indexToDel=idx
				break
			}
		}
		if indexToDel>-1 {
			todo.mockDB.TODODB["TODO"] = append(todo.mockDB.TODODB["TODO"][:indexToDel],todo.mockDB.TODODB["TODO"][indexToDel+1:]...)
		}
	}
	
	todo.eventBus.Publish(e.TODO_DB_UPDATED)
}


