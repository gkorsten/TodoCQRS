package todohandler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gkorsten/TodoCQRS/internal/DB"
	"github.com/gkorsten/TodoCQRS/pages"
)

type todoHandlers struct {
	tdb *DB.TodoDBService
}

func NewTodoHandlers(t *DB.TodoDBService) *todoHandlers {
	return &todoHandlers{
		tdb: t,
	}
}

func (t *todoHandlers) AddTodo(w http.ResponseWriter, r *http.Request) {
	slog.Info("todoHandler: Add TODO received from client")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	task := r.FormValue("newTodo")
	if task == "" {
		http.Error(w, "Empty todo received", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	newTodo := pages.TodoItem{
		Done: false,
		Task: task,
	}

	t.tdb.AddTodo(newTodo)
}

func (t *todoHandlers) RemoveCompleted(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)

	slog.Info("HUSE", "form", r.FormValue("newTodo"))
	var completed []int
	for key := range r.Form {
		if strings.HasPrefix(key, "CB_") {
			t := strings.Split(key, "_")
			val, err := strconv.Atoi(t[1])
			if err == nil {
				completed = append(completed, val)
			}
		}
	}
	t.tdb.DeleteRange(completed)
}

func (t *todoHandlers) ToggleCheckbox(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Failed to parse form: ", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	x, _ := strconv.Atoi(id)
	t.tdb.ToggleStatus(x)
}