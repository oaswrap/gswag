package excludehidden

import (
	"net/http"

	"github.com/oaswrap/gswag/test/util"
)

type PublicResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type InternalResource struct {
	Host string `json:"host"`
	Load int    `json:"load"`
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /public", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, PublicResource{ID: 1, Name: "widget"})
	})

	mux.HandleFunc("GET /internal", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, InternalResource{Host: "node-1", Load: 42})
	})

	mux.HandleFunc("GET /secret", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, map[string]string{"key": "value"})
	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, map[string]int{"requests": 1000})
	})

	mux.HandleFunc("GET /admin/health", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return mux
}
