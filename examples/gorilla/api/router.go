// Package api is a minimal Gorilla Mux HTTP server used by the gswag Gorilla example tests.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Task is the domain model for the tasks resource.
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// CreateTaskRequest is the request body for creating a task.
type CreateTaskRequest struct {
	Title string `json:"title"`
}

var tasks = []Task{
	{ID: 1, Title: "Buy groceries", Done: false},
	{ID: 2, Title: "Write tests", Done: true},
}

// NewRouter returns a configured Gorilla Mux router.
func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, tasks)
	}).Methods(http.MethodGet)

	r.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		var req CreateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		task := Task{ID: len(tasks) + 1, Title: req.Title, Done: false}
		tasks = append(tasks, task)
		writeJSON(w, http.StatusCreated, task)
	}).Methods(http.MethodPost)

	r.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}
		for _, t := range tasks {
			if t.ID == id {
				writeJSON(w, http.StatusOK, t)
				return
			}
		}
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	}).Methods(http.MethodGet)

	r.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodDelete)

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
