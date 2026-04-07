package allmethods

import (
	"encoding/json"
	"net/http"

	"github.com/oaswrap/gswag/test/util"
)

type AllMethodsModel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewRouter() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /allmethods", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, map[string]string{"method": "GET"})
	})

	r.HandleFunc("POST /allmethods", func(w http.ResponseWriter, r *http.Request) {
		var in AllMethodsModel
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		util.WriteJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("PUT /allmethods", func(w http.ResponseWriter, r *http.Request) {
		var in AllMethodsModel
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		util.WriteJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("PATCH /allmethods", func(w http.ResponseWriter, r *http.Request) {
		var in AllMethodsModel
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		util.WriteJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("DELETE /allmethods", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
